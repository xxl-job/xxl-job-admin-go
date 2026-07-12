package trigger

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
)

// makeSendCh 测试用：带缓冲发送队列。
func makeSendCh() chan<- *v1.AdminMessage {
	return make(chan *v1.AdminMessage, 8)
}

// newTestDBOnce 在 newTestDB 基础上额外 migrate Executor（TriggerOnce 内 t.db.First(&exec, jobGroup) 需要）。
func newTestDBOnce(t *testing.T) *gorm.DB {
	db := newTestDB(t)
	if err := db.AutoMigrate(&model.Executor{}); err != nil {
		t.Fatalf("migrate executor: %v", err)
	}
	return db
}

// TestTriggerOnceParamOverrideAndAddr 验证：临时参数覆盖、指定地址绕过路由命中、跳过 dedup。
func TestTriggerOnceParamOverrideAndAddr(t *testing.T) {
	db := newTestDBOnce(t)
	reg := registry.NewMemory()
	reg.OnConnect("app", "e1", &registry.ExecutorConn{ID: "e1", AppName: "app", Address: "10.0.0.1:9999", Send: makeSendCh()})
	trig := NewTrigger(db, reg, NewBlockTracker(), zap.NewNop())

	job := &model.Job{ID: 3, JobGroup: 0, ExecutorHandler: "h", ExecutorParam: "default", ExecutorRouteStrategy: 1}

	// 需 DB 里有对应 executor 分组（trigger 内 t.db.First(&exec, job.JobGroup)）
	db.Create(&model.Executor{ID: 1, AppName: "app"})
	job.JobGroup = 1

	// 指定地址 + 临时参数
	if err := trig.TriggerOnce(context.Background(), job, "OVERRIDE", "10.0.0.1:9999"); err != nil {
		t.Fatalf("TriggerOnce: %v", err)
	}

	var run model.TaskRun
	db.First(&run)
	if run.ExecutorParam != "OVERRIDE" {
		t.Fatalf("临时参数应覆盖为 OVERRIDE，got %q", run.ExecutorParam)
	}
	if run.ExecutorAddress != "10.0.0.1:9999" {
		t.Fatalf("应命中指定地址，got %q", run.ExecutorAddress)
	}
}

// TestTriggerOnceAddrNotFound 验证指定地址不在线时返回错误并补写失败记录。
func TestTriggerOnceAddrNotFound(t *testing.T) {
	db := newTestDBOnce(t)
	reg := registry.NewMemory()
	trig := NewTrigger(db, reg, NewBlockTracker(), zap.NewNop())
	db.Create(&model.Executor{ID: 1, AppName: "app"})
	job := &model.Job{ID: 3, JobGroup: 1, ExecutorHandler: "h", ExecutorParam: "p"}

	err := trig.TriggerOnce(context.Background(), job, "", "9.9.9.9:1")
	if err == nil {
		t.Fatal("地址不在线应返回错误")
	}
	var n int64
	db.Model(&model.TaskRun{}).Count(&n)
	if n != 1 {
		t.Fatalf("应补写 1 条失败记录，got %d", n)
	}
}
