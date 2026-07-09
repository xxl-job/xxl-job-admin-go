// Package metrics Prometheus 指标定义与 /metrics 端点。
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// TriggerTotal 任务触发总数（按 handler 维度）
	TriggerTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xxl_job_trigger_total",
		Help: "Total job triggers",
	}, []string{"handler"})

	// TriggerFailTotal 触发失败总数
	TriggerFailTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "xxl_job_trigger_fail_total",
		Help: "Total job trigger failures",
	}, []string{"handler"})

	// JobRunning 运行中任务数
	JobRunning = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "xxl_job_running",
		Help: "Number of running jobs",
	})

	// ExecutorOnline 在线执行器数（按 app）
	ExecutorOnline = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "xxl_job_executor_online",
		Help: "Online executors per app",
	}, []string{"app"})
)

func init() {
	prometheus.MustRegister(TriggerTotal, TriggerFailTotal, JobRunning, ExecutorOnline)
}

// Handler 返回 /metrics HTTP handler。
func Handler() http.Handler {
	return promhttp.Handler()
}
