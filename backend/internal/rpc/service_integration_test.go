//go:build integration

// 真集成测试：testcontainers MySQL 验 handleTaskResult 在真 DB 下的行为
// （更新 task_run / 释放槽位 / 失败告警 / 记录不存在时 ReleaseByLogID 防卡死）。
// `go test -tags=integration ./internal/rpc/` 触发，无 Docker 自动跳过。

package rpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"go.uber.org/zap"

	"xxl-job-admin-go/internal/alarm"
	v1 "xxl-job-admin-go/internal/gen/xxl/v1"
	"xxl-job-admin-go/internal/model"
)

func setupMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()
	container, err := mysql.Run(ctx, "mysql:8.0")
	if err != nil {
		t.Skipf("跳过：无法启动 MySQL 容器: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	dsn, cerr := container.ConnectionString(ctx)
	if cerr != nil {
		t.Fatalf("connection string: %v", cerr)
	}
	dsn += "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	if err := db.AutoMigrate(&model.Job{}, &model.TaskRun{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

// mockCompleter 记录 OnTaskDone / ReleaseByLogID 调用。
type mockCompleter struct {
	mu        sync.Mutex
	doneCalls []uint64
	releases  []int64
}

func (m *mockCompleter) OnTaskDone(jobID uint64, logID int64, code int16) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.doneCalls = append(m.doneCalls, jobID)
}
func (m *mockCompleter) ReleaseByLogID(logID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releases = append(m.releases, logID)
}

// mockAlerter 记录告警事件。notifyCh 用于同步等待异步告警 goroutine（handleTaskResult 中 go 调用）。
type mockAlerter struct {
	mu       sync.Mutex
	notified []alarm.Event
	notifyCh chan struct{}
}

func newMockAlerter() *mockAlerter { return &mockAlerter{notifyCh: make(chan struct{}, 1)} }

func (m *mockAlerter) Notify(_ context.Context, e alarm.Event) error {
	m.mu.Lock()
	m.notified = append(m.notified, e)
	m.mu.Unlock()
	if m.notifyCh != nil {
		select { case m.notifyCh <- struct{}{}: default: }
	}
	return nil
}

// TestHandleTaskResult_Success 成功回报：更新 task_run + OnTaskDone，不告警。
func TestHandleTaskResult_Success(t *testing.T) {
	db := setupMySQL(t)
	job := model.Job{JobDesc: "t", ExecutorHandler: "demoJob"}
	db.Create(&job)
	run := model.TaskRun{JobID: job.ID, ExecutorHandler: "demoJob", ExecutorAddress: "10.0.0.1:9999", TriggerTime: time.Now()}
	db.Create(&run)

	mc := &mockCompleter{}
	ma := &mockAlerter{}
	svc := NewExecutorLinkService(nil, db, mc, ma, nil, zap.NewNop())

	svc.handleTaskResult(&v1.TaskResult{LogId: int64(run.ID), HandleCode: 200, HandleMsg: "ok", HandleTime: time.Now().UnixMilli()})

	var updated model.TaskRun
	db.First(&updated, run.ID)
	if updated.HandleCode != 200 {
		t.Errorf("handle_code 未更新为 200，got %d", updated.HandleCode)
	}
	if len(mc.doneCalls) != 1 || mc.doneCalls[0] != job.ID {
		t.Errorf("OnTaskDone 应调用 1 次 (jobID=%d)", job.ID)
	}
	if len(ma.notified) != 0 {
		t.Error("成功不应触发告警")
	}
}

// TestHandleTaskResult_Failure 失败回报：更新 + OnTaskDone + 告警。
func TestHandleTaskResult_Failure(t *testing.T) {
	db := setupMySQL(t)
	job := model.Job{JobDesc: "t", ExecutorHandler: "failJob"}
	db.Create(&job)
	run := model.TaskRun{JobID: job.ID, ExecutorHandler: "failJob", ExecutorAddress: "10.0.0.2:9999", TriggerTime: time.Now()}
	db.Create(&run)

	mc := &mockCompleter{}
	ma := newMockAlerter()
	svc := NewExecutorLinkService(nil, db, mc, ma, nil, zap.NewNop())

	svc.handleTaskResult(&v1.TaskResult{LogId: int64(run.ID), HandleCode: 500, HandleMsg: "boom", HandleTime: time.Now().UnixMilli()})
	<-ma.notifyCh // 等异步告警 goroutine 完成

	if len(mc.doneCalls) != 1 {
		t.Error("失败也应 OnTaskDone（释放槽位）")
	}
	if len(ma.notified) != 1 {
		t.Error("失败应触发告警")
	}
}

// TestHandleTaskResult_NotFound task_run 不存在：ReleaseByLogID 防卡死，不 panic。
func TestHandleTaskResult_NotFound(t *testing.T) {
	db := setupMySQL(t)
	mc := &mockCompleter{}
	ma := &mockAlerter{}
	svc := NewExecutorLinkService(nil, db, mc, ma, nil, zap.NewNop())

	svc.handleTaskResult(&v1.TaskResult{LogId: 999999, HandleCode: 200, HandleTime: time.Now().UnixMilli()})

	if len(mc.releases) != 1 || mc.releases[0] != 999999 {
		t.Error("记录不存在时应 ReleaseByLogID 防卡死")
	}
	if len(mc.doneCalls) != 0 {
		t.Error("记录不存在不应 OnTaskDone")
	}
}
