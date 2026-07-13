package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"xxl-job-admin-go/pkg/executor"
	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
)

func TestHTTPCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ?status=N → 直接返回状态码 N（用于测试判定归一化）；否则 200 + body
		if s := r.URL.Query().Get("status"); s != "" {
			if code, err := strconv.Atoi(s); err == nil {
				w.WriteHeader(code)
				return
			}
		}
		_, _ = w.Write([]byte("hello " + r.Method))
	}))
	defer srv.Close()

	jc := executor.NewJobContext(1, func(m *v1.ExecutorMessage) {})

	// GET 200 → 成功，返回响应体
	code, msg := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`"}`)
	if code != 200 || msg != "hello GET" {
		t.Fatalf("GET 200: got %d/%s", code, msg)
	}

	// POST 200 → 成功
	code2, _ := HTTPCall(context.Background(), jc, `{"method":"POST","url":"`+srv.URL+`","body":"x"}`)
	if code2 != 200 {
		t.Fatalf("POST 200: got %d", code2)
	}

	// 201/204 归一化为成功（修正：早期版本把 HTTP 原始码当任务码，201 会判失败）
	if code, _ := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`?status=201"}`); code != 200 {
		t.Fatalf("201 应归一化为成功: got %d", code)
	}
	if code, _ := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`?status=204"}`); code != 200 {
		t.Fatalf("204 应归一化为成功: got %d", code)
	}

	// 4xx/5xx 归一化为失败
	if code, _ := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`?status=404"}`); code != 500 {
		t.Fatalf("404 应判失败: got %d", code)
	}
	if code, _ := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`?status=500"}`); code != 500 {
		t.Fatalf("500 应判失败: got %d", code)
	}

	// 参数缺失
	code4, _ := HTTPCall(context.Background(), jc, `{}`)
	if code4 != 500 {
		t.Fatalf("缺 url 应 500: got %d", code4)
	}
}
