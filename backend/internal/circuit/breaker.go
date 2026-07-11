// Package circuit 自研轻量熔断器（按 app 维度）。
//
// 连续失败达阈值 → open（拒绝调度），冷却后 half-open（允许一次探测），
// 探测成功 → closed，失败 → 重新 open。参考 hystrix/zeus counter 思路。
//
// 集成点：trigger 路由前 Allow(app)；rpc handleTaskResult OnResult(app, success)。
package circuit

import (
	"sync"
	"time"
)

// Breaker 按 app 维度的熔断器。
type Breaker struct {
	mu        sync.Mutex
	apps      map[string]*counter
	threshold int           // 连续失败触发熔断的阈值
	cooldown  time.Duration // open 状态冷却时长
}

type counter struct {
	consecutiveFail int
	openUntil       time.Time // 零值=closed；未来时间=open
}

// New 创建熔断器。threshold<=0 默认 5，cooldown<=0 默认 30s。
func New(threshold int, cooldown time.Duration) *Breaker {
	if threshold <= 0 {
		threshold = 5
	}
	if cooldown <= 0 {
		cooldown = 30 * time.Second
	}
	return &Breaker{apps: map[string]*counter{}, threshold: threshold, cooldown: cooldown}
}

// Allow 是否允许调度该 app。open 拒绝，closed/half-open 允许。
func (b *Breaker) Allow(app string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.apps[app]
	if c == nil {
		return true
	}
	return time.Now().After(c.openUntil) // open 期内拒绝，过后 half-open 放行
}

// IsOpen 当前是否熔断（指标/调试用）。
func (b *Breaker) IsOpen(app string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.apps[app]
	if c == nil {
		return false
	}
	return time.Now().Before(c.openUntil)
}

// OnResult 记录执行结果，更新熔断状态。
//   - success：重置连续失败计数，关闭熔断（closed）
//   - fail：累计连续失败，达阈值则进入 open（设置 openUntil）
func (b *Breaker) OnResult(app string, success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.apps[app]
	if c == nil {
		c = &counter{}
		b.apps[app] = c
	}
	if success {
		c.consecutiveFail = 0
		c.openUntil = time.Time{}
		return
	}
	c.consecutiveFail++
	if c.consecutiveFail >= b.threshold {
		c.openUntil = time.Now().Add(b.cooldown)
	}
}
