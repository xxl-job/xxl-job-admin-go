package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/scheduler"
)

// assertJobOwner 校验当前用户对 job 的访问权：admin 放行，否则仅创建者。
func assertJobOwner(c *gin.Context, job *model.Job) bool {
	return isAdmin(c) || job.CreatorID == currentUID(c)
}

// listJobs GET /api/jobs?group=&page=&size=
// 普通用户仅见自己创建的任务；管理员见全部。
func (r *Router) listJobs(c *gin.Context) {
	page, size := parsePagination(c)
	q := r.db.Model(&model.Job{})
	if !isAdmin(c) {
		q = q.Where("creator_id = ?", currentUID(c))
	}
	if g := c.Query("group"); g != "" {
		q = q.Where("job_group = ?", g)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		fail(c, 500, "查询任务总数失败: "+err.Error())
		return
	}
	var jobs []model.Job
	if err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&jobs).Error; err != nil {
		fail(c, 500, "查询任务列表失败: "+err.Error())
		return
	}
	ok(c, gin.H{"list": jobs, "total": total})
}

// getJob GET /api/jobs/:id
func (r *Router) getJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权访问该任务")
		return
	}
	ok(c, job)
}

// createJob POST /api/jobs —— creator_id 自动填当前用户。
func (r *Router) createJob(c *gin.Context) {
	var job model.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		fail(c, 500, err.Error())
		return
	}
	job.ID = 0
	job.TriggerStatus = 0            // 默认停止
	job.CreatorID = currentUID(c)    // 创建者：当前登录用户
	if err := r.db.Create(&job).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, job)
}

// updateJob PUT /api/jobs/:id —— 仅创建者或管理员。
func (r *Router) updateJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权操作该任务")
		return
	}
	var body model.Job
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, 500, err.Error())
		return
	}
	// 记录原调度配置，用于检测变更并热重算下次触发
	oldSchedType := job.ScheduleType
	oldSchedConf := job.ScheduleConf
	job.JobDesc = body.JobDesc
	job.JobGroup = body.JobGroup
	job.ScheduleType = body.ScheduleType
	job.ScheduleConf = body.ScheduleConf
	job.ExecutorHandler = body.ExecutorHandler
	job.ExecutorParam = body.ExecutorParam
	job.ExecutorRouteStrategy = body.ExecutorRouteStrategy
	job.ExecutorBlockStrategy = body.ExecutorBlockStrategy
	job.ExecutorTimeout = body.ExecutorTimeout
	job.ExecutorFailRetryCount = body.ExecutorFailRetryCount
	job.ChildJobID = body.ChildJobID
	job.Cluster = body.Cluster
	// GLUE 代码变更则保存历史版本（最多 30）
	if job.GlueType != 0 && body.GlueSource != "" && body.GlueSource != job.GlueSource {
		if err := r.db.Create(&model.XxlJobLogGlue{
			JobID:      job.ID,
			GlueType:   job.GlueType,
			GlueSource: job.GlueSource,
			Version:    int(time.Now().Unix()),
		}).Error; err != nil {
			fail(c, 500, "保存 GLUE 历史版本失败: "+err.Error())
			return
		}
		// 保留最近 30 版本，删除更早的
		if err := r.db.Where("job_id = ? AND id NOT IN (SELECT id FROM (SELECT id FROM xxl_job_logglue WHERE job_id = ? ORDER BY id DESC LIMIT 30) AS t)", job.ID, job.ID).Delete(&model.XxlJobLogGlue{}).Error; err != nil {
			fail(c, 500, "清理 GLUE 历史版本失败: "+err.Error())
			return
		}
	}
	job.GlueType = body.GlueType
	job.GlueSource = body.GlueSource
	// 调度配置热生效：运行中任务改 CRON/频率后重算下次触发（无需 stop/start）
	if (job.ScheduleType != oldSchedType || job.ScheduleConf != oldSchedConf) && job.TriggerStatus == 1 {
		next, err := scheduler.NextTriggerTime(job.ScheduleType, job.ScheduleConf, time.Now())
		if err != nil {
			fail(c, 500, "调度配置无效: "+err.Error())
			return
		}
		job.TriggerNextTime = next
	}
	if err := r.db.Save(&job).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, job)
}

// deleteJob DELETE /api/jobs/:id —— 仅创建者或管理员。
func (r *Router) deleteJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权操作该任务")
		return
	}
	if err := r.db.Delete(&job).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, nil)
}

// startJob POST /api/jobs/:id/start —— 计算下次触发时间并启动。
func (r *Router) startJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权操作该任务")
		return
	}
	next, err := scheduler.NextTriggerTime(job.ScheduleType, job.ScheduleConf, time.Now())
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := r.db.Model(&job).Updates(map[string]interface{}{
		"trigger_status":    1,
		"trigger_next_time": next,
	}).Error; err != nil {
		fail(c, 500, "启动任务失败: "+err.Error())
		return
	}
	ok(c, gin.H{"trigger_next_time": next})
}

// stopJob POST /api/jobs/:id/stop —— 仅创建者或管理员。
func (r *Router) stopJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权操作该任务")
		return
	}
	if err := r.db.Model(&job).Update("trigger_status", 0).Error; err != nil {
		fail(c, 500, "停止任务失败: "+err.Error())
		return
	}
	ok(c, nil)
}

// triggerJob POST /api/jobs/:id/trigger —— 手动触发一次，支持临时参数与指定机器。
// body 可选: { "executor_param": "...", "executor_address": "..." }
func (r *Router) triggerJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权操作该任务")
		return
	}
	var body struct {
		ExecutorParam   string `json:"executor_param"`
		ExecutorAddress string `json:"executor_address"`
	}
	_ = c.ShouldBindJSON(&body) // body 可空（兼容旧调用）

	if body.ExecutorAddress != "" {
		// 指定机器：同步返回结果（含错误），便于前端内联提示
		if err := r.trig.TriggerOnce(c.Request.Context(), &job, body.ExecutorParam, body.ExecutorAddress); err != nil {
			fail(c, 400, "触发失败: "+err.Error())
			return
		}
		ok(c, gin.H{"msg": "triggered"})
		return
	}

	// 自动路由：异步触发
	go r.trig.TriggerOnce(c.Request.Context(), &job, body.ExecutorParam, "")
	ok(c, gin.H{"msg": "triggered"})
}

// listGlueVersions GET /api/jobs/:id/glue-versions —— GLUE 代码版本历史（最近 30）
func (r *Router) listGlueVersions(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	if !assertJobOwner(c, &job) {
		fail(c, 403, "无权访问该任务")
		return
	}
	var versions []model.XxlJobLogGlue
	if err := r.db.Where("job_id = ?", job.ID).Order("id desc").Limit(30).Find(&versions).Error; err != nil {
		fail(c, 500, "查询 GLUE 版本失败: "+err.Error())
		return
	}
	ok(c, gin.H{"list": versions})
}
