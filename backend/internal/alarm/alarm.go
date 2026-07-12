// Package alarm 告警：任务失败时触发，支持多渠道（日志/钉钉/企微/飞书 webhook + 邮件 SMTP）。
// 默认 LogNotifier 仅记录日志；按配置注入 WebhookNotifier/EmailNotifier 启用真实渠道。
package alarm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

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

// formatMessage 统一告警文本。
func formatMessage(e Event) string {
	return fmt.Sprintf("[XXL-JOB 告警] 任务#%d %s 执行失败(code=%d)\n日志ID: %d\n执行器: %s\n消息: %s",
		e.JobID, e.JobDesc, e.HandleCode, e.LogID, e.Address, e.HandleMsg)
}

// WebhookNotifier 通用 webhook 告警（钉钉/企微/飞书 text 消息）。
type WebhookNotifier struct {
	Platform string // dingtalk | wechat | feishu
	URL      string
	Client   *http.Client
	Log      *zap.Logger
}

// Notify POST text 消息到 webhook。
func (n *WebhookNotifier) Notify(ctx context.Context, e Event) error {
	msg := formatMessage(e)
	var payload map[string]any
	switch n.Platform {
	case "feishu":
		payload = map[string]any{"msg_type": "text", "content": map[string]any{"text": msg}}
	default: // dingtalk / wechat 同构
		payload = map[string]any{"msgtype": "text", "text": map[string]any{"content": msg}}
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook %s: HTTP %d", n.Platform, resp.StatusCode)
	}
	return nil
}

// EmailNotifier 邮件告警（SMTP）。Host 为空时跳过（未配置）。
type EmailNotifier struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	To       []string
	Log      *zap.Logger
}

// Notify 经 SMTP 发送告警邮件。
func (n *EmailNotifier) Notify(_ context.Context, e Event) error {
	if n.Host == "" || len(n.To) == 0 {
		return nil // 未配置，跳过
	}
	subject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(
		fmt.Sprintf("[XXL-JOB 告警] 任务#%d %s 执行失败", e.JobID, e.JobDesc))))
	body := formatMessage(e)
	msg := []byte(strings.Join([]string{
		"From: " + n.From,
		"To: " + strings.Join(n.To, ","),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n"))
	addr := fmt.Sprintf("%s:%d", n.Host, n.Port)
	var auth smtp.Auth
	if n.User != "" {
		auth = smtp.PlainAuth("", n.User, n.Password, n.Host)
	}
	return smtp.SendMail(addr, auth, n.From, n.To, msg)
}
