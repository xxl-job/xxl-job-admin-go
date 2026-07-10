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
		if !isAdmin(c) {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 403, "data": nil, "msg": "需要管理员权限"})
			return
		}
		c.Next()
	}
}

// currentUID 从 context 取当前登录用户 ID（authMiddleware 注入）。
func currentUID(c *gin.Context) uint64 {
	if v, ok := c.Get("uid"); ok {
		if u, ok := v.(uint64); ok {
			return u
		}
	}
	return 0
}

// isAdmin 当前用户是否管理员。
func isAdmin(c *gin.Context) bool {
	if v, ok := c.Get("role"); ok {
		if r, ok := v.(int8); ok {
			return r == auth.RoleAdmin
		}
	}
	return false
}
