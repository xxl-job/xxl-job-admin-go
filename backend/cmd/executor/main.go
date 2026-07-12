// executor 通用执行器二进制：连接 admin，注册内置通用 handler，启用脚本任务。
// 镜像入口：ADMIN_ADDR / APP_NAME / ACCESS_TOKEN 环境变量，或 -addr/-app/-token flag。
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"xxl-job-admin-go/cmd/executor/handlers"
	"xxl-job-admin-go/pkg/executor"
)

func main() {
	addr := flag.String("addr", os.Getenv("ADMIN_ADDR"), "admin gRPC 地址（如 http://admin:9090）")
	app := flag.String("app", os.Getenv("APP_NAME"), "执行器 app 名称")
	token := flag.String("token", os.Getenv("ACCESS_TOKEN"), "调度中心 access_token（可空）")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	ex := executor.New(*addr, *app, *token, logger)
	ex.WithScriptTask(true) // 镜像含 sh+python3，启用 Shell/Python GLUE 脚本

	handlers.Register(ex) // 注册内置通用 handler：httpCall/shellExec/mysqlQuery/sendWebhook

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()
	if err := ex.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("executor run", zap.Error(err))
	}
}
