package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
)

// httpTaskConfig HTTP 任务（glue_type=4）配置，序列化为 JSON 存入 job.glue_source。
// 由 admin 前端结构化表单生成；执行器解析后按此发起 HTTP 请求。
type httpTaskConfig struct {
	URL            string            `json:"url"`                       // 目标地址，必填
	Method         string            `json:"method,omitempty"`          // 默认 GET；支持 GET/POST/PUT/DELETE/PATCH/HEAD
	Headers        map[string]string `json:"headers,omitempty"`         // 自定义请求头
	Body           string            `json:"body,omitempty"`            // 请求体（原始字符串）
	Timeout        int32             `json:"timeout,omitempty"`         // 秒；0 → 回退 ExecutorTimeout → 30
	ExpectedStatus []int             `json:"expected_status,omitempty"` // 期望状态码；空 → 任意 2xx 算成功
}

// HTTPRequest HTTP 请求配置（公开）。供 glue_type=4 任务与内置 httpCall/sendWebhook handler 复用，
// 避免三处重复实现「构造请求→发送→读响应」。
type HTTPRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    string
	Timeout int32 // 秒；<=0 用默认 30
}

const (
	defaultHTTPTimeoutSecs = 30    // HTTP 任务/handler 默认超时（秒），任务与 Job 都未配置时兜底
	httpRespBodyLimit      = 1 << 16 // 读取响应体上限 64KB，防止大响应撑爆 handle_msg
)

// DoHTTP 发起 HTTP 请求，返回 (HTTP 状态码, 响应体, 错误)。
// 仅负责方法默认值/超时/发送/读响应，不做成败判定（判定由 IsHTTPSuccess 完成）。
// 超时派生自 ctx；body 仅 POST/PUT/PATCH 携带。响应体最多读 httpRespBodyLimit。
func DoHTTP(ctx context.Context, r HTTPRequest) (status int, body string, err error) {
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if method == "" {
		method = http.MethodGet
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeoutSecs
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if r.Body != "" && methodAllowsBody(method) {
		bodyReader = strings.NewReader(r.Body)
	}
	req, err := http.NewRequestWithContext(cctx, method, r.URL, bodyReader)
	if err != nil {
		return 0, "", err
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, httpRespBodyLimit))
	return resp.StatusCode, string(b), nil
}

// IsHTTPSuccess 判定 HTTP 状态码是否算任务成功：
// 期望列表非空 → 命中才算成功；空 → 任意 2xx 算成功。
// 这是「HTTP 状态码 → 任务结果码」的统一归一化基准（避免把 201/204 当失败）。
func IsHTTPSuccess(status int, expected []int) bool {
	if len(expected) > 0 {
		for _, e := range expected {
			if status == e {
				return true
			}
		}
		return false
	}
	return status >= 200 && status < 300
}

// Truncate 截断字符串到 n 字节并加省略标记，用于 handle_msg 等不宜过长的场景。
func Truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(截断)"
}

// runHTTP 执行 glue_type=4 HTTP 任务：解析 glue_source JSON，渲染占位符，
// 调 DoHTTP 发请求，按 IsHTTPSuccess 判定成败归一化为 200/500。
// 请求/响应摘要通过 jc.Log 回传，前端实时日志可见。
// ctx 来自 executeTask，任务取消（cancelRun）会中断进行中的 HTTP 请求。
func runHTTP(ctx context.Context, t *v1.TriggerTask, jc *JobContext) (int, string) {
	var cfg httpTaskConfig
	if err := json.Unmarshal([]byte(t.GlueSource), &cfg); err != nil {
		return 500, "http config 解析失败: " + err.Error()
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return 500, "http config 非法: url 不能为空"
	}

	render := func(s string) string { return renderHTTPVars(s, t) }
	headers := make(map[string]string, len(cfg.Headers))
	for k, v := range cfg.Headers {
		headers[k] = render(v)
	}

	// 超时：cfg.Timeout > 0 → 用之；否则回退 ExecutorTimeout；都为 0 由 DoHTTP 兜底默认。
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = t.ExecutorTimeout
	}
	actualTimeout := timeout
	if actualTimeout <= 0 {
		actualTimeout = defaultHTTPTimeoutSecs
	}

	method := strings.ToUpper(strings.TrimSpace(cfg.Method))
	if method == "" {
		method = http.MethodGet
	}
	jc.Log(fmt.Sprintf("[http] %s %s (timeout=%ds)", method, render(cfg.URL), actualTimeout))

	status, body, err := DoHTTP(ctx, HTTPRequest{
		URL:     render(cfg.URL),
		Method:  method,
		Headers: headers,
		Body:    render(cfg.Body),
		Timeout: timeout,
	})
	if err != nil {
		jc.Log("[http] 请求错误: " + err.Error())
		return 500, "http request: " + err.Error()
	}
	jc.Log(fmt.Sprintf("[http] 响应 HTTP %d, body: %s", status, Truncate(body, 500)))

	if !IsHTTPSuccess(status, cfg.ExpectedStatus) {
		return 500, fmt.Sprintf("http %d: %s", status, Truncate(body, 500))
	}
	return 200, fmt.Sprintf("ok: HTTP %d", status)
}

// renderHTTPVars 替换占位符：{{param}} {{shardIndex}} {{shardTotal}} {{logId}}。
func renderHTTPVars(s string, t *v1.TriggerTask) string {
	r := strings.NewReplacer(
		"{{param}}", t.ExecutorParams,
		"{{shardIndex}}", fmt.Sprintf("%d", t.ShardIndex),
		"{{shardTotal}}", fmt.Sprintf("%d", t.ShardTotal),
		"{{logId}}", fmt.Sprintf("%d", t.LogId),
	)
	return r.Replace(s)
}

// methodAllowsBody 判断该方法是否应携带请求体（避免 GET/HEAD/DELETE 误带 body）。
func methodAllowsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	}
	return false
}
