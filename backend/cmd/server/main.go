// Package main xxl-job-admin-go 启动入口。
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c" //nolint:staticcheck // h2c 明文 HTTP/2（gRPC）；Go 1.25 http.Server.Protocols 迁移列后续

	"xxl-job-admin-go/internal/alarm"
	"xxl-job-admin-go/internal/api"
	"xxl-job-admin-go/internal/auth"
	"xxl-job-admin-go/internal/conf"
	"xxl-job-admin-go/internal/logstream"
	"xxl-job-admin-go/internal/pkg/lock"
	"xxl-job-admin-go/internal/registry"
	"xxl-job-admin-go/internal/rpc"
	"xxl-job-admin-go/internal/scheduler"
	"xxl-job-admin-go/internal/store"
	"xxl-job-admin-go/internal/trigger"
)

func main() {
	// 1. 配置
	cfgPath := "configs/config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		cfgPath = p
	}
	cfg, err := conf.Load(cfgPath)
	if err != nil {
		panic(err)
	}

	// 2. logger
	logger := newLogger(cfg.Log)
	defer func() { _ = logger.Sync() }()
	logger.Info("starting xxl-job-admin-go", zap.String("node_id", cfg.NodeID))

	// 3. DB
	st, err := store.New(cfg.DB)
	if err != nil {
		logger.Fatal("init store", zap.Error(err))
	}
	if err := st.AutoMigrate(); err != nil {
		logger.Fatal("auto migrate", zap.Error(err))
	}
	if err := st.Seed(); err != nil {
		logger.Fatal("seed", zap.Error(err))
	}
	logger.Info("db ready")

	// 4. registry + logstream
	reg := registry.NewMemory()
	hub := logstream.New()

	// 5. 告警 + 触发器 + 调度器（仅 leader 节点运行调度）
	alerter := alarm.NewMulti(logger, &alarm.LogNotifier{Log: logger}, &alarm.EmailNotifier{Log: logger})
	block := trigger.NewBlockTracker()
	trig := trigger.NewTrigger(st.DB(), reg, block, logger)
	pool := scheduler.NewPool(200, 100, 500*time.Millisecond)
	sched := scheduler.New(st.DB(), pool, trig, logger)

	// 6. 租约锁 + leader 选举
	locker := lock.NewMySQLLeaseLocker(st.DB(), cfg.NodeID, cfg.Lock.LeaseTTL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runLease(ctx, locker, cfg, logger, func(leadCtx context.Context) {
		go scheduler.RunPartitionMaintainer(leadCtx, st.DB(), logger) // 仅 leader 维护分区
		sched.Run(leadCtx)
	})

	// 7. RPC server（执行器连此端口）
	execLink := rpc.NewExecutorLinkService(reg, st.DB(), trig, alerter, hub, logger)
	adminSvc := rpc.NewAdminService(trig, st.DB())
	rpcServer := &http.Server{
		Addr:    cfg.Server.GRPCAddr,
		Handler: h2c.NewHandler(rpc.NewServer(execLink, adminSvc, cfg.AccessToken), &http2.Server{}), //nolint:staticcheck // h2c 明文 HTTP/2（gRPC）；Go 1.25 http.Server.Protocols 迁移列后续
	}

	// 8. API server（前端 REST + /metrics + WS 日志流）
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)
	apiRouter := api.NewRouter(st.DB(), jwtMgr, reg, trig, hub)
	apiServer := &http.Server{
		Addr:    cfg.Server.HTTPAddr,
		Handler: apiRouter.Engine(),
	}

	// 9. 启动两个 HTTP 服务
	go func() {
		logger.Info("rpc server listening", zap.String("addr", rpcServer.Addr))
		if err := rpcServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("rpc server", zap.Error(err))
		}
	}()
	go func() {
		logger.Info("api server listening", zap.String("addr", apiServer.Addr))
		if err := apiServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("api server", zap.Error(err))
		}
	}()

	// 10. 优雅停机
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Info("shutting down", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	cancel()
	_ = locker.Release(context.Background(), "schedule_leader")

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("api shutdown", zap.Error(err))
	}
	if err := rpcServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("rpc shutdown", zap.Error(err))
	}
	pool.Stop()
	if sqlDB, err := st.DB().DB(); err == nil {
		_ = sqlDB.Close()
	}
	logger.Info("stopped")
}

func newLogger(c conf.LogConfig) *zap.Logger {
	var zcfg zap.Config
	if c.Format == "json" {
		zcfg = zap.NewProductionConfig()
	} else {
		zcfg = zap.NewDevelopmentConfig()
	}
	switch c.Level {
	case "debug":
		zcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zcfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zcfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zcfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	l, _ := zcfg.Build()
	return l
}

// runLease leader 租约续约循环。
func runLease(ctx context.Context, locker lock.Locker, cfg *conf.Config, log *zap.Logger, onLead func(context.Context)) {
	const key = "schedule_leader"
	var leadCancel context.CancelFunc
	isLeader := false

	tryAcquire := func() {
		got, err := locker.TryAcquire(ctx, key)
		if err != nil {
			log.Error("lease acquire", zap.Error(err))
			return
		}
		if got && !isLeader {
			isLeader = true
			leadCtx, cancel := context.WithCancel(ctx)
			leadCancel = cancel
			log.Info("acquired leadership, scheduler started", zap.String("node", cfg.NodeID))
			go onLead(leadCtx)
		}
	}
	tryAcquire()

	ticker := time.NewTicker(cfg.Lock.RenewInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if leadCancel != nil {
				leadCancel()
			}
			return
		case <-ticker.C:
			ok, err := locker.Renew(ctx, key)
			if err != nil {
				log.Error("lease renew", zap.Error(err))
				continue
			}
			if !ok {
				if isLeader {
					isLeader = false
					if leadCancel != nil {
						leadCancel()
						leadCancel = nil
					}
					log.Warn("lost leadership, scheduler stopped", zap.String("node", cfg.NodeID))
				}
				tryAcquire()
			}
		}
	}
}
