// Package api 提供给前端 vue-admin 的 RESTful 接口。
package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"xxl-job-admin-go/internal/auth"
	"xxl-job-admin-go/internal/logstream"
	"xxl-job-admin-go/internal/metrics"
	"xxl-job-admin-go/internal/registry"
	"xxl-job-admin-go/internal/trigger"
	"xxl-job-admin-go/internal/web"
)

// Router 聚合 API 依赖。
type Router struct {
	db   *gorm.DB
	jwt  *auth.JWTManager
	reg  registry.Registry
	trig *trigger.Trigger
	hub  *logstream.Hub
}

// NewRouter 创建 API 路由器。
func NewRouter(db *gorm.DB, jwt *auth.JWTManager, reg registry.Registry, trig *trigger.Trigger, hub *logstream.Hub) *Router {
	return &Router{db: db, jwt: jwt, reg: reg, trig: trig, hub: hub}
}

// Engine 装配 Gin 引擎与路由。
func (r *Router) Engine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Logger(), gin.Recovery())

	e.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	e.GET("/metrics", gin.WrapH(metrics.Handler()))

	api := e.Group("/api")

	// 认证（无需 token）
	authPub := api.Group("/auth")
	{
		authPub.POST("/sessions", r.loginSession)
		authPub.POST("/tokens/refresh", r.refreshSession)
	}
	authAuthed := api.Group("/auth", r.authMiddleware())
	{
		authAuthed.DELETE("/sessions", r.logoutSession)
		authAuthed.GET("/users/me", r.me)
	}

	// WebSocket 实时日志（query token 自认证，不走标准中间件）
	api.GET("/logs/:id/stream", r.streamLog)

	// 业务（认证 + 审计）
	biz := api.Group("", r.authMiddleware(), r.auditMiddleware())
	{
		// 查询：普通用户可读
		biz.GET("/jobs", r.listJobs)
		// preview-schedule 必须先于 /jobs/:id 注册，否则会被当作 :id 通配
		biz.GET("/jobs/preview-schedule", r.previewSchedule)
		biz.GET("/jobs/:id", r.getJob)
		biz.GET("/jobs/:id/glue-versions", r.listGlueVersions)
		biz.GET("/executors", r.listExecutors)
		biz.GET("/logs", r.listLogs)
		biz.POST("/logs/:id/cancel", r.cancelLog)
		biz.GET("/dashboard", r.dashboard)

		// 任务：普通用户可管理自己创建的（handler 内按 creator_id 校验），管理员全部
		biz.POST("/jobs", r.createJob)
		biz.PUT("/jobs/:id", r.updateJob)
		biz.DELETE("/jobs/:id", r.deleteJob)
		biz.POST("/jobs/:id/start", r.startJob)
		biz.POST("/jobs/:id/stop", r.stopJob)
		biz.POST("/jobs/:id/trigger", r.triggerJob)

		// 执行器/用户/审计：仅管理员
		biz.POST("/executors", r.requireAdmin(), r.createExecutor)
		biz.PUT("/executors/:id", r.requireAdmin(), r.updateExecutor)
		biz.DELETE("/executors/:id", r.requireAdmin(), r.deleteExecutor)

		biz.GET("/audit", r.requireAdmin(), r.listAudit)

		biz.GET("/users", r.requireAdmin(), r.listUsers)
		biz.POST("/users", r.requireAdmin(), r.createUser)
		biz.PUT("/users/:id", r.requireAdmin(), r.updateUser)
		biz.DELETE("/users/:id", r.requireAdmin(), r.deleteUser)
	}

	// 前端静态文件 + SPA fallback（go:embed）
	web.RegisterStatic(e)

	return e
}
