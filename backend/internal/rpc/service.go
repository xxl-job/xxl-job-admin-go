// Package rpc 实现 ConnectRPC 服务。
package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/alarm"
	"xxl-job-admin-go/internal/circuit"
	"xxl-job-admin-go/internal/logstream"
	"xxl-job-admin-go/internal/metrics"
	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
	"xxl-job-admin-go/internal/trigger"
	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
	"xxl-job-admin-go/pkg/gen/xxl/v1/xxlv1connect"
)

// errNotImpl 占位错误。
var errNotImpl = errors.New("not implemented")

// TaskCompleter 任务完成通知接口。
type TaskCompleter interface {
	OnTaskDone(jobID uint64, logID int64, handleCode int16)
	// ReleaseByLogID 在仅知 logID（task_run 查询失败等异常）时按 logID 释放阻塞槽位。
	ReleaseByLogID(logID int64)
}

// ExecutorLinkService 实现 ExecutorLink 双向流。
type ExecutorLinkService struct {
	reg       registry.Registry
	db        *gorm.DB
	completer TaskCompleter
	alerter   alarm.Notifier
	breaker   *circuit.Breaker
	hub       *logstream.Hub
	log       *zap.Logger
}

// NewExecutorLinkService 创建执行器链路服务。
func NewExecutorLinkService(reg registry.Registry, db *gorm.DB, completer TaskCompleter, alerter alarm.Notifier, hub *logstream.Hub, log *zap.Logger) *ExecutorLinkService {
	return &ExecutorLinkService{reg: reg, db: db, completer: completer, alerter: alerter, hub: hub, log: log}
}

// SetBreaker 注入熔断器（handleTaskResult 按结果更新 app 熔断状态）。
func (s *ExecutorLinkService) SetBreaker(b *circuit.Breaker) { s.breaker = b }

// Link 处理执行器长连接。
func (s *ExecutorLinkService) Link(ctx context.Context, stream *connect.BidiStream[v1.ExecutorMessage, v1.AdminMessage]) error {
	peer := stream.Peer()
	s.log.Info("executor connected", zap.String("addr", peer.Addr))

	sendCh := make(chan *v1.AdminMessage, 100)
	done := make(chan struct{})
	conn := &registry.ExecutorConn{Send: sendCh, Done: done, Address: peer.Addr}

	var app, id string
	registered := false

	go func() {
		for {
			select {
			case msg := <-sendCh:
				if err := stream.Send(msg); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	defer func() {
		close(done)
		if registered {
			s.reg.OnDisconnect(app, id)
			s.log.Info("executor disconnected", zap.String("app", app), zap.String("id", id))
		}
	}()

	for {
		msg, err := stream.Receive()
		if err != nil {
			return err
		}
		switch p := msg.Payload.(type) {
		case *v1.ExecutorMessage_Register:
			app, id = p.Register.AppName, msg.ExecutorId
			conn.ID = id
			conn.AppName = app
			conn.Handlers = p.Register.Handlers
			conn.Meta = p.Register.Meta
			conn.Cluster = p.Register.Meta["cluster"] // 流量泳道：从注册 meta 提取 cluster 标签
			if conn.Cluster == "" {
				conn.Cluster = "default"
			}
			s.reg.OnConnect(app, id, conn)
			registered = true
			s.log.Info("executor registered",
				zap.String("app", app), zap.String("id", id), zap.Strings("handlers", p.Register.Handlers))
			sendCh <- &v1.AdminMessage{
				Payload: &v1.AdminMessage_RegisterAck{RegisterAck: &v1.RegisterAck{Accepted: true}},
			}
		case *v1.ExecutorMessage_Heartbeat:
			// 心跳复用流
		case *v1.ExecutorMessage_TaskResult:
			s.handleTaskResult(p.TaskResult)
		case *v1.ExecutorMessage_LogChunk:
			if s.hub != nil {
				s.hub.Publish(p.LogChunk.LogId, p.LogChunk.LogContent) // 推送到订阅前端
			}
			s.log.Debug("log chunk", zap.Int64("log_id", p.LogChunk.LogId), zap.Int64("to_line", p.LogChunk.ToLine))
		case *v1.ExecutorMessage_Load:
			s.reg.UpdateLoad(app, id, registry.LoadReport{
				ActiveTasks: p.Load.ActiveTasks,
				CPUUsage:    p.Load.CpuUsage,
				QueueDepth:  p.Load.QueueDepth,
			})
		}
	}
}

// handleTaskResult 更新 task_run、释放阻塞槽位、失败时告警 + metrics。
func (s *ExecutorLinkService) handleTaskResult(r *v1.TaskResult) {
	var run model.TaskRun
	if err := s.db.Select("job_id", "executor_handler", "executor_address").First(&run, r.LogId).Error; err != nil {
		// task_run 查询失败（记录不存在/DB 故障）：无法关联 jobID，但仍需释放阻塞槽位，
		// 避免 BlockDiscardLater 策略下该 job 永久卡死、后续触发全被丢弃。
		s.log.Warn("query task_run failed, release slot by logID only",
			zap.Int64("log_id", r.LogId), zap.Error(err))
		if s.completer != nil {
			s.completer.ReleaseByLogID(r.LogId)
		}
		return
	}
	updates := map[string]interface{}{
		"handle_code": r.HandleCode,
		"handle_msg":  r.HandleMsg,
		"handle_time": time.UnixMilli(r.HandleTime),
	}
	if err := s.db.Model(&model.TaskRun{}).Where("id = ?", r.LogId).Updates(updates).Error; err != nil {
		s.log.Error("update task_run result", zap.Int64("log_id", r.LogId), zap.Error(err))
		return
	}
	if s.completer != nil && run.JobID != 0 {
		s.completer.OnTaskDone(run.JobID, r.LogId, int16(r.HandleCode))
	}
	if s.breaker != nil && run.ExecutorApp != "" {
		s.breaker.OnResult(run.ExecutorApp, r.HandleCode == 200)
	}
	if r.HandleCode != 200 {
		metrics.TriggerFailTotal.WithLabelValues(run.ExecutorHandler).Inc()
		if s.alerter != nil {
			go func() {
				_ = s.alerter.Notify(context.Background(), alarm.Event{
					JobID: run.JobID, LogID: r.LogId,
					HandleCode: int16(r.HandleCode), HandleMsg: r.HandleMsg,
					Address: run.ExecutorAddress,
				})
			}()
		}
	}
	s.log.Info("task result", zap.Int64("log_id", r.LogId), zap.Int32("code", r.HandleCode))
}

var _ xxlv1connect.ExecutorLinkHandler = (*ExecutorLinkService)(nil)

// AdminService 对外 RPC。
type AdminService struct {
	trig *trigger.Trigger
	db   *gorm.DB
}

// NewAdminService 创建 admin RPC 服务。
func NewAdminService(trig *trigger.Trigger, db *gorm.DB) *AdminService {
	return &AdminService{trig: trig, db: db}
}

// TriggerTask 人工/API 触发任务。
func (s *AdminService) TriggerTask(ctx context.Context, req *connect.Request[v1.TriggerTaskRequest]) (*connect.Response[v1.TriggerTaskResponse], error) {
	var job model.Job
	if err := s.db.First(&job, req.Msg.JobId).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("job %d not found", req.Msg.JobId))
	}
	if req.Msg.ExecutorParams != "" {
		job.ExecutorParam = req.Msg.ExecutorParams
	}
	s.trig.Trigger(ctx, &job)
	return connect.NewResponse(&v1.TriggerTaskResponse{
		Result: &v1.ReturnT{Code: 200, Msg: "success"},
	}), nil
}

// StreamLogs 流式拉取执行日志（gRPC server stream，P8 可接 hub）。
func (s *AdminService) StreamLogs(ctx context.Context, req *connect.Request[v1.LogRequest], stream *connect.ServerStream[v1.LogChunk]) error {
	return connect.NewError(connect.CodeUnimplemented, errNotImpl)
}

// QueryTaskStatus 查询任务运行状态。
func (s *AdminService) QueryTaskStatus(ctx context.Context, req *connect.Request[v1.StatusRequest]) (*connect.Response[v1.StatusResponse], error) {
	var run model.TaskRun
	if err := s.db.Select("handle_code").First(&run, req.Msg.LogId).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&v1.StatusResponse{Code: int32(run.HandleCode), Msg: "ok"}), nil
}

var _ xxlv1connect.AdminServiceHandler = (*AdminService)(nil)

// NewServer 装配 ConnectRPC handler。accessToken 非空时校验请求头（调度中心↔执行器认证）。
func NewServer(execLink *ExecutorLinkService, adminSvc *AdminService, accessToken string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(xxlv1connect.NewExecutorLinkHandler(execLink))
	mux.Handle(xxlv1connect.NewAdminServiceHandler(adminSvc))
	if accessToken == "" {
		return mux
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-XXL-Job-Access-Token") != accessToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
