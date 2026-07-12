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

	// 非 /api/ 路径不返回 404 JSON（走静态/SPA fallback）
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/job/list", nil)
	r.ServeHTTP(w2, req2)
	if strings.Contains(w2.Body.String(), `"not found"`) {
		t.Fatalf("非 /api/ 路径不应返回 404 JSON，got %s", w2.Body.String())
	}
}
