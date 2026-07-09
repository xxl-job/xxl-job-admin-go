package scheduler

import (
	"sync"
	"time"
)

// workerPool 固定大小 worker 池。
type workerPool struct {
	tasks chan func()
	wg    sync.WaitGroup
}

func newWorkerPool(n int) *workerPool {
	if n < 1 {
		n = 1
	}
	p := &workerPool{tasks: make(chan func(), n*4)}
	for i := 0; i < n; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

func (p *workerPool) worker() {
	defer p.wg.Done()
	for t := range p.tasks {
		t()
	}
}

func (p *workerPool) submit(t func()) { p.tasks <- t }

func (p *workerPool) stop() {
	close(p.tasks)
	p.wg.Wait()
}

// Pool 快慢双线程池：慢任务降级到 slow pool，避免耗尽调度线程。
// 对标 Java 版 JobTriggerPoolHelper 的快慢隔离设计。
type Pool struct {
	fast          *workerPool
	slow          *workerPool
	slowThreshold time.Duration
	mu            sync.Mutex
	slowJobs      map[uint64]struct{} // 已标记为慢的 job
}

// NewPool 创建快慢线程池。fastN/slowN 为 worker 数，threshold 为慢任务降级阈值。
func NewPool(fastN, slowN int, threshold time.Duration) *Pool {
	return &Pool{
		fast:          newWorkerPool(fastN),
		slow:          newWorkerPool(slowN),
		slowThreshold: threshold,
		slowJobs:      make(map[uint64]struct{}),
	}
}

// Submit 提交任务。慢任务（标记或 hint 超阈值）走 slow pool。
func (p *Pool) Submit(jobID uint64, task func(), hint time.Duration) {
	if hint > p.slowThreshold || p.isSlow(jobID) {
		p.slow.submit(task)
		return
	}
	p.fast.submit(task)
}

// MarkSlow 标记 job 为慢任务（后续调度降级到 slow pool）。
func (p *Pool) MarkSlow(jobID uint64) {
	p.mu.Lock()
	p.slowJobs[jobID] = struct{}{}
	p.mu.Unlock()
}

// MarkFast 取消慢标记。
func (p *Pool) MarkFast(jobID uint64) {
	p.mu.Lock()
	delete(p.slowJobs, jobID)
	p.mu.Unlock()
}

func (p *Pool) isSlow(jobID uint64) bool {
	p.mu.Lock()
	_, ok := p.slowJobs[jobID]
	p.mu.Unlock()
	return ok
}

// Stop 关闭池。
func (p *Pool) Stop() {
	p.fast.stop()
	p.slow.stop()
}
