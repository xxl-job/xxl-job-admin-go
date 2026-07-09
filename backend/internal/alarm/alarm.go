// Package alarm 告警：任务失败时触发，支持多渠道扩展（邮件/钉钉/企微）。
// 默认 LogNotifier 仅记录日志；生产可注入 EmailNotifier 等真实渠道。
package alarm

import (
	"context"

	"go.uber.org/zap"
)

// Event 告警事件。
type Event struct {
	JobID      uint64
	JobDesc    string
	LogID      int64
	HandleCode int16
	HandleMsg  string
	Address    string
}

// Notifier 告警通知接口（SOLID-O：可扩展新渠道而无需修改既有代码）。
type Notifier interface {
	Notify(ctx context.Context, e Event) error
}

// MultiNotifier 多渠道聚合，任一失败仅记录不影响其他渠道。
type MultiNotifier struct {
	notifiers []Notifier
	log       *zap.Logger
}

// NewMulti 创建多渠道告警器。
func NewMulti(log *zap.Logger, notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers, log: log}
}

// Notify 逐个通知。
func (m *MultiNotifier) Notify(ctx context.Context, e Event) error {
	for _, n := range m.notifiers {
		if err := n.Notify(ctx, e); err != nil {
			m.log.Error("alarm notify failed", zap.Error(err))
		}
	}
	return nil
}

// LogNotifier 日志告警（默认，无外部依赖）。
type LogNotifier struct{ Log *zap.Logger }

// Notify 记录告警日志。
func (n *LogNotifier) Notify(_ context.Context, e Event) error {
	n.Log.Warn("job alarm",
		zap.Uint64("job_id", e.JobID),
		zap.Int64("log_id", e.LogID),
		zap.Int16("code", e.HandleCode),
		zap.String("msg", e.HandleMsg),
		zap.String("addr", e.Address))
	return nil
}

// EmailNotifier 邮件告警占位（P6 接 net/smtp + 模板）。
type EmailNotifier struct {
	Log *zap.Logger
}

// Notify 邮件告警（当前仅记日志）。
func (n *EmailNotifier) Notify(_ context.Context, e Event) error {
	n.Log.Info("email alarm (stub)", zap.Uint64("job_id", e.JobID), zap.String("addr", e.Address))
	return nil
}
