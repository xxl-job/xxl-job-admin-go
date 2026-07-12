package trigger

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"xxl-job-admin-go/internal/model"
)

// TestCleanOldTaskRuns 验证：超过保留天数的记录被分批删除，保留期内的不动。
// 不启动定时器，直接调用内部 cleanOldTaskRuns（注入固定 now 与 retainDays）。
func TestCleanOldTaskRuns(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	old := now.Add(-31 * 24 * time.Hour) // 31 天前，超过保留 30 天
	fresh := now.Add(-10 * 24 * time.Hour)

	db.Create(&model.TaskRun{JobID: 1, TriggerTime: old, ExecutorHandler: "h"})
	db.Create(&model.TaskRun{JobID: 1, TriggerTime: fresh, ExecutorHandler: "h"})

	cleanOldTaskRuns(db, now, 30, 1000, zap.NewNop())

	var n int64
	db.Model(&model.TaskRun{}).Count(&n)
	if n != 1 {
		t.Fatalf("应保留 1 条（10 天前），got %d", n)
	}
	// 剩余应是 fresh 那条
	var remain model.TaskRun
	db.First(&remain)
	if !remain.TriggerTime.Equal(fresh) {
		t.Fatalf("删错了，剩余 %v", remain.TriggerTime)
	}
	_ = context.Background
}
