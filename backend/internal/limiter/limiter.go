// Package limiter 按 app 维度的令牌桶限流（golang.org/x/time/rate）。
//
// 保护执行器不被某 app 的高频任务压垮：每个 app 独立令牌桶（rps + burst），
// 超速率的触发被拒绝（scheduler 记 error，下次调度再试）。
//
// 集成点：trigger 路由前 Allow(app)，与熔断器同位置。
package limiter

import (
	"sync"

	"golang.org/x/time/rate"
)

// Limiter 按 app 维度的限流器。
type Limiter struct {
	mu    sync.Mutex
	apps  map[string]*rate.Limiter
	rps   rate.Limit
	burst int
}

// New 创建限流器。rps<=0 默认 100，burst<=0 默认 100。
func New(rps, burst int) *Limiter {
	if rps <= 0 {
		rps = 100
	}
	if burst <= 0 {
		burst = 100
	}
	return &Limiter{apps: map[string]*rate.Limiter{}, rps: rate.Limit(rps), burst: burst}
}

// Allow 是否允许触发该 app。lazy 创建 per-app 令牌桶。
func (l *Limiter) Allow(app string) bool {
	l.mu.Lock()
	lim := l.apps[app]
	if lim == nil {
		lim = rate.NewLimiter(l.rps, l.burst)
		l.apps[app] = lim
	}
	l.mu.Unlock()
	return lim.Allow()
}
