// Package registry 执行器注册中心：维护 app_name → 在线执行器连接集合。
//
// 注册态由 gRPC 长连接维护——连接在即在线，断连自动摘除。
// 零外部依赖（无 Redis、无 DB 轮询），下线感知秒级。
// ExecutorConn 持有 admin→executor 的发送队列，触发器通过它下发 TriggerTask。
package registry

import (
	"sync"
	"sync/atomic"

	v1 "xxl-job-admin-go/internal/gen/xxl/v1"
	"xxl-job-admin-go/internal/metrics"
)

// LoadReport 执行器负载（路由依据）。
type LoadReport struct {
	ActiveTasks int32
	CPUUsage    float64
	QueueDepth  int32
}

// ExecutorConn 表示一个在线执行器连接。
type ExecutorConn struct {
	ID       string
	AppName  string
	Address  string
	Cluster  string // 流量泳道：执行器所属 cluster，来自注册 meta["cluster"]，空归为 default
	Handlers []string
	Meta     map[string]string

	Send chan<- *v1.AdminMessage // admin → executor 发送队列（触发器只写）
	Done <-chan struct{}         // 连接关闭信号（只读）

	load atomic.Pointer[LoadReport] // 最新负载，原子读
}

// Load 返回当前负载快照。
func (c *ExecutorConn) Load() *LoadReport {
	return c.load.Load()
}

// setLoad 原子更新负载。
func (c *ExecutorConn) setLoad(l LoadReport) {
	c.load.Store(&l)
}

// Registry 注册中心接口。
type Registry interface {
	OnConnect(app, id string, c *ExecutorConn)
	OnDisconnect(app, id string)
	Addresses(app string) []string
	Conns(app string) []*ExecutorConn
	ConnsByCluster(app, cluster string) []*ExecutorConn // 流量泳道：返回 app 下指定 cluster 的连接
	UpdateLoad(app, id string, load LoadReport)
}

type memoryRegistry struct {
	mu sync.RWMutex
	m  map[string]map[string]*ExecutorConn // app -> id -> conn
}

// NewMemory 创建内存注册中心。
func NewMemory() Registry {
	return &memoryRegistry{m: make(map[string]map[string]*ExecutorConn)}
}

func (r *memoryRegistry) OnConnect(app, id string, c *ExecutorConn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.m[app] == nil {
		r.m[app] = make(map[string]*ExecutorConn)
	}
	r.m[app][id] = c
	metrics.ExecutorOnline.WithLabelValues(app).Set(float64(len(r.m[app])))
}

func (r *memoryRegistry) OnDisconnect(app, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.m[app] != nil {
		delete(r.m[app], id)
		if len(r.m[app]) == 0 {
			delete(r.m, app) // 清理空 app 集合，避免 Gauge 残留非零
		}
		metrics.ExecutorOnline.WithLabelValues(app).Set(float64(len(r.m[app])))
	}
}

func (r *memoryRegistry) Addresses(app string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.m[app]))
	for _, c := range r.m[app] {
		out = append(out, c.Address)
	}
	return out
}

func (r *memoryRegistry) Conns(app string) []*ExecutorConn {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*ExecutorConn, 0, len(r.m[app]))
	for _, c := range r.m[app] {
		out = append(out, c)
	}
	return out
}

func (r *memoryRegistry) ConnsByCluster(app, cluster string) []*ExecutorConn {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*ExecutorConn, 0, len(r.m[app]))
	for _, c := range r.m[app] {
		if c.Cluster == cluster {
			out = append(out, c)
		}
	}
	return out
}

func (r *memoryRegistry) UpdateLoad(app, id string, load LoadReport) {
	r.mu.RLock()
	c := r.m[app][id]
	r.mu.RUnlock()
	if c != nil {
		c.setLoad(load)
	}
}
