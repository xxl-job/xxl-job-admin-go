package trigger

import (
	"testing"

	"go.uber.org/zap"

	v1 "xxl-job-admin-go/internal/gen/xxl/v1"
	"xxl-job-admin-go/internal/registry"
)

// TestSendCancelDispatchesCancelTask 验证 cover-early 阻塞策略下，
// trackStart 记录的执行器连接能收到 CancelTask；trackDone 后不再发送。
func TestSendCancelDispatchesCancelTask(t *testing.T) {
	trig := NewTrigger(nil, registry.NewMemory(), NewBlockTracker(), zap.NewNop())

	// 构造一个在线执行器连接（带缓冲发送队列）
	send := make(chan *v1.AdminMessage, 4)
	conn := &registry.ExecutorConn{
		ID:      "e1",
		AppName: "app",
		Address: "127.0.0.1:9999",
		Send:    send,
	}

	const jobID uint64 = 1
	const logID int64 = 100

	// 模拟首次触发：记录运行中实例与其连接
	trig.trackStart(jobID, logID, conn)

	// cover-early 取消旧实例
	trig.sendCancel(jobID, logID)

	select {
	case msg := <-send:
		c, ok := msg.Payload.(*v1.AdminMessage_Cancel)
		if !ok {
			t.Fatalf("期望 AdminMessage_Cancel，实际 %T", msg.Payload)
		}
		if c.Cancel.JobId != int64(jobID) || c.Cancel.LogId != logID {
			t.Fatalf("CancelTask 字段错误：job=%d log=%d", c.Cancel.JobId, c.Cancel.LogId)
		}
	default:
		t.Fatal("期望已下发 CancelTask，但发送队列为空")
	}

	// 实例完成后再次取消：无运行中连接，不应发送
	trig.trackDone(jobID, logID)
	trig.sendCancel(jobID, logID)
	select {
	case <-send:
		t.Fatal("trackDone 后不应再下发消息")
	default:
		// 预期：队列为空
	}
}

// TestSendCancelNoRunningConn 验证无运行中连接时安全跳过（不 panic）。
func TestSendCancelNoRunningConn(t *testing.T) {
	trig := NewTrigger(nil, registry.NewMemory(), NewBlockTracker(), zap.NewNop())
	// 未 trackStart 直接 cancel，应仅记日志、不 panic
	trig.sendCancel(1, 999)
}
