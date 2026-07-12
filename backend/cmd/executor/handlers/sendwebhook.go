package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"xxl-job-admin-go/pkg/executor"
)

type sendWebhookParam struct {
	URL  string          `json:"url"`
	JSON json.RawMessage `json:"json"`
}

// SendWebhook 通用 webhook 通知（钉钉/企微/飞书 JSON 直接发），返回响应状态码+体。
func SendWebhook(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
	var p sendWebhookParam
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return 500, "参数解析失败: " + err.Error()
	}
	if p.URL == "" {
		return 500, "url 必填"
	}
	req, err := http.NewRequestWithContext(ctx, "POST", p.URL, bytes.NewReader(p.JSON))
	if err != nil {
		return 500, "构造请求失败: " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 500, "发送失败: " + err.Error()
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	jc.Log("webhook -> " + resp.Status)
	return resp.StatusCode, string(b)
}
