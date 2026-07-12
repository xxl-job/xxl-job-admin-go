package web

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterStaticRouteLogic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterStatic(r)

	// /api/ 路径返回 404 JSON（不回退 SPA）
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	r.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Fatalf("/api/ 应 404，got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "not found") {
		t.Fatalf("/api/ 应返回 not found JSON，got %s", w.Body.String())
	}

	// / 首页不应 301/308 重定向（曾因 http.FileServer 目录规范化导致 ./ 循环）
	wRoot := httptest.NewRecorder()
	reqRoot := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(wRoot, reqRoot)
	if wRoot.Code == 301 || wRoot.Code == 302 || wRoot.Code == 307 || wRoot.Code == 308 {
		t.Fatalf("/ 不应重定向，got %d Location=%s", wRoot.Code, wRoot.Header().Get("Location"))
	}

	// 非 /api/ 路径不返回 /api/ 风格的 not found JSON：
	// 生产模式回退 index.html（200）；dev 模式（dist 无 index.html）返回 "frontend not built"。
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/job/list", nil)
	r.ServeHTTP(w2, req2)
	if strings.Contains(w2.Body.String(), `"not found"`) {
		t.Fatalf("非 /api/ 路径不应返回 /api/ 风格 not found JSON，got %s", w2.Body.String())
	}
}
