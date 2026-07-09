// Package trigger 任务触发：路由策略 + 阻塞策略 + 下发。
package trigger

import (
	"hash/fnv"
	"math/rand"
	"sync"
	"sync/atomic"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
)

// 路由策略常量（对标 xxl-job ExecutorRouteStrategyEnum）
const (
	RouteFirst             int8 = 1
	RouteLast              int8 = 2
	RouteRound             int8 = 3
	RouteRandom            int8 = 4
	RouteConsistentHash    int8 = 5
	RouteLFU               int8 = 6
	RouteLRU               int8 = 7
	RouteFailover          int8 = 8
	RouteBusyover          int8 = 9
	RouteShardingBroadcast int8 = 10
)

// Router 路由策略：从候选连接中选择目标（一个或多个）。
type Router interface {
	Route(job *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn
}

// Routers 路由策略注册表（策略模式，DRY）。
type Routers struct {
	m map[int8]Router
}

// NewRouters 注册全部路由策略。
func NewRouters() *Routers {
	r := &Routers{m: map[int8]Router{}}
	r.register(RouteFirst, &firstRouter{})
	r.register(RouteLast, &lastRouter{})
	r.register(RouteRound, &roundRouter{})
	r.register(RouteRandom, &randomRouter{})
	r.register(RouteConsistentHash, &consistentHashRouter{})
	r.register(RouteLFU, &lfuRouter{})
	r.register(RouteLRU, &lruRouter{})
	r.register(RouteFailover, &firstRouter{}) // 失败重试由 trigger 层处理
	r.register(RouteBusyover, &busyoverRouter{})
	r.register(RouteShardingBroadcast, &broadcastRouter{})
	return r
}

func (r *Routers) register(t int8, router Router) { r.m[t] = router }

// Get 按策略获取路由器。
func (r *Routers) Get(t int8) Router { return r.m[t] }

// === 实现 ===

type firstRouter struct{}

func (*firstRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	return conns[:1]
}

type lastRouter struct{}

func (*lastRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	return conns[len(conns)-1:]
}

type roundRouter struct{ n uint64 }

func (r *roundRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	i := atomic.AddUint64(&r.n, 1) - 1
	idx := int(i % uint64(len(conns)))
	return conns[idx : idx+1]
}

type randomRouter struct{}

func (*randomRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	idx := rand.Intn(len(conns))
	return conns[idx : idx+1]
}

type consistentHashRouter struct{}

func (*consistentHashRouter) Route(job *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	h := fnv.New32a()
	h.Write([]byte(job.ExecutorHandler))
	idx := int(h.Sum32() % uint32(len(conns)))
	return conns[idx : idx+1]
}

type lfuRouter struct {
	mu sync.Mutex
	m  map[string]int // addr -> 使用次数
}

func (r *lfuRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.m == nil {
		r.m = map[string]int{}
	}
	var best *registry.ExecutorConn
	bestCount := int(^uint(0) >> 1)
	for _, c := range conns {
		if cnt := r.m[c.Address]; cnt < bestCount {
			bestCount = cnt
			best = c
		}
	}
	if best != nil {
		r.m[best.Address]++
		return []*registry.ExecutorConn{best}
	}
	return nil
}

type lruRouter struct {
	mu sync.Mutex
	m  map[string]int64 // addr -> 最近使用 unix nano
}

func (r *lruRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.m == nil {
		r.m = map[string]int64{}
	}
	var best *registry.ExecutorConn
	var bestT int64
	for i, c := range conns {
		t := r.m[c.Address]
		if i == 0 || t < bestT {
			bestT = t
			best = c
		}
	}
	if best != nil {
		r.m[best.Address] = timeNowNanos()
		return []*registry.ExecutorConn{best}
	}
	return nil
}

type busyoverRouter struct{}

func (*busyoverRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	if len(conns) == 0 {
		return nil
	}
	var best *registry.ExecutorConn
	bestLoad := int32(^int32(0) >> 1)
	for _, c := range conns {
		var l int32
		if lp := c.Load(); lp != nil {
			l = lp.ActiveTasks
		}
		if l < bestLoad {
			bestLoad = l
			best = c
		}
	}
	if best != nil {
		return []*registry.ExecutorConn{best}
	}
	return nil
}

type broadcastRouter struct{}

func (*broadcastRouter) Route(_ *model.Job, conns []*registry.ExecutorConn) []*registry.ExecutorConn {
	out := make([]*registry.ExecutorConn, len(conns))
	copy(out, conns)
	return out // 分片广播：全部执行器
}
