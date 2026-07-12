package trigger

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
)

// newTestDB 构造 sqlite 内存库并 AutoMigrate TaskRun。
func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.TaskRun{}, &model.JobTriggerDedup{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// TestRecordTriggerFailWritesTaskRun 验证失败路径补写 trigger_code=500 的 TaskRun，
// 不插 dedup 行。
func TestRecordTriggerFailWritesTaskRun(t *testing.T) {
	db := newTestDB(t)
	trig := NewTrigger(db, registry.NewMemory(), NewBlockTracker(), zap.NewNop())

	job := &model.Job{ID: 7, ExecutorHandler: "h", ExecutorParam: "p"}
	trig.recordTriggerFail(job, "1.2.3.4:9999", "no online executor")

	var run model.TaskRun
	if err := db.First(&run).Error; err != nil {
		t.Fatalf("应补写一条 TaskRun: %v", err)
	}
	if run.TriggerCode != 500 || run.ExecutorHandler != "h" || run.ExecutorParam != "p" {
		t.Fatalf("字段错误: %+v", run)
	}
	if run.ExecutorAddress != "1.2.3.4:9999" || run.TriggerMsg == "" {
		t.Fatalf("地址/备注缺失: %+v", run)
	}
	var dedup int64
	db.Model(&model.JobTriggerDedup{}).Count(&dedup)
	if dedup != 0 {
		t.Fatalf("失败记录不应写 dedup，got %d", dedup)
	}
	_ = time.Now
}
