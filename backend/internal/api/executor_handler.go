package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/model"
)

// listExecutors GET /api/executors
// 合并视图：DB 分组 + 内存在线未分组 app（虚拟项 id=0）。
func (r *Router) listExecutors(c *gin.Context) {
	var exs []model.Executor
	if err := r.db.Order("id desc").Find(&exs).Error; err != nil {
		fail(c, 500, "查询执行器列表失败: "+err.Error())
		return
	}

	// DB 已有的 app 集合
	known := make(map[string]struct{}, len(exs))
	for _, e := range exs {
		known[e.AppName] = struct{}{}
	}

	// 内存在线但 DB 未配置的 app → 补虚拟项
	for _, app := range r.reg.Apps() {
		if _, ok := known[app]; ok {
			continue
		}
		exs = append(exs, model.Executor{
			ID: 0, AppName: app, Title: "(未配置)", AddressType: 0,
		})
	}

	out := make([]gin.H, 0, len(exs))
	for _, e := range exs {
		out = append(out, gin.H{
			"executor": e,
			"online":   r.reg.Addresses(e.AppName),
		})
	}
	ok(c, gin.H{"list": out})
}

// createExecutor POST /api/executors
func (r *Router) createExecutor(c *gin.Context) {
	var ex model.Executor
	if err := c.ShouldBindJSON(&ex); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ex.ID = 0
	if err := r.db.Create(&ex).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, ex)
}

// updateExecutor PUT /api/executors/:id
func (r *Router) updateExecutor(c *gin.Context) {
	var ex model.Executor
	if err := r.db.First(&ex, c.Param("id")).Error; err != nil {
		fail(c, 404, "executor not found")
		return
	}
	var body model.Executor
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ex.AppName = body.AppName
	ex.Title = body.Title
	ex.AddressType = body.AddressType
	ex.AddressList = body.AddressList
	if err := r.db.Save(&ex).Error; err != nil {
		fail(c, 500, "保存执行器失败: "+err.Error())
		return
	}
	ok(c, ex)
}

// deleteExecutor DELETE /api/executors/:id
func (r *Router) deleteExecutor(c *gin.Context) {
	if err := r.db.Delete(&model.Executor{}, c.Param("id")).Error; err != nil {
		fail(c, 500, "删除执行器失败: "+err.Error())
		return
	}
	ok(c, nil)
}

// listLogs GET /api/logs?job_id=&page=&size=&status=
func (r *Router) listLogs(c *gin.Context) {
	page, size := parsePagination(c)
	q := r.db.Model(&model.TaskRun{})
	if j := c.Query("job_id"); j != "" {
		q = q.Where("job_id = ?", j)
	}
	if s := c.Query("status"); s == "success" {
		q = q.Where("handle_code = ?", 200)
	} else if s == "fail" {
		q = q.Where("handle_code <> ? AND handle_code <> 0", 200)
	}
	if h := c.Query("handler"); h != "" {
		q = q.Where("executor_handler LIKE ?", "%"+h+"%")
	}
	if st := c.Query("start_time"); st != "" {
		q = q.Where("trigger_time >= ?", st)
	}
	if et := c.Query("end_time"); et != "" {
		q = q.Where("trigger_time <= ?", et)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		fail(c, 500, "查询日志总数失败: "+err.Error())
		return
	}
	var logs []model.TaskRun
	if err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&logs).Error; err != nil {
		fail(c, 500, "查询日志列表失败: "+err.Error())
		return
	}
	ok(c, gin.H{"list": logs, "total": total})
}

// dashboard GET /api/dashboard —— 运行报表。
func (r *Router) dashboard(c *gin.Context) {
	// count 封装 Count 调用，失败时 fail 并通过返回的 good=false 让外层提前返回。
	count := func(tx *gorm.DB) (int64, bool) {
		var n int64
		if err := tx.Count(&n).Error; err != nil {
			fail(c, 500, "dashboard 查询失败: "+err.Error())
			return 0, false
		}
		return n, true
	}

	jobTotal, good := count(r.db.Model(&model.Job{}))
	if !good {
		return
	}
	jobRunning, good := count(r.db.Model(&model.Job{}).Where("trigger_status = 1"))
	if !good {
		return
	}
	execTotal, good := count(r.db.Model(&model.Executor{}))
	if !good {
		return
	}

	today := time.Now().Format("2006-01-02")
	todayTotal, good := count(r.db.Model(&model.TaskRun{}).Where("DATE(trigger_time) = ?", today))
	if !good {
		return
	}
	todaySuccess, good := count(r.db.Model(&model.TaskRun{}).Where("DATE(trigger_time) = ? AND handle_code = ?", today, 200))
	if !good {
		return
	}
	todayFail, good := count(r.db.Model(&model.TaskRun{}).Where("DATE(trigger_time) = ? AND handle_code <> ? AND handle_code <> 0", today, 200))
	if !good {
		return
	}

	ok(c, gin.H{
		"job_total":     jobTotal,
		"job_running":   jobRunning,
		"exec_total":    execTotal,
		"today_total":   todayTotal,
		"today_success": todaySuccess,
		"today_fail":    todayFail,
	})
}
