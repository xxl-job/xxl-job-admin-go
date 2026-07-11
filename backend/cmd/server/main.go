// Package main xxl-job-admin-go 启动入口。
package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c" //nolint:staticcheck // h2c 明文 HTTP/2（gRPC）；Go 1.25 http.Server.Protocols 迁移列后续

	"xxl-job-admin-go/internal/alarm"
	"xxl-job-admin-go/internal/api"
	"xxl-job-admin-go/internal/auth"
	"xxl-job-admin-go/internal/circuit"
	"xxl-job-admin-go/internal/conf"
	"xxl-job-admin-go/internal/limiter"
	"xxl-job-admin-go/internal/logstream"
	"xxl-job-admin-go/internal/pkg/lock"
	"xxl-job-admin-go/internal/registry"
	"xxl-job-admin-go/internal/rpc"
	"xxl-job-admin-go/internal/scheduler"
	"xxl-job-admin-go/internal/store"
	"xxl-job-admin-go/internal/trigger"
)

func main() {
	demo := flag.Bool("demo", false, "demo 模式：seed 示例数据 + 内嵌执行器，开箱即用（生产勿用）")
	flag.Parse()

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
	// 告警通道：默认日志；按配置追加钉钉/企微/飞书 webhook + 邮件 SMTP（留空则不启用）
	hc := &http.Client{Timeout: 5 * time.Second}
	notifiers := []alarm.Notifier{&alarm.LogNotifier{Log: logger}}
	if cfg.Alarm.Dingtalk != "" {
		notifiers = append(notifiers, &alarm.WebhookNotifier{Platform: "dingtalk", URL: cfg.Alarm.Dingtalk, Client: hc, Log: logger})
	}
	if cfg.Alarm.Wechat != "" {
		notifiers = append(notifiers, &alarm.WebhookNotifier{Platform: "wechat", URL: cfg.Alarm.Wechat, Client: hc, Log: logger})
	}
	if cfg.Alarm.Feishu != "" {
		notifiers = append(notifiers, &alarm.WebhookNotifier{Platform: "feishu", URL: cfg.Alarm.Feishu, Client: hc, Log: logger})
	}
	if cfg.Alarm.Email.Host != "" {
		notifiers = append(notifiers, &alarm.EmailNotifier{
			Host: cfg.Alarm.Email.Host, Port: cfg.Alarm.Email.Port,
			User: cfg.Alarm.Email.User, Password: cfg.Alarm.Email.Password,
			From: cfg.Alarm.Email.From, To: strings.Split(cfg.Alarm.Email.To, ","), Log: logger,
		})
	}
	alerter := alarm.NewMulti(logger, notifiers...)
	block := trigger.NewBlockTracker()
	trig := trigger.NewTrigger(st.DB(), reg, block, logger)
	breaker := circuit.New(5, 30*time.Second) // 连续 5 次失败熔断 30s，冷却后半开探测
	trig.SetBreaker(breaker)
	rl := limiter.New(100, 100) // 每 app 100 rps + burst 100，保护执行器不被过载
	trig.SetLimiter(rl)
	pool := scheduler.NewPool(200, 100, 500*time.Millisecond)
	sched := scheduler.New(st.DB(), pool, trig, logger)

	// 6. 租约锁 + leader 选举
	locker := lock.NewMySQLLeaseLocker(st.DB(), cfg.NodeID, cfg.Lock.LeaseTTL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runLease(ctx, locker, cfg, logger, func(leadCtx context.Context) {
		go scheduler.RunPartitionMaintainer(leadCtx, st.DB(), logger) // 仅 leader 维护分区
		go trigger.RunDedupCleaner(leadCtx, st.DB(), logger)          // 仅 leader 清理触发去重
		sched.Run(leadCtx)
	})

	// demo 模式：seed 示例数据 + 内嵌执行器连本机 gRPC（开箱即用，生产勿用）
	if *demo {
		if err := demoSeed(st.DB()); err != nil {
			logger.Warn("demo seed", zap.Error(err))
		}
		go runDemoExecutor(ctx, logger)
	}

	// 7. RPC server（执行器连此端口）
	execLink := rpc.NewExecutorLinkService(reg, st.DB(), trig, alerter, hub, logger)
	execLink.SetBreaker(breaker)
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
