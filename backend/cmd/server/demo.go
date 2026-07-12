package main

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/scheduler"
	"xxl-job-admin-go/pkg/executor"
)

// demoSeed 插入示例执行器分组与任务（幂等，已存在则跳过）。
// 仅在 --demo 模式调用，不污染生产数据。
func demoSeed(db *gorm.DB) error {
	var exec model.Executor
	err := db.Where("app_name = ?", "demo-app").First(&exec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		exec = model.Executor{AppName: "demo-app", Title: "示例执行器", AddressType: 0}
		if err := db.Create(&exec).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 已有示例任务则跳过
	var cnt int64
	if db.Model(&model.Job{}).Where("job_group = ?", exec.ID).Count(&cnt).Error == nil && cnt > 0 {
		return nil
	}

	jobs := []model.Job{
		{JobGroup: exec.ID, JobDesc: "示例·CRON每5秒", ScheduleType: 1, ScheduleConf: "0/5 * * * * ?", ExecutorHandler: "demoJob", ExecutorParam: "hello-cron", ExecutorRouteStrategy: 1, TriggerStatus: 1, CreatorID: 1},
		{JobGroup: exec.ID, JobDesc: "示例·固定频率30s", ScheduleType: 2, ScheduleConf: "30", ExecutorHandler: "demoJob", ExecutorParam: "hello-fixed", ExecutorRouteStrategy: 3, TriggerStatus: 1, CreatorID: 1},
		{JobGroup: exec.ID, JobDesc: "示例·慢任务(停止)", ScheduleType: 2, ScheduleConf: "60", ExecutorHandler: "slowJob", ExecutorParam: "slow-demo", ExecutorRouteStrategy: 1, TriggerStatus: 0, CreatorID: 1},
	}
	for i := range jobs {
		if next, err := scheduler.NextTriggerTime(jobs[i].ScheduleType, jobs[i].ScheduleConf, time.Now()); err == nil {
			jobs[i].TriggerNextTime = next
		}
		if err := db.Create(&jobs[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// runDemoExecutor 内嵌示例执行器：连本机 gRPC (:9090)，注册 demoJob/slowJob。
// 断线自动重连，gRPC server 未就绪时重试。阻塞直到 ctx 取消。
func runDemoExecutor(ctx context.Context, logger *zap.Logger) {
	ex := executor.New("http://localhost:9090", "demo-app", "", logger)
	ex.WithHTTPTask(true) // demo 开启 HTTP 任务能力，便于开箱验证 glue_type=4
	ex.Register("demoJob", func(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
		jc.Log("demoJob 开始执行, param=" + param)
		time.Sleep(200 * time.Millisecond)
		jc.Log("demoJob 执行完成")
		return 200, "success: " + param
	})
	ex.Register("slowJob", func(ctx context.Context, jc *executor.JobContext, param string) (int, string) {
		jc.Log("slowJob 开始, 预计 2s")
		time.Sleep(2 * time.Second)
		jc.Log("slowJob 完成")
		return 200, "done"
	})
	logger.Info("demo executor starting (connects to localhost:9090, app=demo-app)")
	if err := ex.Run(ctx); err != nil && ctx.Err() == nil {
		logger.Error("demo executor stopped", zap.Error(err))
	}
}
