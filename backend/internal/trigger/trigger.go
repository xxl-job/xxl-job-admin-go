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
	"xxl-job-admin-go/internal/limiter"
	"xxl-job-admin-go/internal/metrics"
	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
	v1 "xxl-job-admin-go/pkg/gen/xxl/v1"
)

// timeNowNanos 返回当前 unix 纳秒（lruRouter 用）。封装便于测试。
// 注意：调度循环已在 scheduler 包运行，此处仅为 LRU 时间戳。
var timeNowNanos = func() int64 { return time.Now().UnixNano() }

// Trigger 实现调度触发：创建日志 → 阻塞检查 → 路由 → 下发 TriggerTask。
// 实现 scheduler.Trigger 接口与 rpc.TaskCompleter 接口。
type Trigger struct {
	db          *gorm.DB
	reg         registry.Registry
	routers     *Routers
	block       *BlockTracker
	breaker     *circuit.Breaker
	rateLimiter *limiter.Limiter
	log         *zap.Logger

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

// SetLimiter 注入限流器（按 app 维度，路由前 Allow）。
func (t *Trigger) SetLimiter(l *limiter.Limiter) { t.rateLimiter = l }

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
	count, canRetry := t.recordRetry(jobID, int16(job.ExecutorFailRetryCount))
	if !canRetry {
		t.log.Info("retry exhausted", zap.Uint64("job_id", jobID), zap.Int("max", job.ExecutorFailRetryCount))
		return
	}
	t.log.Info("retry job", zap.Uint64("job_id", jobID), zap.Int16("retry", count))
	go t.Trigger(context.Background(), &job)
}

// recordRetry 记录一次重试计数（纯计数逻辑，无异步副作用，便于测试）。
// 返回 (本次更新后的计数, 是否仍可重试)；已达上限时清除计数并返回 (0, false)。
func (t *Trigger) recordRetry(jobID uint64, max int16) (int16, bool) {
	t.retryMu.Lock()
	defer t.retryMu.Unlock()
	if t.retryCount[jobID] >= max {
		delete(t.retryCount, jobID)
		return 0, false
	}
	t.retryCount[jobID]++
	return t.retryCount[jobID], true
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
		t.recordTriggerFail(job, "", fmt.Sprintf("执行器分组 %d 未找到: %v", job.JobGroup, err))
		return fmt.Errorf("find executor: %w", err)
	}

	// 熔断检查：app 级别，open 状态拒绝调度（避免持续向故障 app 投递）
	if t.breaker != nil && !t.breaker.Allow(exec.AppName) {
		t.recordTriggerFail(job, "", fmt.Sprintf("app %q 熔断中", exec.AppName))
		return fmt.Errorf("app %q circuit breaker open", exec.AppName)
	}
	if t.rateLimiter != nil && !t.rateLimiter.Allow(exec.AppName) {
		t.recordTriggerFail(job, "", fmt.Sprintf("app %q 被限流", exec.AppName))
		return fmt.Errorf("app %q rate limited", exec.AppName)
	}

	// 3. 路由（流量泳道：job 指定 cluster 时仅候选该 cluster 的执行器，否则全部）
	var conns []*registry.ExecutorConn
	if job.Cluster != "" {
		conns = t.reg.ConnsByCluster(exec.AppName, job.Cluster)
	} else {
		conns = t.reg.Conns(exec.AppName)
	}
	if len(conns) == 0 {
		t.recordTriggerFail(job, "", fmt.Sprintf("app %q cluster %q 无在线执行器", exec.AppName, job.Cluster))
		return fmt.Errorf("no online executor for app %q cluster %q", exec.AppName, job.Cluster)
	}
	router := t.routers.Get(job.ExecutorRouteStrategy)
	if router == nil {
		t.recordTriggerFail(job, "", fmt.Sprintf("未知路由策略 %d", job.ExecutorRouteStrategy))
		return fmt.Errorf("unknown route strategy %d", job.ExecutorRouteStrategy)
	}
	targets := router.Route(job, conns)
	if len(targets) == 0 {
		t.recordTriggerFail(job, "", "路由器未返回目标")
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
	if !t.sendCancelByLogID(jobID, logID) {
		t.log.Warn("cancel: no running conn", zap.Int64("log_id", logID))
		return
	}
}

// sendCancelByLogID 按 logID 取运行中连接并下发 CancelTask。
// 命中且成功入队返回 true；无运行记录或队列满返回 false。供 sendCancel 与 CancelRun 复用。
// jobID 在公开 CancelRun 路径传 0（仅按 logID 取消），内部 cover-early 路径传真实 jobID。
func (t *Trigger) sendCancelByLogID(jobID uint64, logID int64) bool {
	t.runConnMu.Lock()
	conn := t.runningConns[logID]
	t.runConnMu.Unlock()
	if conn == nil {
		return false
	}
	msg := &v1.AdminMessage{
		Payload: &v1.AdminMessage_Cancel{
			Cancel: &v1.CancelTask{JobId: int64(jobID), LogId: logID},
		},
	}
	select {
	case conn.Send <- msg:
		return true
	default:
		t.log.Warn("cancel: send queue full", zap.Int64("log_id", logID))
		return false
	}
}

// CancelRun 按 logID 终止运行中任务。命中且成功入队返回 true；无运行记录返回 false。
func (t *Trigger) CancelRun(logID int64) bool {
	return t.sendCancelByLogID(0, logID)
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
		t.markTriggerFail(logID, "执行器发送队列已满")
		t.trackDone(job.ID, logID)
	}
}

// markTriggerFail 将已创建的 TaskRun(logID) 标记为触发失败（trigger_code=500）。
// 用于 dispatch 发送队列满等"已建日志但下发失败"的场景，原地更新避免双行。
func (t *Trigger) markTriggerFail(logID int64, reason string) {
	if logID == 0 {
		return
	}
	if err := t.db.Model(&model.TaskRun{}).Where("id = ?", logID).Updates(map[string]any{
		"trigger_code": int16(500),
		"trigger_msg":  fmt.Sprintf("触发失败: %s", reason),
	}).Error; err != nil {
		t.log.Error("mark trigger fail", zap.Int64("log_id", logID), zap.Error(err))
	}
}

// recordTriggerFail 触发失败补写一条 TaskRun（trigger_code=500），便于日志页回溯"为什么没跑"。
// 不写 JobTriggerDedup（失败/异常路径不应被 fire_second 幂等拦截）。
func (t *Trigger) recordTriggerFail(job *model.Job, addr, reason string) {
	run := &model.TaskRun{
		JobID:           job.ID,
		ExecutorAddress: addr,
		ExecutorApp:     "", // 失败前可能未确定 app，留空
		ExecutorHandler: job.ExecutorHandler,
		ExecutorParam:   job.ExecutorParam,
		TriggerTime:     time.Now(),
		TriggerCode:     500,
		TriggerMsg:      fmt.Sprintf("触发失败: %s", reason),
	}
	if err := t.db.Create(run).Error; err != nil {
		t.log.Error("record trigger fail", zap.Uint64("job_id", job.ID), zap.Error(err))
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
		TriggerMsg:      fmt.Sprintf("路由策略=%d, 目标机器=%s, 触发结果=成功", job.ExecutorRouteStrategy, conn.Address),
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

// TriggerOnce 手动触发一次：paramOverride 非空则覆盖本次参数（不落库）；
// addrOverride 非空则按地址精确匹配在线连接、绕过路由。
// 跳过阻塞策略与触发幂等（手动一次性触发）。
func (t *Trigger) TriggerOnce(ctx context.Context, job *model.Job, paramOverride, addrOverride string) error {
	var exec model.Executor
	if err := t.db.First(&exec, job.JobGroup).Error; err != nil {
		t.recordTriggerFail(job, "", fmt.Sprintf("执行器分组 %d 未找到: %v", job.JobGroup, err))
		return fmt.Errorf("find executor: %w", err)
	}

	param := job.ExecutorParam
	if paramOverride != "" {
		param = paramOverride
	}

	var conn *registry.ExecutorConn
	if addrOverride != "" {
		for _, c := range t.reg.Conns(exec.AppName) {
			if c.Address == addrOverride {
				conn = c
				break
			}
		}
		if conn == nil {
			t.recordTriggerFail(job, addrOverride, fmt.Sprintf("指定地址 %s 不在 app %q 在线列表", addrOverride, exec.AppName))
			return fmt.Errorf("address %s not online in app %q", addrOverride, exec.AppName)
		}
	} else {
		conns := t.reg.Conns(exec.AppName)
		if job.Cluster != "" {
			conns = t.reg.ConnsByCluster(exec.AppName, job.Cluster)
		}
		if len(conns) == 0 {
			t.recordTriggerFail(job, "", fmt.Sprintf("app %q cluster %q 无在线执行器", exec.AppName, job.Cluster))
			return fmt.Errorf("no online executor for app %q cluster %q", exec.AppName, job.Cluster)
		}
		router := t.routers.Get(job.ExecutorRouteStrategy)
		if router == nil {
			// 写失败日志，避免路由策略非法时静默失败（无日志可查）
			t.recordTriggerFail(job, "", fmt.Sprintf("未知路由策略 %d", job.ExecutorRouteStrategy))
			return fmt.Errorf("unknown route strategy %d", job.ExecutorRouteStrategy)
		}
		targets := router.Route(job, conns)
		if len(targets) == 0 {
			return fmt.Errorf("router returned no target")
		}
		conn = targets[0]
	}

	// 临时参数构造一份 job 副本下发，不污染原对象
	shadow := *job
	shadow.ExecutorParam = param
	logID := t.createTaskRunOnce(&shadow, conn)
	if logID == 0 {
		return fmt.Errorf("create task_run failed")
	}
	t.trackStart(job.ID, logID, conn)
	t.dispatch(&shadow, conn, logID, 0, 1)
	return nil
}

// createTaskRunOnce 手动触发用：不写 JobTriggerDedup（手动触发不应被 fire_second 拦截）。
func (t *Trigger) createTaskRunOnce(job *model.Job, conn *registry.ExecutorConn) int64 {
	run := &model.TaskRun{
		JobID:           job.ID,
		ExecutorAddress: conn.Address,
		ExecutorApp:     conn.AppName,
		ExecutorHandler: job.ExecutorHandler,
		ExecutorParam:   job.ExecutorParam,
		TriggerTime:     time.Now(),
		TriggerCode:     200,
		TriggerMsg:      fmt.Sprintf("手动触发, 路由=%d, 目标机器=%s", job.ExecutorRouteStrategy, conn.Address),
	}
	if err := t.db.Create(run).Error; err != nil {
		t.log.Error("create task_run (once)", zap.Uint64("job_id", job.ID), zap.Error(err))
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
