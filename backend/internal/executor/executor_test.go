package executor

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	v1 "xxl-job-admin-go/internal/gen/xxl/v1"
)

// noopSender 测试用：丢弃所有回传消息，不阻塞。
func noopSender(_ *v1.ExecutorMessage) {}

// TestRunWithTimeout 验证超时控制（ExecutorTimeout 单位为秒）：
// 超时返回 500/timeout；正常完成返回 handler 结果；
// handler 不响应 ctx 时仍只返回 timeout（不产生第二结果）。
func TestRunWithTimeout(t *testing.T) {
	e := New("http://localhost:9090", "test-app", "", zap.NewNop())
	jc := &JobContext{LogID: 1, sender: noopSender}

	run := func(timeout int32, handler string) (int, string) {
		taskCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return e.runWithTimeout(taskCtx, cancel, jc, &v1.TriggerTask{
			ExecutorHandler: handler, ExecutorTimeout: timeout,
		})
	}

	// 1. 超时：handler 响应 ctx，1s 超时后快速返回 timeout
	e.Register("slow", func(ctx context.Context, jc *JobContext, param string) (int, string) {
		select {
		case <-time.After(2 * time.Second):
			return 200, "done"
		case <-ctx.Done():
			return 500, "cancelled"
		}
	})
	start := time.Now()
	code, msg := run(1, "slow")
	if code != 500 || msg != "timeout" {
		t.Fatalf("超时应返回 500/timeout，got %d/%s", code, msg)
	}
	if d := time.Since(start); d > 1500*time.Millisecond {
		t.Fatalf("超时后应在宽限期内返回，实际 %v", d)
	}

	// 2. 正常完成：handler 快速返回 200
	e.Register("fast", func(ctx context.Context, jc *JobContext, param string) (int, string) {
		return 200, "ok"
	})
	code2, msg2 := run(10, "fast")
	if code2 != 200 || msg2 != "ok" {
		t.Fatalf("正常完成应返回 200/ok，got %d/%s", code2, msg2)
	}

	// 3. handler 不响应 ctx：1s 超时后 handler 仍跑完（2s），runWithTimeout 仍只返回 timeout
	e.Register("ignoreCtx", func(ctx context.Context, jc *JobContext, param string) (int, string) {
		time.Sleep(2 * time.Second) // 不检查 ctx
		return 200, "late"
	})
	code3, msg3 := run(1, "ignoreCtx")
	if code3 != 500 || msg3 != "timeout" {
		t.Fatalf("handler 不响应 ctx 时应返回 500/timeout（不返回 late），got %d/%s", code3, msg3)
	}

	// 4. ExecutorTimeout=0：无超时，等 handler 完成
	e.Register("notimeout", func(ctx context.Context, jc *JobContext, param string) (int, string) {
		time.Sleep(200 * time.Millisecond)
		return 200, "done"
	})
	code4, msg4 := run(0, "notimeout")
	if code4 != 200 || msg4 != "done" {
		t.Fatalf("无超时应返回 handler 结果，got %d/%s", code4, msg4)
	}
}
