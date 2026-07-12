// Package executor Go 执行器 SDK：连接 admin 的 ExecutorLink 双向流，
// 注册任务 handler，接收触发并执行，回传 TaskResult 与 LogChunk。心跳保活，断线自动重连。
//
// 用法：
//
//	ex := executor.New("http://localhost:9090", "my-app", logger)
//	ex.Register("handlerName", func(ctx, jc, param) (int, string) {
//	    jc.Log("开始执行")
//	    return 200, "ok"
//	})
//	ex.Run(ctx)
package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"go.uber.org/zap"
	"golang.org/x/net/http2"

	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
	"xxl-job-admin-go/pkg/gen/xxl/v1/xxlv1connect"
)

// HandlerFunc 任务处理函数：返回 (code, msg)，code 200 表示成功。
// jc.Log 用于向 admin 回传实时日志（前端 WebSocket 可查看）。
type HandlerFunc func(ctx context.Context, jc *JobContext, param string) (code int, msg string)

// JobContext 单次任务执行上下文，提供日志回传能力。
type JobContext struct {
	LogID  int64
	logMu  sync.Mutex
	line   int64
	sender func(*v1.ExecutorMessage) // 回传消息的发送函数
}

// Log 回传一行日志到 admin（前端实时日志 WebSocket 接收）。
// sender 为 nil 时静默忽略（便于测试与无回传场景构造 JobContext）。
func (j *JobContext) Log(msg string) {
	if j.sender == nil {
		return
	}
	j.logMu.Lock()
	defer j.logMu.Unlock()
	j.line++
	j.sender(&v1.ExecutorMessage{Payload: &v1.ExecutorMessage_LogChunk{LogChunk: &v1.LogChunk{
		LogId:      j.LogID,
		FromLine:   j.line,
		ToLine:     j.line + 1,
		LogContent: []byte(msg + "\n"),
		End:        false,
	}}})
}

// NewJobContext 构造任务上下文，供外部包（测试、自定义 handler）创建带日志回传的 JobContext。
// sender 为日志/结果消息发送函数（通常绑定到 executor 的 stream.Send）；为 nil 时 Log 静默忽略。
func NewJobContext(logID int64, sender func(*v1.ExecutorMessage)) *JobContext {
	return &JobContext{LogID: logID, sender: sender}
}

// linkStream 双向流类型别名。
type linkStream = *connect.BidiStreamForClient[v1.ExecutorMessage, v1.AdminMessage]

// Executor Go 执行器。
type Executor struct {
	serverAddr string
	appName    string
	version    string
	id         string
	handlers   map[string]HandlerFunc
	log        *zap.Logger
	mu         sync.RWMutex
	sendMu     sync.Mutex

	runMu        sync.Mutex
	running      map[int64]context.CancelFunc // logID → 取消函数（CancelTask 用）
	accessToken  string                       // 调度中心认证令牌
	enableScript bool                         // 是否启用 GLUE/脚本任务（默认禁用，安全考虑）
}

// New 创建执行器。accessToken 与 admin 配置一致（为空则不发送）。
func New(serverAddr, appName, accessToken string, log *zap.Logger) *Executor {
	hostname, _ := os.Hostname()
	return &Executor{
		serverAddr:  serverAddr,
		appName:     appName,
		version:     "go-1.0",
		id:          fmt.Sprintf("%s-%s-%d", appName, hostname, os.Getpid()),
		handlers:    map[string]HandlerFunc{},
		log:         log,
		running:     map[int64]context.CancelFunc{},
		accessToken: accessToken,
	}
}

// Register 注册任务 handler。
func (e *Executor) Register(name string, h HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[name] = h
}

// WithScriptTask 启用 GLUE Go(yaegi) 与 Shell/Python 脚本任务执行。
// 默认禁用：脚本/GLUE 可执行任意代码，仅在受控环境显式开启。详见 SECURITY.md。
func (e *Executor) WithScriptTask(enable bool) *Executor {
	e.enableScript = enable
	return e
}

// Run 连接 admin 并处理任务，阻塞直到 ctx 取消。断线自动重连。
func (e *Executor) Run(ctx context.Context) error {
	for {
		if err := e.runOnce(ctx); err != nil {
			e.log.Error("link error", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
			e.log.Info("reconnecting to admin")
		}
	}
}

// newH2CClient 创建支持 cleartext HTTP/2 的客户端（gRPC over h2c）。
// Go http.DefaultClient 仅支持 TLS 下的 HTTP/2，cleartext h2c 需显式配置。
func newH2CClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}
}

// tokenInterceptor 注入 X-XXL-Job-Access-Token 请求头（调度中心认证）。
func tokenInterceptor(token string) connect.Interceptor {
	return &tokenAuth{token: token}
}

type tokenAuth struct{ token string }

func (t *tokenAuth) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		req.Header().Set("X-XXL-Job-Access-Token", t.token)
		return next(ctx, req)
	}
}

func (t *tokenAuth) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		conn.RequestHeader().Set("X-XXL-Job-Access-Token", t.token)
		return conn
	}
}

func (t *tokenAuth) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func (e *Executor) runOnce(ctx context.Context) error {
	// runCtx 绑定本次连接生命周期：runOnce 返回时取消，停止心跳 goroutine（修复旧版心跳跨重连泄漏）。
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []connect.ClientOption{connect.WithGRPC()}
	if e.accessToken != "" {
		opts = append(opts, connect.WithInterceptors(tokenInterceptor(e.accessToken)))
	}
	client := xxlv1connect.NewExecutorLinkClient(newH2CClient(), e.serverAddr, opts...)
	stream := client.Link(runCtx)

	// 1. 注册
	e.mu.RLock()
	names := make([]string, 0, len(e.handlers))
	for n := range e.handlers {
		names = append(names, n)
	}
	e.mu.RUnlock()
	if err := e.send(stream, &v1.ExecutorMessage{
		ExecutorId: e.id,
		Payload: &v1.ExecutorMessage_Register{
			Register: &v1.RegisterRequest{AppName: e.appName, Version: e.version, Handlers: names},
		},
	}); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// 2. 心跳
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-t.C:
				_ = e.send(stream, &v1.ExecutorMessage{Payload: &v1.ExecutorMessage_Heartbeat{
					Heartbeat: &v1.Heartbeat{Timestamp: time.Now().UnixMilli()},
				}})
			}
		}
	}()

	// 3. 接收并分发
	for {
		msg, err := stream.Receive()
		if err != nil {
			return fmt.Errorf("receive: %w", err)
		}
		switch p := msg.Payload.(type) {
		case *v1.AdminMessage_RegisterAck:
			e.log.Info("registered", zap.Bool("accepted", p.RegisterAck.Accepted))
		case *v1.AdminMessage_Trigger:
			go e.handleTrigger(ctx, stream, p.Trigger)
		case *v1.AdminMessage_Cancel:
			e.cancelRun(p.Cancel.LogId)
		case *v1.AdminMessage_Config:
			// 配置更新（忽略）
		}
	}
}

// handleTrigger 执行任务并回传结果，支持超时、取消与日志回传。
// 结果唯一发送点：runWithTimeout 返回唯一 (code,msg)，handleTrigger 末尾发送一次。
func (e *Executor) handleTrigger(ctx context.Context, stream linkStream, t *v1.TriggerTask) {
	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	e.registerRun(t.LogId, cancel)
	defer e.unregisterRun(t.LogId)

	jc := NewJobContext(t.LogId, func(m *v1.ExecutorMessage) { _ = e.send(stream, m) })
	code, msg := e.runWithTimeout(taskCtx, cancel, jc, t)
	jc.Log("[result] code=" + strconv.Itoa(code) + " msg=" + msg)
	_ = e.send(stream, resultMsg(t.LogId, int32(code), msg))
}

// runWithTimeout 执行任务并返回唯一结果：超时返回 500/timeout，否则返回 handler 结果。
// 修复旧版超时后 handler 仍发第二条结果的双结果 bug：done 缓冲为 1，超时后 ctx 取消使
// handler goroutine 走 Done 分支丢弃结果。handler 不响应 ctx 时给 2s 宽限期等待退出，
// 超宽限则放弃等待（记录泄漏），但结果仍为唯一 timeout。
func (e *Executor) runWithTimeout(taskCtx context.Context, cancel context.CancelFunc, jc *JobContext, t *v1.TriggerTask) (int, string) {
	type result struct {
		code int
		msg  string
	}
	done := make(chan result, 1) // buffered：handler 返回后即使无人接收也不阻塞
	go func() {
		c, m := e.executeTask(taskCtx, jc, t)
		done <- result{c, m} // buffer(1) 不阻塞；结果唯一性由 handleTrigger 单发送点保证
	}()

	if t.ExecutorTimeout > 0 {
		select {
		case r := <-done:
			return r.code, r.msg
		case <-time.After(time.Duration(t.ExecutorTimeout) * time.Second):
			cancel()
			jc.Log("[timeout] 任务超时中断")
			// 等 handler 退出（宽限 2s），避免 goroutine 永久泄漏；handler 应响应 ctx
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				e.log.Warn("handler 未响应 ctx 取消，放弃等待", zap.Int64("log_id", t.LogId))
			}
			return 500, "timeout"
		}
	}
	r := <-done
	return r.code, r.msg
}

// registerRun/unregisterRun/cancelRun 维护运行中任务的取消函数。
func (e *Executor) registerRun(logID int64, cancel context.CancelFunc) {
	e.runMu.Lock()
	e.running[logID] = cancel
	e.runMu.Unlock()
}

func (e *Executor) unregisterRun(logID int64) {
	e.runMu.Lock()
	delete(e.running, logID)
	e.runMu.Unlock()
}

func (e *Executor) cancelRun(logID int64) {
	e.runMu.Lock()
	defer e.runMu.Unlock()
	if cancel, ok := e.running[logID]; ok {
		cancel()
		e.log.Info("task cancelled", zap.Int64("log_id", logID))
	}
}

// executeTask 按任务类型执行：GLUE Go(yaegi)、Shell/Python 脚本(os/exec) 或 BEAN(handler)。
func (e *Executor) executeTask(ctx context.Context, jc *JobContext, t *v1.TriggerTask) (int, string) {
	switch {
	case t.GlueType == 1 && t.GlueSource != "":
		if !e.enableScript {
			return 500, "GLUE 任务未启用：执行器需显式调用 WithScriptTask(true)"
		}
		jc.Log("[glue] 开始执行 GLUE Go 代码")
		return runGlue(t.GlueSource, t.ExecutorParams)
	case t.GlueType == 2 && t.GlueSource != "":
		if !e.enableScript {
			return 500, "脚本任务未启用：执行器需显式调用 WithScriptTask(true)"
		}
		jc.Log("[script] 开始执行 Shell 脚本")
		return runScript("sh", t.GlueSource, t.ExecutorParams, jc)
	case t.GlueType == 3 && t.GlueSource != "":
		if !e.enableScript {
			return 500, "脚本任务未启用：执行器需显式调用 WithScriptTask(true)"
		}
		jc.Log("[script] 开始执行 Python 脚本")
		return runScript("python3", t.GlueSource, t.ExecutorParams, jc)
	}
	e.mu.RLock()
	h, ok := e.handlers[t.ExecutorHandler]
	e.mu.RUnlock()
	if !ok {
		return 500, "handler not found: " + t.ExecutorHandler
	}
	return h(ctx, jc, t.ExecutorParams)
}

// runGlue 用 yaegi 解释执行 GLUE Go 代码。
// glue_source 须定义: func Run(param string) (int, string)
func runGlue(src, param string) (int, string) {
	i := interp.New(interp.Options{})
	if err := i.Use(stdlib.Symbols); err != nil {
		return 500, "glue use stdlib: " + err.Error()
	}
	if _, err := i.Eval(src); err != nil {
		return 500, "glue eval: " + err.Error()
	}
	v, err := i.Eval("Run")
	if err != nil {
		return 500, "glue Run not found: " + err.Error()
	}
	fn, ok := v.Interface().(func(string) (int, string))
	if !ok {
		return 500, "glue Run signature mismatch (expect func(string)(int,string))"
	}
	return fn(param)
}

// runScript 用系统解释器执行 Shell/Python 脚本，输出通过 jc.Log 回传。
func runScript(interpreter, src, param string, jc *JobContext) (int, string) {
	cmd := exec.Command(interpreter, "-c", src)
	cmd.Env = append(os.Environ(), "XXL_JOB_PARAM="+param)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		jc.Log(string(out))
	}
	if err != nil {
		return 500, "script error: " + err.Error()
	}
	return 200, "script ok"
}

// send 并发安全的 Send。
func (e *Executor) send(stream linkStream, msg *v1.ExecutorMessage) error {
	e.sendMu.Lock()
	defer e.sendMu.Unlock()
	return stream.Send(msg)
}

func resultMsg(logID int64, code int32, msg string) *v1.ExecutorMessage {
	return &v1.ExecutorMessage{Payload: &v1.ExecutorMessage_TaskResult{TaskResult: &v1.TaskResult{
		LogId: logID, HandleCode: code, HandleMsg: msg, HandleTime: time.Now().UnixMilli(),
	}}}
}
