package handlers

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"xxl-job-admin-go/pkg/executor"
)

type shellExecParam struct {
	Cmd     string `json:"cmd"`
	Timeout int    `json:"timeout"` // 秒，0=默认 60
}

// ShellExec 执行 shell 命令，回传 stdout/stderr。
func ShellExec(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
	var p shellExecParam
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return 500, "参数解析失败: " + err.Error()
	}
	if p.Cmd == "" {
		return 500, "cmd 必填"
	}
	timeout := time.Duration(p.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", p.Cmd)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		jc.Log(string(out))
	}
	if err != nil {
		return 500, "执行失败: " + err.Error()
	}
	return 200, string(out)
}
