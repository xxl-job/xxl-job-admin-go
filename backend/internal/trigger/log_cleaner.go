package trigger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/model"
)

// RunLogCleaner 周期清理过期 task_runs（仅 leader 调用）。
// retainDays<=0 表示关闭清理。每小时一次，分批删除避免长事务。
func RunLogCleaner(ctx context.Context, db *gorm.DB, retainDays int, log *zap.Logger) {
	if retainDays <= 0 {
		log.Info("log cleaner disabled (retain_days<=0)")
		return
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	clean := func() { cleanOldTaskRuns(db, time.Now(), retainDays, 1000, log) }
	clean() // 启动即跑一次
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clean()
		}
	}
}

// cleanOldTaskRuns 删除 threshold 之前的 task_runs，每批 batchSize 条。
// 抽出来便于测试注入固定时间。
func cleanOldTaskRuns(db *gorm.DB, now time.Time, retainDays, batchSize int, log *zap.Logger) {
	cutoff := now.Add(-time.Duration(retainDays) * 24 * time.Hour)
	for {
		var ids []uint64
		if err := db.Model(&model.TaskRun{}).
			Where("trigger_time < ?", cutoff).
			Limit(batchSize).
			Pluck("id", &ids).Error; err != nil {
			log.Warn("log cleaner scan", zap.Error(err))
			return
		}
		if len(ids) == 0 {
			return
		}
		if err := db.Where("id IN ?", ids).Delete(&model.TaskRun{}).Error; err != nil {
			log.Warn("log cleaner delete", zap.Error(err))
			return
		}
		log.Info("log cleaner deleted batch", zap.Int("n", len(ids)))
		if len(ids) < batchSize {
			return
		}
	}
}
