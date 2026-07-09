package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/scheduler"
)

// listJobs GET /api/jobs?group=&page=&size=
func (r *Router) listJobs(c *gin.Context) {
	page, size := parsePagination(c)
	q := r.db.Model(&model.Job{})
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
	ok(c, job)
}

// createJob POST /api/jobs
func (r *Router) createJob(c *gin.Context) {
	var job model.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		fail(c, 500, err.Error())
		return
	}
	job.ID = 0
	job.TriggerStatus = 0 // 默认停止
	if err := r.db.Create(&job).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, job)
}

// updateJob PUT /api/jobs/:id
func (r *Router) updateJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	var body model.Job
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, 500, err.Error())
		return
	}
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
	if err := r.db.Save(&job).Error; err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, job)
}

// deleteJob DELETE /api/jobs/:id
func (r *Router) deleteJob(c *gin.Context) {
	if err := r.db.Delete(&model.Job{}, c.Param("id")).Error; err != nil {
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

// stopJob POST /api/jobs/:id/stop
func (r *Router) stopJob(c *gin.Context) {
	if err := r.db.Model(&model.Job{}).Where("id = ?", c.Param("id")).Update("trigger_status", 0).Error; err != nil {
		fail(c, 500, "停止任务失败: "+err.Error())
		return
	}
	ok(c, nil)
}

// triggerJob POST /api/jobs/:id/trigger —— 手动触发一次。
func (r *Router) triggerJob(c *gin.Context) {
	var job model.Job
	if err := r.db.First(&job, c.Param("id")).Error; err != nil {
		fail(c, 404, "job not found")
		return
	}
	go r.trig.Trigger(c.Request.Context(), &job)
	ok(c, gin.H{"msg": "triggered"})
}

// listGlueVersions GET /api/jobs/:id/glue-versions —— GLUE 代码版本历史（最近 30）
func (r *Router) listGlueVersions(c *gin.Context) {
	var versions []model.XxlJobLogGlue
	if err := r.db.Where("job_id = ?", c.Param("id")).Order("id desc").Limit(30).Find(&versions).Error; err != nil {
		fail(c, 500, "查询 GLUE 版本失败: "+err.Error())
		return
	}
	ok(c, gin.H{"list": versions})
}
