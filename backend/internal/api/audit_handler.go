package api

import (
	"github.com/gin-gonic/gin"

	"xxl-job-admin-go/internal/model"
)

// listAudit GET /api/audit?page=&size= —— 审计日志查询。
func (r *Router) listAudit(c *gin.Context) {
	page, size := parsePagination(c)
	q := r.db.Model(&model.AuditLog{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		fail(c, 500, "查询审计总数失败: "+err.Error())
		return
	}
	var logs []model.AuditLog
	if err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&logs).Error; err != nil {
		fail(c, 500, "查询审计列表失败: "+err.Error())
		return
	}
	ok(c, gin.H{"list": logs, "total": total})
}
