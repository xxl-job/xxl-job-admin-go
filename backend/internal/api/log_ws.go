package api

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsUpgrader WebSocket 升级器。CheckOrigin 做同源校验防御 CSWSH（跨站 WebSocket 劫持）。
// 开发环境前端跨域时，用 WS_ALLOWED_ORIGINS 环境变量（逗号分隔）放行额外 Origin。
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: checkSameOrigin,
}

// allowedOrigins 额外放行的 Origin 白名单（来自 WS_ALLOWED_ORIGINS 环境变量）。
var allowedOrigins = strings.Split(os.Getenv("WS_ALLOWED_ORIGINS"), ",")

// checkSameOrigin 同源校验：无 Origin（非浏览器客户端）放行；有 Origin 时要求 host 匹配，
// 或命中 WS_ALLOWED_ORIGINS 白名单。收紧 return true 带来的 CSWSH 风险。
func checkSameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	for _, o := range allowedOrigins {
		if o = strings.TrimSpace(o); o != "" {
			if strings.EqualFold(o, origin) || strings.EqualFold(o, u.Host) {
				return true
			}
		}
	}
	return false
}

// streamLog GET /api/logs/:id/stream?token=xxx —— WebSocket 实时日志推送。
// 浏览器 WS 不能自定义 header，token 通过 query 传递。
func (r *Router) streamLog(c *gin.Context) {
	token := c.Query("token")
	if _, err := r.jwt.Parse(token); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "data": nil, "msg": "unauthorized"})
		return
	}
	logID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 400, "data": nil, "msg": "invalid log id"})
		return
	}
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	ch := r.hub.Subscribe(logID)
	defer r.hub.Unsubscribe(logID, ch)
	for data := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
