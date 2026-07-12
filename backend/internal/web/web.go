// Package web 嵌入前端构建产物（go:embed），提供静态文件服务与 SPA fallback。
// 开发模式前端走 vite dev（5174），dist 仅占位；生产 Dockerfile 注入前端 build 到 dist/。
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var distFS embed.FS

// indexHTML 预读 index.html 内容；dev 模式 dist 无 index.html 时 hasIndex=false。
var indexHTML, hasIndex = readIndex()

func readIndex() ([]byte, bool) {
	b, err := distFS.ReadFile("dist/index.html")
	if err != nil {
		return nil, false
	}
	return b, true
}

// RegisterStatic 注册前端静态文件 + SPA fallback。
// /api/ 未匹配路由返回 404 JSON；其余路径优先按静态文件服务（目录不回退），找不到则回退 index.html。
func RegisterStatic(r *gin.Engine) {
	sub, _ := fs.Sub(distFS, "dist")
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			c.JSON(404, gin.H{"code": 404, "msg": "not found"})
			return
		}
		// 尝试服务静态文件（目录不回退，避免 301 循环）
		if p != "/" {
			if f, err := sub.Open(strings.TrimPrefix(p, "/")); err == nil {
				stat, _ := f.Stat()
				_ = f.Close()
				if !stat.IsDir() {
					c.FileFromFS(p, http.FS(sub))
					return
				}
			}
		}
		// SPA fallback：直接返回 index.html 内容（绕过 http.FileServer 的目录规范化）
		if hasIndex {
			c.Data(200, "text/html; charset=utf-8", indexHTML)
			return
		}
		c.JSON(404, gin.H{"code": 404, "msg": "frontend not built (dev mode: use vite 5174)"})
	})
}
