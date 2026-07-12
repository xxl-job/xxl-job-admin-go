package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

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

// HTTPCall 通用 HTTP 请求 handler，返回 HTTP 状态码 + 响应体。
func HTTPCall(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
	var p httpCallParam
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return 500, "参数解析失败: " + err.Error()
	}
	if p.URL == "" {
		return 500, "url 必填"
	}
	if p.Method == "" {
		p.Method = "GET"
	}
	timeout := time.Duration(p.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	cli := &http.Client{Timeout: timeout}
	var body io.Reader
	if p.Body != "" {
		body = strings.NewReader(p.Body)
	}
	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, body)
	if err != nil {
		return 500, "构造请求失败: " + err.Error()
	}
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return 500, "请求失败: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	jc.Log("HTTP " + p.Method + " " + p.URL + " -> " + resp.Status)
	return resp.StatusCode, string(b)
}
