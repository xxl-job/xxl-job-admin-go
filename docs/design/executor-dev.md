# 执行器开发指南

执行器是任务的实际执行者，通过 gRPC 双向流（`ExecutorLink`）接入调度中心。本仓库提供 Go 执行器 SDK（`internal/executor`），其他语言可按 proto 自行实现。

## 快速开始

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"xxl-job-admin-go/internal/executor"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ex := executor.New("http://localhost:9090", "my-app", "", logger)

	// 注册任务 handler
	ex.Register("sayHello", func(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
		jc.Log("收到参数: " + param)
		// 业务逻辑...
		jc.Log("执行完成")
		return 200, "hello " + param
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	if err := ex.Run(ctx); err != nil && err != context.Canceled {
		logger.Fatal("executor stopped", zap.Error(err))
	}
}
```

## SDK 核心

### 创建执行器

```go
ex := executor.New(serverAddr, appName, accessToken, logger)
```

- `serverAddr`：admin 的 gRPC 地址（如 `http://localhost:9090`）
- `appName`：执行器分组标识，需与 admin 中创建的执行器分组 `app_name` 一致

### 注册 Handler

```go
type HandlerFunc func(ctx context.Context, jc *JobContext, param string) (code int, msg string)
```

- `jc.Log(msg)`：回传实时日志到 admin（前端 WebSocket 可查看）
- 返回 `code`：200 成功，其他失败（触发告警 + metrics）
- `param`：来自任务的 `executor_param`

### 运行

```go
ex.Run(ctx)
```

阻塞运行，自动：
1. 连接 admin（gRPC h2c）
2. 注册（上报 app_name + handler 列表）
3. 心跳保活（15s）
4. 接收 `TriggerTask` 并发执行
5. 回传 `TaskResult` + `LogChunk`
6. 断线自动重连（3s）

## 内置能力

| 能力 | 说明 |
|---|---|
| 心跳 | 复用长连接，15s 一次 |
| 断线重连 | 3s 重试 |
| 超时控制 | `TriggerTask.executor_timeout` 秒，超时回传 500 |
| 日志回传 | `jc.Log` → `LogChunk` → admin Hub → 前端 WebSocket |
| 负载上报 | 可选，`LoadReport`（P6 扩展点） |
| 任务取消 | `CancelTask`（执行器收到后调 `cancelRun` 取消 per-task context；调度侧 cover-early 阻塞策略触发下发） |

## 阻塞策略（执行器侧）

admin 的阻塞策略（serial/discard-later/cover-early）在调度侧决策。串行（serial）语义下，执行器应保证同 job 串行执行——SDK 当前并发执行，如需严格串行可在 handler 内加锁或用队列。

## 多语言执行器

非 Go 语言可按 `proto/xxl/v1/executor.proto` 实现 `ExecutorLink.Link` 双向流：

1. 建立 gRPC 双向流到 `http://admin:9090/xxl.v1.ExecutorLink/Link`
2. 发送 `ExecutorMessage.Register`（app_name + handlers）
3. 循环接收 `AdminMessage`：
   - `Trigger` → 执行任务 → 发送 `TaskResult`
   - `Cancel` → 取消任务
4. 周期发送 `Heartbeat`（复用流）
5. 失败时发送 `LoadReport`（可选）

## 示例

`cmd/executor-sample` 提供 `demoJob`（快速）与 `slowJob`（2s，验证快慢线程池降级）两个 handler，可直接运行验证：

```bash
go run ./cmd/executor-sample -addr http://localhost:9090 -app demo-app
```
