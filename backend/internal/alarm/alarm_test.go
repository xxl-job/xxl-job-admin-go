package alarm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// TestLogNotifier 默认日志告警通道不报错。
func TestLogNotifier(t *testing.T) {
	n := &LogNotifier{Log: zap.NewNop()}
	if err := n.Notify(context.Background(), Event{JobID: 1, JobDesc: "demo"}); err != nil {
		t.Fatalf("LogNotifier.Notify: %v", err)
	}
}

// TestWebhookNotifier 验证钉钉/企微/飞书 webhook payload 格式与消息内容。
func TestWebhookNotifier(t *testing.T) {
	for _, platform := range []string{"dingtalk", "wechat", "feishu"} {
		t.Run(platform, func(t *testing.T) {
			var got map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &got)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			n := &WebhookNotifier{Platform: platform, URL: srv.URL, Client: srv.Client(), Log: zap.NewNop()}
			if err := n.Notify(context.Background(), Event{JobID: 7, JobDesc: "demo", HandleMsg: "boom"}); err != nil {
				t.Fatalf("Notify: %v", err)
			}
			if got == nil {
				t.Fatal("webhook 未收到 payload")
			}
			// 提取各平台 text 字段
			var text string
			if platform == "feishu" {
				c, ok := got["content"].(map[string]any)
				if !ok {
					t.Fatal("feishu 缺 content")
				}
				text, _ = c["text"].(string)
			} else {
				c, ok := got["text"].(map[string]any)
				if !ok {
					t.Fatal("缺 text")
				}
				text, _ = c["content"].(string)
			}
			if !strings.Contains(text, "任务#7") || !strings.Contains(text, "boom") {
				t.Errorf("告警文本不含任务ID或错误消息: %q", text)
			}
		})
	}
}

// TestWebhookNotifier_HTTPError 上游非 2xx 应返回错误。
func TestWebhookNotifier_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	n := &WebhookNotifier{Platform: "dingtalk", URL: srv.URL, Client: srv.Client(), Log: zap.NewNop()}
	if err := n.Notify(context.Background(), Event{JobID: 1}); err == nil {
		t.Fatal("HTTP 500 应返回错误")
	}
}

// TestEmailNotifier_Unconfigured 未配置 SMTP 应静默跳过。
func TestEmailNotifier_Unconfigured(t *testing.T) {
	n := &EmailNotifier{} // Host 空
	if err := n.Notify(context.Background(), Event{JobID: 1}); err != nil {
		t.Fatalf("未配置应跳过(nil error)，got %v", err)
	}
}

// TestMultiNotifier 验证多渠道聚合：每个渠道都被调用，单渠道失败不影响其他。
func TestMultiNotifier(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// 故意混入一个失败 URL，验证不影响 webhook 渠道
	bad := &WebhookNotifier{Platform: "dingtalk", URL: "http://127.0.0.1:0/notexist", Client: srv.Client(), Log: zap.NewNop()}
	good := &WebhookNotifier{Platform: "dingtalk", URL: srv.URL, Client: srv.Client(), Log: zap.NewNop()}
	m := NewMulti(zap.NewNop(), &LogNotifier{Log: zap.NewNop()}, bad, good)

	if err := m.Notify(context.Background(), Event{JobID: 1}); err != nil {
		t.Fatalf("MultiNotifier: %v", err)
	}
	if hits != 1 {
		t.Errorf("good webhook 应被命中1次，实际 %d", hits)
	}
}
