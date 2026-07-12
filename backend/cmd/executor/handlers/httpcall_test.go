package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"xxl-job-admin-go/pkg/executor"
	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
)

func TestHTTPCall(t *testing.T) {
	// mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fail") == "1" {
			w.WriteHeader(500)
			return
		}
		_, _ = w.Write([]byte("hello " + r.Method))
	}))
	defer srv.Close()

	jc := executor.NewJobContext(1, func(m *v1.ExecutorMessage) {})

	// GET 200
	code, msg := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`"}`)
	if code != 200 || msg != "hello GET" {
		t.Fatalf("GET 200: got %d/%s", code, msg)
	}

	// POST
	code2, _ := HTTPCall(context.Background(), jc, `{"method":"POST","url":"`+srv.URL+`","body":"x"}`)
	if code2 != 200 {
		t.Fatalf("POST 200: got %d", code2)
	}

	// 服务端 500 透传
	code3, _ := HTTPCall(context.Background(), jc, `{"method":"GET","url":"`+srv.URL+`?fail=1"}`)
	if code3 != 500 {
		t.Fatalf("500 透传: got %d", code3)
	}

	// 参数缺失
	code4, _ := HTTPCall(context.Background(), jc, `{}`)
	if code4 != 500 {
		t.Fatalf("缺 url 应 500: got %d", code4)
	}
}
