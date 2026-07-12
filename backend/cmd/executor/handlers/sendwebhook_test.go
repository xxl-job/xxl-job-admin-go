package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"xxl-job-admin-go/pkg/executor"
)

func TestSendWebhook(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	jc := &executor.JobContext{LogID: 1}
	code, _ := SendWebhook(context.Background(), jc, `{"url":"`+srv.URL+`","json":{"msg":"hi"}}`)
	if code != 200 {
		t.Fatalf("200: got %d", code)
	}
	if !contains(gotBody, `"msg":"hi"`) {
		t.Fatalf("webhook 应发送 json，got %s", gotBody)
	}
	// 缺 url
	code2, _ := SendWebhook(context.Background(), jc, `{}`)
	if code2 != 500 {
		t.Fatalf("缺 url 应 500: got %d", code2)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && stringContains(s, sub)))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
