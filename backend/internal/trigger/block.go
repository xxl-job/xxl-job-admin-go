package trigger

import "sync"

// 阻塞策略常量（对标 xxl-job ExecutorBlockStrategyEnum）
const (
	BlockSerial       int8 = 0 // 单机串行（默认）
	BlockDiscardLater int8 = 1 // 丢弃后续
	BlockCoverEarly   int8 = 2 // 覆盖之前
)

// BlockTracker 跟踪每个 job 的运行中实例，用于阻塞策略决策。
type BlockTracker struct {
	mu       sync.Mutex
	running  map[uint64]map[int64]struct{} // jobID -> set(logID)
	logToJob map[int64]uint64              // logID -> jobID 反查，用于仅知 logID 时释放槽位
}

// NewBlockTracker 创建阻塞跟踪器。
func NewBlockTracker() *BlockTracker {
	return &BlockTracker{running: map[uint64]map[int64]struct{}{}, logToJob: map[int64]uint64{}}
}

// Allow 检查是否允许触发。
// 返回 (allow, cancelLogIDs)：cover-early 时 cancelLogIDs 为需取消的旧实例。
func (b *BlockTracker) Allow(jobID uint64, strategy int8) (bool, []int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	running := b.running[jobID]
	hasRunning := len(running) > 0
	switch strategy {
	case BlockDiscardLater:
		if hasRunning {
			return false, nil // 丢弃本次
		}
		return true, nil
	case BlockCoverEarly:
		var cancelIDs []int64
		for id := range running {
			cancelIDs = append(cancelIDs, id)
		}
		return true, cancelIDs
	default: // BlockSerial：允许，执行器侧串行排队
		return true, nil
	}
}

// OnStart 标记实例开始运行。
func (b *BlockTracker) OnStart(jobID uint64, logID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.running[jobID] == nil {
		b.running[jobID] = map[int64]struct{}{}
	}
	b.running[jobID][logID] = struct{}{}
	b.logToJob[logID] = jobID // 记录反查，供异常路径按 logID 释放
}

// OnDone 标记实例完成。
func (b *BlockTracker) OnDone(jobID uint64, logID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if m := b.running[jobID]; m != nil {
		delete(m, logID)
		if len(m) == 0 {
			delete(b.running, jobID)
		}
	}
	delete(b.logToJob, logID)
}

// ReleaseByLogID 在仅知 logID（无法关联 jobID，如 task_run 查询失败）时按 logID 释放槽位。
// 防止 BlockDiscardLater 策略下，结果回传时 DB 查询失败导致该 job 永久卡死。
func (b *BlockTracker) ReleaseByLogID(logID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	jobID, ok := b.logToJob[logID]
	if !ok {
		return
	}
	if m := b.running[jobID]; m != nil {
		delete(m, logID)
		if len(m) == 0 {
			delete(b.running, jobID)
		}
	}
	delete(b.logToJob, logID)
}
