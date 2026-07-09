package scheduler

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	partitionRetentionDays = 30 // 日志保留天数
	partitionAheadDays     = 7  // 预建未来分区天数
)

// partitionExists 查询分区是否存在。
func partitionExists(db *gorm.DB, name string) (bool, error) {
	var cnt int64
	err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.PARTITIONS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'task_runs' AND PARTITION_NAME = ?",
		name,
	).Scan(&cnt).Error
	return cnt > 0, err
}

// ManagePartitions 维护 task_runs 按天分区：
//   - REORGANIZE p_max 预建未来 7 天分区（含 p_max 占位）
//   - DROP 超过 30 天的老分区
//
// 仅对分区表有效；非分区表（AutoMigrate 建的）直接跳过。
func ManagePartitions(ctx context.Context, db *gorm.DB, log *zap.Logger) {
	// 检查 task_runs 是否为分区表（非分区表在 information_schema 也有 1 条 PARTITION_NAME=NULL 记录，需排除）
	var partCnt int64
	if err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.PARTITIONS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'task_runs' AND PARTITION_NAME IS NOT NULL",
	).Scan(&partCnt).Error; err != nil || partCnt == 0 {
		return // 非分区表，跳过
	}

	now := time.Now()

	// 1. 预建未来分区（REORGANIZE p_max，因 p_max 为 MAXVALUE 末尾分区）
	for i := 0; i <= partitionAheadDays; i++ {
		day := now.AddDate(0, 0, i)
		name := "p" + day.Format("20060102")
		exists, err := partitionExists(db, name)
		if err != nil || exists {
			continue
		}
		next := day.AddDate(0, 0, 1).Format("2006-01-02")
		sql := fmt.Sprintf(
			"ALTER TABLE task_runs REORGANIZE PARTITION p_max INTO (PARTITION %s VALUES LESS THAN (TO_DAYS('%s')), PARTITION p_max VALUES LESS THAN MAXVALUE)",
			name, next)
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			log.Debug("reorganize p_max (maybe non-partitioned table)", zap.String("part", name), zap.Error(err))
		} else {
			log.Info("added partition", zap.String("part", name))
		}
	}

	// 2. 删除老分区（保留 30 天）
	cutoff := now.AddDate(0, 0, -partitionRetentionDays)
	for i := 0; i < 7; i++ {
		day := cutoff.AddDate(0, 0, -i)
		name := "p" + day.Format("20060102")
		exists, err := partitionExists(db, name)
		if err != nil || !exists {
			continue
		}
		sql := fmt.Sprintf("ALTER TABLE task_runs DROP PARTITION %s", name)
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			log.Debug("drop partition", zap.String("part", name), zap.Error(err))
		} else {
			log.Info("dropped partition", zap.String("part", name))
		}
	}
}

// RunPartitionMaintainer 定时维护（启动一次 + 每 24 小时一次）。仅 leader 调用。
func RunPartitionMaintainer(ctx context.Context, db *gorm.DB, log *zap.Logger) {
	ManagePartitions(ctx, db, log)
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			ManagePartitions(ctx, db, log)
		}
	}
}
