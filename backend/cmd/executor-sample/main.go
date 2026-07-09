// 示例执行器：注册 demo/slow handler，连接 admin，验证端到端调度与实时日志。
//
// 运行：go run ./cmd/executor-sample -addr http://localhost:9090 -app demo-app
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"xxl-job-admin-go/internal/executor"
)

func main() {
	addr := flag.String("addr", "http://localhost:9090", "admin gRPC 地址")
	app := flag.String("app", "demo-app", "执行器 app 名称")
	token := flag.String("token", "", "调度中心 access_token（与 admin 配置一致）")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	ex := executor.New(*addr, *app, *token, logger)

	// demoJob：快速任务，输出实时日志
	ex.Register("demoJob", func(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
		jc.Log("demoJob 开始执行, param=" + param)
		time.Sleep(100 * time.Millisecond)
		jc.Log("demoJob 处理中...")
		time.Sleep(100 * time.Millisecond)
		jc.Log("demoJob 执行完成")
		return 200, "success: " + param
	})

	// slowJob：慢任务（验证快慢线程池降级）
	ex.Register("slowJob", func(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
		jc.Log("slowJob 开始, 预计 2s")
		time.Sleep(2 * time.Second)
		jc.Log("slowJob 完成")
		return 200, "done"
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	logger.Info("executor sample starting", zap.String("app", *app), zap.String("addr", *addr))
	if err := ex.Run(ctx); err != nil && err != context.Canceled {
		logger.Fatal("executor stopped", zap.Error(err))
	}
}
