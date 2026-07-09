package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// 统一响应格式（适配 vue-admin http 拦截器）：{code:0, data, msg}。
// code 0 表示成功；非 0 由前端拦截器抛 HttpError。
const codeOK = 0

// ok 成功响应。
func ok(c *gin.Context, data any) {
	c.JSON(200, gin.H{"code": codeOK, "data": data, "msg": "ok"})
}

// fail 失败响应。
func fail(c *gin.Context, code int, msg string) {
	c.JSON(200, gin.H{"code": code, "data": nil, "msg": msg})
}

// parsePagination 解析分页参数 page/size 并做边界校正，消除各列表接口的重复逻辑（DRY）。
// page 默认 1；size 默认 20，限制 1..200。
func parsePagination(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}
	return page, size
}
