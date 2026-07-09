package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xxl-job-admin-go/internal/auth"
	"xxl-job-admin-go/internal/model"
)

// authMiddleware 校验 Authorization: Bearer <token>，注入用户信息到 context。
func (r *Router) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 401, "data": nil, "msg": "未认证"})
			return
		}
		claims, err := r.jwt.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 401, "data": nil, "msg": "token 无效"})
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// auditMiddleware 记录非 GET 操作到 audit_logs（敏感操作审计）。
func (r *Router) auditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == "GET" || c.Writer.Status() >= 500 {
			return
		}
		_ = r.db.Create(&model.AuditLog{
			Actor:  c.GetString("username"),
			Action: c.Request.Method + " " + c.FullPath(),
			Target: c.Request.URL.Path,
		}).Error
	}
}

// requireAdmin 授权中间件：仅管理员（auth.RoleAdmin）可通过，否则 403。
// 与 auth.HasPermission 的角色判定一致；普通用户调用写/管理接口将被拒绝。
func (r *Router) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		roleVal, isInt8 := role.(int8)
		if !ok || !isInt8 || roleVal != auth.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 403, "data": nil, "msg": "需要管理员权限"})
			return
		}
		c.Next()
	}
}
