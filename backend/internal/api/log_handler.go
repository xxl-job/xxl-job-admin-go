package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"xxl-job-admin-go/internal/model"
)

// cancelLog POST /api/logs/:id/cancel —— 终止运行中任务。
// 权限：admin，或该 TaskRun 所属 job 的创建者。
// 返回 canceled=true 表示已下发 CancelTask；任务已结束（无运行连接）则 canceled=false。
func (r *Router) cancelLog(c *gin.Context) {
	logID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || logID <= 0 {
		fail(c, 400, "非法日志ID")
		return
	}
	var run model.TaskRun
	if err := r.db.First(&run, logID).Error; err != nil {
		fail(c, 404, "日志不存在")
		return
	}
	if !isAdmin(c) {
		var job model.Job
		if err := r.db.First(&job, run.JobID).Error; err != nil {
			fail(c, 404, "任务不存在")
			return
		}
		if job.CreatorID != currentUID(c) {
			fail(c, 403, "无权操作该任务")
			return
		}
	}
	canceled := r.trig.CancelRun(logID)
	ok(c, gin.H{"canceled": canceled})
}
