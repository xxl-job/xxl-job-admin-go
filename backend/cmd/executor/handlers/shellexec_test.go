package handlers

import (
	"context"
	"strings"
	"testing"

	"xxl-job-admin-go/pkg/executor"
)

func TestShellExec(t *testing.T) {
	jc := &executor.JobContext{LogID: 1}
	code, msg := ShellExec(context.Background(), jc, `{"cmd":"echo hello","timeout":5}`)
	if code != 200 || !strings.Contains(msg, "hello") {
		t.Fatalf("echo: got %d/%s", code, msg)
	}
	// 缺 cmd
	code2, _ := ShellExec(context.Background(), jc, `{}`)
	if code2 != 500 {
		t.Fatalf("缺 cmd 应 500: got %d", code2)
	}
}
