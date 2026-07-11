package trigger

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/circuit"
	v1 "xxl-job-admin-go/internal/gen/xxl/v1"
	"xxl-job-admin-go/internal/metrics"
	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
)

// timeNowNanos 返回当前 unix 纳秒（lruRouter 用）。封装便于测试。
// 注意：调度循环已在 scheduler 包运行，此处仅为 LRU 时间戳。
var timeNowNanos = func() int64 { return time.Now().UnixNano() }

// Trigger 实现调度触发：创建日志 → 阻塞检查 → 路由 → 下发 TriggerTask。
// 实现 scheduler.Trigger 接口与 rpc.TaskCompleter 接口。
type Trigger struct {
	db      *gorm.DB
	reg     registry.Registry
	routers *Routers
	block   *BlockTracker
	breaker *circuit.Breaker
	log     *zap.Logger

	retryMu    sync.Mutex
	retryCount map[uint64]int16 // jobID → 当前已重试次数（成功后清除）

	runConnMu    sync.Mutex
	runningConns map[int64]*registry.ExecutorConn // logID → 下发该任务的执行器连接（cover-early 取消用）
}

// NewTrigger 创建触发器。block 由外部传入，与 rpc 共享同一实例。
func NewTrigger(db *gorm.DB, reg registry.Registry, block *BlockTracker, log *zap.Logger) *Trigger {
	return &Trigger{
		db:           db,
		reg:          reg,
		routers:      NewRouters(),
		block:        block,
		log:          log,
		retryCount:   make(map[uint64]int16),
		runningConns: make(map[int64]*registry.ExecutorConn),
	}
}

// SetBreaker 注入熔断器（按 app 维度，路由前 Allow）。
func (t *Trigger) SetBreaker(b *circuit.Breaker) { t.breaker = b }

// Trigger 实现 scheduler.Trigger，由调度循环调用。
func (t *Trigger) Trigger(ctx context.Context, job *model.Job) {
	if err := t.trigger(ctx, job); err != nil {
		t.log.Error("trigger job", zap.Uint64("job_id", job.ID), zap.Error(err))
	}
}

// OnTaskDone 实现 rpc.TaskCompleter，由 rpc 收到 TaskResult 时调用。
// 父任务成功（handleCode=200）后触发子任务编排；失败则按 retry 策略重试。
func (t *Trigger) OnTaskDone(jobID uint64, logID int64, handleCode int16) {
	t.trackDone(jobID, logID)
	if handleCode == 200 {
		t.clearRetry(jobID)
		t.triggerChildren(jobID)
	} else {
		t.retryJob(jobID)
	}
}

// ReleaseByLogID 实现 rpc.TaskCompleter：仅知 logID 时按 logID 释放阻塞槽位。
// 用于 task_run 查询失败等异常路径，避免槽位泄漏导致 job 永久卡死。
func (t *Trigger) ReleaseByLogID(logID int64) {
	t.block.ReleaseByLogID(logID)
}

// retryJob 失败重试：按 job.executor_fail_retry_count 重试，超过则放弃。
func (t *Trigger) retryJob(jobID uint64) {
	var job model.Job
	if err := t.db.First(&job, jobID).Error; err != nil {
		return
	}
	if job.ExecutorFailRetryCount <= 0 {
		return
	}
	t.retryMu.Lock()
	count := t.retryCount[jobID]
	if count >= int16(job.ExecutorFailRetryCount) {
		delete(t.retryCount, jobID)
		t.retryMu.Unlock()
		t.log.Info("retry exhausted", zap.Uint64("job_id", jobID), zap.Int("max", job.ExecutorFailRetryCount))
		return
	}
	t.retryCount[jobID] = count + 1
	t.retryMu.Unlock()
	t.log.Info("retry job", zap.Uint64("job_id", jobID), zap.Int16("retry", count+1))
	go t.Trigger(context.Background(), &job)
}

func (t *Trigger) clearRetry(jobID uint64) {
	t.retryMu.Lock()
	delete(t.retryCount, jobID)
	t.retryMu.Unlock()
}

// triggerChildren 父任务成功后按 child_jobid（逗号分隔）触发子任务。
func (t *Trigger) triggerChildren(parentJobID uint64) {
	var parent model.Job
	if err := t.db.Select("child_jobid").First(&parent, parentJobID).Error; err != nil {
		return
	}
	if strings.TrimSpace(parent.ChildJobID) == "" {
		return
	}
	for _, idStr := range strings.Split(parent.ChildJobID, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		var child model.Job
		if err := t.db.First(&child, id).Error; err != nil {
			t.log.Warn("child job not found", zap.Uint64("child_id", id))
			continue
		}
		t.log.Info("trigger child job", zap.Uint64("parent", parentJobID), zap.Uint64("child", id))
		go t.Trigger(context.Background(), &child)
	}
}

func (t *Trigger) trigger(ctx context.Context, job *model.Job) error {
	// 1. 阻塞策略
	allow, cancelIDs := t.block.Allow(job.ID, job.ExecutorBlockStrategy)
	if !allow {
		t.log.Info("job blocked (discard later)", zap.Uint64("job_id", job.ID))
		return nil
	}
	for _, lid := range cancelIDs {
		t.log.Info("cancel running (cover early)", zap.Uint64("job_id", job.ID), zap.Int64("log_id", lid))
		t.sendCancel(job.ID, lid)
	}

	// 2. 查执行器分组
	var exec model.Executor
	if err := t.db.First(&exec, job.JobGroup).Error; err != nil {
		return fmt.Errorf("find executor: %w", err)
	}

	// 熔断检查：app 级别，open 状态拒绝调度（避免持续向故障 app 投递）
	if t.breaker != nil && !t.breaker.Allow(exec.AppName) {
		return fmt.Errorf("app %q circuit breaker open", exec.AppName)
	}

	// 3. 路由（流量泳道：job 指定 cluster 时仅候选该 cluster 的执行器，否则全部）
	var conns []*registry.ExecutorConn
	if job.Cluster != "" {
		conns = t.reg.ConnsByCluster(exec.AppName, job.Cluster)
	} else {
		conns = t.reg.Conns(exec.AppName)
	}
	if len(conns) == 0 {
		return fmt.Errorf("no online executor for app %q cluster %q", exec.AppName, job.Cluster)
	}
	router := t.routers.Get(job.ExecutorRouteStrategy)
	if router == nil {
		return fmt.Errorf("unknown route strategy %d", job.ExecutorRouteStrategy)
	}
	targets := router.Route(job, conns)
	if len(targets) == 0 {
		return fmt.Errorf("router returned no target")
	}

	// 4. 逐目标创建日志并下发 TriggerTask
	for i, conn := range targets {
		logID := t.createTaskRun(job, conn, i, len(targets))
		if logID == 0 {
			continue
		}
		t.trackStart(job.ID, logID, conn)
		t.dispatch(job, conn, logID, i, len(targets))
	}
	return nil
}

// trackStart 标记实例开始运行，并记录其执行器连接（cover-early 取消时用）。
func (t *Trigger) trackStart(jobID uint64, logID int64, conn *registry.ExecutorConn) {
	t.block.OnStart(jobID, logID)
	t.runConnMu.Lock()
	t.runningConns[logID] = conn
	t.runConnMu.Unlock()
}

// trackDone 标记实例完成，清除执行器连接记录。幂等：重复调用安全。
func (t *Trigger) trackDone(jobID uint64, logID int64) {
	t.block.OnDone(jobID, logID)
	t.runConnMu.Lock()
	delete(t.runningConns, logID)
	t.runConnMu.Unlock()
}

// sendCancel 向执行器下发 CancelTask（cover-early 阻塞策略取消旧实例）。
// 连接已断开或发送队列满时仅记录日志，不阻塞触发流程。
func (t *Trigger) sendCancel(jobID uint64, logID int64) {
	t.runConnMu.Lock()
	conn := t.runningConns[logID]
	t.runConnMu.Unlock()
	if conn == nil {
		t.log.Warn("cancel: no running conn", zap.Int64("log_id", logID))
		return
	}
	msg := &v1.AdminMessage{
		Payload: &v1.AdminMessage_Cancel{
			Cancel: &v1.CancelTask{JobId: int64(jobID), LogId: logID},
		},
	}
	select {
	case conn.Send <- msg:
	default:
		t.log.Warn("cancel: send queue full", zap.Int64("log_id", logID))
	}
}

// dispatch 通过 conn.Send 下发 TriggerTask。发送队列满则放弃并回滚 OnStart。
func (t *Trigger) dispatch(job *model.Job, conn *registry.ExecutorConn, logID int64, shardIndex, shardTotal int) {
	msg := &v1.AdminMessage{
		Payload: &v1.AdminMessage_Trigger{
			Trigger: &v1.TriggerTask{
				JobId:           int64(job.ID),
				ExecutorHandler: job.ExecutorHandler,
				ExecutorParams:  job.ExecutorParam,
				ExecutorTimeout: int32(job.ExecutorTimeout),
				LogId:           logID,
				LogDateTime:     time.Now().UnixMilli(),
				ShardIndex:      int32(shardIndex),
				ShardTotal:      int32(shardTotal),
				GlueType:        int32(job.GlueType),
				GlueSource:      job.GlueSource,
			},
		},
	}
	select {
	case conn.Send <- msg:
	default:
		t.log.Warn("executor send queue full", zap.Uint64("job_id", job.ID), zap.String("addr", conn.Address))
		t.trackDone(job.ID, logID)
	}
}

// createTaskRun 创建执行记录，返回 logID。含触发幂等：抢占 (job_id, fire_second, executor_address)，
// 唯一冲突即代表 leader 切换窗口已对该执行器触发，跳过避免重复执行。
func (t *Trigger) createTaskRun(job *model.Job, conn *registry.ExecutorConn, shardIndex, shardTotal int) int64 {
	fireSec := time.Now().Unix()
	if err := t.db.Create(&model.JobTriggerDedup{
		JobID: job.ID, FireSecond: fireSec, ExecutorAddress: conn.Address,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "Duplicate") {
			t.log.Info("trigger dedup skip (already fired this second)",
				zap.Uint64("job_id", job.ID), zap.String("addr", conn.Address))
		} else {
			t.log.Error("trigger dedup insert", zap.Uint64("job_id", job.ID), zap.Error(err))
		}
		return 0
	}
	run := &model.TaskRun{
		JobID:           job.ID,
		ExecutorAddress: conn.Address,
		ExecutorApp:     conn.AppName,
		ExecutorHandler: job.ExecutorHandler,
		ExecutorParam:   job.ExecutorParam,
		TriggerTime:     time.Now(),
		TriggerCode:     200,
	}
	if shardTotal > 1 {
		run.ExecutorShardingParam = fmt.Sprintf("%d/%d", shardIndex+1, shardTotal)
	}
	if err := t.db.Create(run).Error; err != nil {
		t.log.Error("create task_run", zap.Uint64("job_id", job.ID), zap.Error(err))
		return 0
	}
	metrics.TriggerTotal.WithLabelValues(job.ExecutorHandler).Inc()
	return int64(run.ID)
}

// RunDedupCleaner 周期清理过期的触发去重记录（保留最近 10 分钟），仅 leader 调用。
func RunDedupCleaner(ctx context.Context, db *gorm.DB, log *zap.Logger) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	clean := func() {
		cutoff := time.Now().Add(-10 * time.Minute)
		if err := db.Where("created_at < ?", cutoff).Delete(&model.JobTriggerDedup{}).Error; err != nil {
			log.Warn("clean trigger dedup", zap.Error(err))
		}
	}
	clean()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clean()
		}
	}
}
