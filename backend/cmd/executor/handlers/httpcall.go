package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"xxl-job-admin-go/pkg/executor"
)

// httpCallParam HTTP 调用参数。
type httpCallParam struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Timeout int               `json:"timeout"` // 秒，0=默认 30
}

// HTTPCall 通用 HTTP 请求 handler。成功（2xx）→ 200 + 响应体；否则 → 500。
// 请求/超时/读响应复用 executor.DoHTTP；成败判定统一走 executor.IsHTTPSuccess
// （修正早期版本把 HTTP 原始状态码当任务结果码、导致 201/204 被判失败的缺陷）。
func HTTPCall(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
	var p httpCallParam
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return 500, "参数解析失败: " + err.Error()
	}
	if p.URL == "" {
		return 500, "url 必填"
	}
	status, body, err := executor.DoHTTP(ctx, executor.HTTPRequest{
		URL:     p.URL,
		Method:  p.Method,
		Headers: p.Headers,
		Body:    p.Body,
		Timeout: int32(p.Timeout),
	})
	if err != nil {
		return 500, "请求失败: " + err.Error()
	}
	method := p.Method
	if method == "" {
		method = "GET"
	}
	jc.Log("HTTP " + method + " " + p.URL + " -> " + strconv.Itoa(status))
	if !executor.IsHTTPSuccess(status, nil) {
		return 500, fmt.Sprintf("http %d: %s", status, executor.Truncate(body, 500))
	}
	return 200, body
}
