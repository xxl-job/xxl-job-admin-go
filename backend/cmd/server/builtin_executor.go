package main

import (
	"context"
	"errors"
	"net"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/cmd/executor/handlers"
	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/pkg/executor"
)

// builtinSeed 幂等创建通用执行器分组，使 admin 任务表单可选中、内嵌执行器有分组对应。
// 与 demoSeed 区别：不 seed 示例任务（避免依赖 demoJob/slowJob 这些通用执行器未注册的 handler）。
func builtinSeed(db *gorm.DB, appName string) error {
	var exec model.Executor
	err := db.Where("app_name = ?", appName).First(&exec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		exec = model.Executor{AppName: appName, Title: "内置通用执行器", AddressType: 0}
		return db.Create(&exec).Error
	}
	return err
}

// runBuiltinExecutor 内嵌通用执行器：连本机 admin gRPC，注册 4 个内置 handler
// （httpCall/shellExec/mysqlQuery/sendWebhook），开启脚本与 HTTP 任务能力。
// 用于 all-in-one 体验模式（admin -executor），阻塞直到 ctx 取消；断线自动重连。
func runBuiltinExecutor(ctx context.Context, logger *zap.Logger, app, grpcAddr, accessToken string) {
	// grpcAddr 形如 ":9090" 或 "0.0.0.0:9090"，转成本机连接 URL。
	_, port, err := net.SplitHostPort(grpcAddr)
	if err != nil || port == "" {
		port = "9090"
	}
	serverURL := "http://127.0.0.1:" + port

	ex := executor.New(serverURL, app, accessToken, logger)
	ex.WithScriptTask(true) // 启用 Shell/Python GLUE 脚本
	ex.WithHTTPTask(true)   // 启用 glue_type=4 HTTP 任务
	handlers.Register(ex)   // 注册 httpCall/shellExec/mysqlQuery/sendWebhook

	logger.Info("builtin executor starting", zap.String("app", app), zap.String("admin", serverURL))
	if err := ex.Run(ctx); err != nil && ctx.Err() == nil {
		logger.Error("builtin executor stopped", zap.Error(err))
	}
}
