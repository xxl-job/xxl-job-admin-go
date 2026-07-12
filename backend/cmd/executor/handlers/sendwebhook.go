package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"xxl-job-admin-go/pkg/executor"
)

type sendWebhookParam struct {
	URL  string          `json:"url"`
	JSON json.RawMessage `json:"json"`
}

// SendWebhook 通用 webhook 通知（POST application/json），成功（2xx）→ 200 + 响应体；否则 → 500。
// 复用 executor.DoHTTP（含默认 30s 超时，修正早期版本用 http.DefaultClient 无超时的挂死隐患）。
func SendWebhook(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
	var p sendWebhookParam
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return 500, "参数解析失败: " + err.Error()
	}
	if p.URL == "" {
		return 500, "url 必填"
	}
	status, body, err := executor.DoHTTP(ctx, executor.HTTPRequest{
		URL:     p.URL,
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    string(p.JSON),
	})
	if err != nil {
		return 500, "发送失败: " + err.Error()
	}
	jc.Log("webhook -> " + strconv.Itoa(status))
	if !executor.IsHTTPSuccess(status, nil) {
		return 500, fmt.Sprintf("http %d: %s", status, executor.Truncate(body, 500))
	}
	return 200, body
}
