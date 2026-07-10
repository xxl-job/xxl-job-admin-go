package scheduler

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/metrics"
	"xxl-job-admin-go/internal/model"
)

// Trigger 触发器接口：将任务下发给执行器。
// P2 用 LogTrigger 占位；P3 替换为通过 gRPC 调用执行器。
type Trigger interface {
	Trigger(ctx context.Context, job *model.Job)
}

// LogTrigger P2 占位触发器，仅记录日志。
type LogTrigger struct {
	Log *zap.Logger
}

// Trigger 记录触发日志。
func (t *LogTrigger) Trigger(ctx context.Context, job *model.Job) {
	t.Log.Info("trigger job (P2 stub)",
		zap.Uint64("job_id", job.ID),
		zap.String("handler", job.ExecutorHandler),
		zap.Int64("next_ms", job.TriggerNextTime))
}

// scheduledEntry 时间轮中的待触发条目。
type scheduledEntry struct {
	job            *model.Job
	expectedNextMs int64
}

// Scheduler 调度核心：5s 预读 jobs 入内存时间轮 + 每秒触发到点任务。
//
// 对标 Java 版 JobScheduleHelper：
//   - preload（5s 预读）：减少 DB 查询频率，任务先入内存时间轮
//   - fireSecond（每秒）：触发当前秒到点任务，更新 next_time
type Scheduler struct {
	db      *gorm.DB
	pool    *Pool
	trigger Trigger
	log     *zap.Logger

	mu       sync.Mutex
	ring     map[int64][]*scheduledEntry // 触发秒(unix) → entries
	inFlight map[uint64]bool             // jobID 是否在时间轮中（去重）
}

// New 创建调度器。
func New(db *gorm.DB, pool *Pool, trigger Trigger, log *zap.Logger) *Scheduler {
	return &Scheduler{
		db:       db,
		pool:     pool,
		trigger:  trigger,
		log:      log,
		ring:     make(map[int64][]*scheduledEntry),
		inFlight: make(map[uint64]bool),
	}
}

// Run 启动调度循环（预读 + 触发）。阻塞直到 ctx 取消。
func (s *Scheduler) Run(ctx context.Context) {
	s.log.Info("scheduler started")
	s.preload(ctx) // 立即预读一次

	preloadTicker := time.NewTicker(5 * time.Second)
	defer preloadTicker.Stop()
	fireTicker := time.NewTicker(1 * time.Second)
	defer fireTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return
		case <-preloadTicker.C:
			s.preload(ctx)
		case now := <-fireTicker.C:
			s.fireSecond(ctx, now.Unix())
		}
	}
}

// preload 预读未来 5s 内待触发任务入时间轮。
// 配置热更新：每次预读都用最新 DB 数据刷新时间轮快照——改 CRON/频率后最多 5s 生效，无需 stop/start。
func (s *Scheduler) preload(ctx context.Context) {
	horizon := time.Now().Add(5 * time.Second).UnixMilli()
	var jobs []model.Job
	err := s.db.WithContext(ctx).
		Where("trigger_status = 1 AND trigger_next_time > 0 AND trigger_next_time <= ?", horizon).
		Find(&jobs).Error
	if err != nil {
		s.log.Error("preload jobs", zap.Error(err))
		return
	}
	nowSec := time.Now().Unix()
	s.mu.Lock()
	for i := range jobs {
		j := jobs[i]
		sec := j.TriggerNextTime / 1000
		if sec < nowSec {
			sec = nowSec // 过期任务放当前秒立即触发（补偿）
		}
		// 移除该 job 的旧快照（配置/next 可能已变），用最新 DB 数据重新入轮
		s.removeEntryLocked(j.ID)
		s.ring[sec] = append(s.ring[sec], &scheduledEntry{job: &jobs[i], expectedNextMs: j.TriggerNextTime})
		s.inFlight[j.ID] = true
	}
	s.mu.Unlock()

	// 更新运行中任务数指标（已启动 trigger_status=1 的任务总数）
	var running int64
	if err := s.db.WithContext(ctx).Model(&model.Job{}).Where("trigger_status = 1").Count(&running).Error; err == nil {
		metrics.JobRunning.Set(float64(running))
	}
}

// removeEntryLocked 从时间轮移除指定 job 的旧快照（调用方持 s.mu）。
func (s *Scheduler) removeEntryLocked(jobID uint64) {
	for sec, entries := range s.ring {
		for i, e := range entries {
			if e.job.ID == jobID {
				s.ring[sec] = append(entries[:i], entries[i+1:]...)
				if len(s.ring[sec]) == 0 {
					delete(s.ring, sec)
				}
				return
			}
		}
	}
}

// fireSecond 触发当前秒及之前所有过期任务，并更新下次触发时间。
func (s *Scheduler) fireSecond(ctx context.Context, sec int64) {
	s.mu.Lock()
	var entries []*scheduledEntry
	for k, e := range s.ring {
		if k <= sec { // 当前秒及之前（过期补偿）
			entries = append(entries, e...)
			delete(s.ring, k)
		}
	}
	s.mu.Unlock()

	for _, e := range entries {
		s.mu.Lock()
		delete(s.inFlight, e.job.ID)
		s.mu.Unlock()

		job := e.job
		s.pool.Submit(job.ID, func() {
			start := time.Now()
			s.trigger.Trigger(ctx, job)
			if d := time.Since(start); d > s.pool.slowThreshold {
				s.pool.MarkSlow(job.ID)
				s.log.Warn("job slow, demoted to slow pool",
					zap.Uint64("job_id", job.ID), zap.Duration("dur", d))
			}
		}, 0)

		// 计算并更新下次触发时间（基于当前时间，避免过期 job 反复触发）
		next, err := NextTriggerTime(job.ScheduleType, job.ScheduleConf, time.Now())
		if err != nil {
			s.log.Error("compute next trigger", zap.Uint64("job_id", job.ID), zap.Error(err))
			continue
		}
		if err := s.db.Model(&model.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"trigger_last_time": job.TriggerNextTime,
			"trigger_next_time": next,
		}).Error; err != nil {
			s.log.Error("update next trigger", zap.Uint64("job_id", job.ID), zap.Error(err))
			continue
		}
		job.TriggerNextTime = next

		// 若下次触发在预读窗口内，重新入时间轮
		nextSec := next / 1000
		if nextSec <= time.Now().Add(5*time.Second).Unix() {
			s.mu.Lock()
			if !s.inFlight[job.ID] {
				s.inFlight[job.ID] = true
				s.ring[nextSec] = append(s.ring[nextSec], &scheduledEntry{job: job, expectedNextMs: next})
			}
			s.mu.Unlock()
		}
	}
}
