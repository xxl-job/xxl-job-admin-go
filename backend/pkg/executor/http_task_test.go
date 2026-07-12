package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
)

// testJC 构造一个丢弃日志的 JobContext，仅用于测试回传不报错。
func testJC() *JobContext {
	return &JobContext{LogID: 1, sender: func(*v1.ExecutorMessage) {}}
}

func TestRunHTTP_GetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	code, msg := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: `{"url":"` + srv.URL + `"}`}, testJC())
	if code != 200 {
		t.Fatalf("expect 200, got %d: %s", code, msg)
	}
}

func TestRunHTTP_PostParamRender(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := `{"url":"` + srv.URL + `","method":"POST","body":"{\"p\":\"{{param}}\"}"}`
	code, _ := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: cfg, ExecutorParams: "hello"}, testJC())
	if code != 200 {
		t.Fatal("expect 200")
	}
	if want := `{"p":"hello"}`; gotBody != want {
		t.Fatalf("body not rendered: got %q want %q", gotBody, want)
	}
}

func TestRunHTTP_ExpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
	}))
	defer srv.Close()

	// 期望 [200]，实际 201 → 不命中 → 失败
	cfg := `{"url":"` + srv.URL + `","expected_status":[200]}`
	if code, _ := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: cfg}, testJC()); code != 500 {
		t.Fatalf("expect 500 for unmatched status, got %d", code)
	}
	// 期望 [200,201] → 命中 → 成功
	cfg2 := `{"url":"` + srv.URL + `","expected_status":[200,201]}`
	if code, _ := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: cfg2}, testJC()); code != 200 {
		t.Fatalf("expect 200 for matched status, got %d", code)
	}
}

func TestRunHTTP_Non2xxFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	// 无 expected_status，默认 2xx 才算成功 → 500 失败
	code, _ := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: `{"url":"` + srv.URL + `"}`}, testJC())
	if code != 500 {
		t.Fatalf("expect 500 for 5xx, got %d", code)
	}
}

func TestRunHTTP_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	code, msg := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: `{"url":"` + srv.URL + `","timeout":1}`}, testJC())
	if code != 500 {
		t.Fatalf("expect 500 on timeout, got %d: %s", code, msg)
	}
}

func TestRunHTTP_BadConfig(t *testing.T) {
	// 非 JSON
	if code, _ := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: "not-json"}, testJC()); code != 500 {
		t.Fatal("expect 500 for invalid json")
	}
	// url 空
	if code, _ := runHTTP(context.Background(), &v1.TriggerTask{GlueSource: `{"url":""}`}, testJC()); code != 500 {
		t.Fatal("expect 500 for empty url")
	}
}

func TestExecuteTask_HTTPDisabled(t *testing.T) {
	ex := &Executor{} // enableHTTP 默认 false
	code, msg := ex.executeTask(
		context.Background(),
		&JobContext{sender: func(*v1.ExecutorMessage) {}},
		&v1.TriggerTask{GlueType: 4, GlueSource: `{"url":"http://example.com"}`},
	)
	if code != 500 {
		t.Fatalf("expect 500 when HTTP disabled, got %d: %s", code, msg)
	}
}

func TestExecuteTask_HTTPEnabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ex := (&Executor{}).WithHTTPTask(true)
	code, _ := ex.executeTask(
		context.Background(),
		&JobContext{sender: func(*v1.ExecutorMessage) {}},
		&v1.TriggerTask{GlueType: 4, GlueSource: `{"url":"` + srv.URL + `"}`},
	)
	if code != 200 {
		t.Fatalf("expect 200 when HTTP enabled, got %d", code)
	}
}

// TestRenderHTTPVars 占位符渲染（单元级）。
func TestRenderHTTPVars(t *testing.T) {
	got := renderHTTPVars("p={{param}} i={{shardIndex}}/{{shardTotal}} l={{logId}}", &v1.TriggerTask{
		ExecutorParams: "x", ShardIndex: 2, ShardTotal: 5, LogId: 99,
	})
	if want := "p=x i=2/5 l=99"; got != want {
		t.Fatalf("render mismatch: got %q want %q", got, want)
	}
}
