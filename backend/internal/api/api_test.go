package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xxl-job-admin-go/internal/auth"
	"xxl-job-admin-go/internal/logstream"
	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
	"xxl-job-admin-go/internal/trigger"
)

// setupTestRouter 用 sqlite 内存库构造测试 Router。
func setupTestRouter(t *testing.T) (*Router, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Job{}, &model.Executor{}, &model.TaskRun{}, &model.XxlJobLogGlue{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	db.Create(&model.User{Username: "admin", Password: string(hash), Role: 1})

	jwtMgr := auth.NewJWTManager("test-secret", time.Hour)
	reg := registry.NewMemory()
	block := trigger.NewBlockTracker()
	trig := trigger.NewTrigger(db, reg, block, zap.NewNop())
	return NewRouter(db, jwtMgr, reg, trig, logstream.New()), db
}

func doJSON(r *Router, method, path, token string, body any) (int, map[string]any) {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.Engine().ServeHTTP(w, req)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

func TestLoginSuccess(t *testing.T) {
	r, _ := setupTestRouter(t)
	code, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	assert.Equal(t, 200, code)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["accessToken"])
	assert.NotEmpty(t, data["refreshToken"])
}

func TestLoginFail(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "wrong"})
	assert.NotEqual(t, float64(0), resp["code"], "错误密码应失败")
}

func TestAuthRequired(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "GET", "/api/jobs", "", nil)
	assert.NotEqual(t, float64(0), resp["code"], "无 token 应被拦截")
}

func TestJobCRUDWithAuth(t *testing.T) {
	r, db := setupTestRouter(t)
	db.Create(&model.Executor{AppName: "demo-app", Title: "Demo"})

	// 登录拿 token
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)

	// 创建 job
	_, resp = doJSON(r, "POST", "/api/jobs", token, map[string]any{
		"job_group": 1, "job_desc": "test job", "schedule_type": 1, "schedule_conf": "0 * * * * ?",
		"executor_handler": "demoJob", "executor_route_strategy": 1, "executor_block_strategy": 0,
	})
	assert.Equal(t, float64(0), resp["code"])
	jobData := resp["data"].(map[string]any)
	assert.Equal(t, "test job", jobData["job_desc"])

	// 列表
	_, resp = doJSON(r, "GET", "/api/jobs", token, nil)
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(1), data["total"])

	// 启动
	jobID := jobData["id"]
	_, resp = doJSON(r, "POST", "/api/jobs/"+itoa(jobID)+"/start", token, nil)
	assert.Equal(t, float64(0), resp["code"])

	// 停止
	_, resp = doJSON(r, "POST", "/api/jobs/"+itoa(jobID)+"/stop", token, nil)
	assert.Equal(t, float64(0), resp["code"])

	// 删除
	_, resp = doJSON(r, "DELETE", "/api/jobs/"+itoa(jobID), token, nil)
	assert.Equal(t, float64(0), resp["code"])
}

func TestUsersMe(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)
	_, resp = doJSON(r, "GET", "/api/auth/users/me", token, nil)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]any)
	assert.Equal(t, "admin", data["username"])
	assert.Contains(t, data["roles"], "super_admin")
}

func TestExecutorCRUD(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)

	// 创建 executor
	_, resp = doJSON(r, "POST", "/api/executors", token, map[string]any{"app_name": "app1", "title": "T", "address_type": 0})
	assert.Equal(t, float64(0), resp["code"])

	// 列表
	_, resp = doJSON(r, "GET", "/api/executors", token, nil)
	data := resp["data"].(map[string]any)
	assert.Equal(t, 1, len(data["list"].([]any)))
}

func TestUserCRUD(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)

	// 创建普通用户
	_, resp = doJSON(r, "POST", "/api/users", token, map[string]any{"username": "u1", "password": "p1", "role": 0})
	assert.Equal(t, float64(0), resp["code"])

	// 列表（admin + u1）
	_, resp = doJSON(r, "GET", "/api/users", token, nil)
	data := resp["data"].(map[string]any)
	assert.Equal(t, 2, len(data["list"].([]any)))

	// 禁删超级管理员 id=1
	_, resp = doJSON(r, "DELETE", "/api/users/1", token, nil)
	assert.NotEqual(t, float64(0), resp["code"], "不应删除超级管理员")
}

func TestDashboard(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)

	_, resp = doJSON(r, "GET", "/api/dashboard", token, nil)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(0), data["exec_total"]) // 初始无执行器
	// job_total 初始 0
	assert.Contains(t, data, "job_total")
}

// itoa 简单数字转字符串（避免引入 strconv 到测试 helper 的类型断言链）。
func itoa(v any) string {
	switch n := v.(type) {
	case float64:
		return strconvFormatInt(int64(n))
	default:
		return ""
	}
}

func strconvFormatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// TestListExecutorsMergesOnline 验证：DB 有分组的展示其在线地址；
// 内存在线但 DB 无分组的 app 以虚拟项(id=0)展示。
func TestListExecutorsMergesOnline(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)

	// DB 建分组 configured-app
	doJSON(r, "POST", "/api/executors", token, map[string]any{"app_name": "configured-app", "title": "已配置", "address_type": 0})

	// 内存注册两台：一台属 configured-app，一台属未配置的 loose-app
	r.reg.OnConnect("configured-app", "c1", &registry.ExecutorConn{AppName: "configured-app", Address: "1.1.1.1:9999"})
	r.reg.OnConnect("loose-app", "l1", &registry.ExecutorConn{AppName: "loose-app", Address: "2.2.2.2:9999"})

	_, resp = doJSON(r, "GET", "/api/executors", token, nil)
	list := resp["data"].(map[string]any)["list"].([]any)
	// 期望 2 项：configured-app（有在线）+ loose-app（虚拟项 id=0）
	if len(list) != 2 {
		t.Fatalf("expect 2 entries, got %d (%+v)", len(list), list)
	}
	byApp := map[string]map[string]any{}
	for _, it := range list {
		m := it.(map[string]any)
		ex := m["executor"].(map[string]any)
		byApp[ex["app_name"].(string)] = m
	}
	if m, ok := byApp["loose-app"]; !ok || m["executor"].(map[string]any)["id"] != float64(0) {
		t.Fatalf("loose-app 应以虚拟项 id=0 展示，got %+v", byApp["loose-app"])
	}
}

// TestRBAC_NonAdminForbidden 验证普通用户(role=0)任务按创建者隔离：
// 可创建/管理自己的 job，不能触碰他人的 job（403）；executor/users/audit 仍仅管理员。
func TestRBAC_NonAdminForbidden(t *testing.T) {
	r, _ := setupTestRouter(t)
	token, err := r.jwt.Sign(99, "normal", auth.RoleUser)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	// admin 先建一个 job 作为"他人任务"
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	adminToken := resp["data"].(map[string]any)["accessToken"].(string)
	_, resp = doJSON(r, "POST", "/api/jobs", adminToken, map[string]any{
		"job_group": 1, "job_desc": "admin-job", "schedule_type": 1, "schedule_conf": "0 * * * * ?",
		"executor_handler": "h", "executor_route_strategy": 1, "executor_block_strategy": 0,
	})
	assert.Equal(t, float64(0), resp["code"], "admin 建 job")
	adminJobID := itoa(resp["data"].(map[string]any)["id"])

	// 普通用户操作 admin 的 job → 403（owner 校验）
	_, resp = doJSON(r, "DELETE", "/api/jobs/"+adminJobID, token, nil)
	assert.Equal(t, float64(403), resp["code"], "普通用户删他人 job 应 403")
	_, resp = doJSON(r, "POST", "/api/jobs/"+adminJobID+"/trigger", token, nil)
	assert.Equal(t, float64(403), resp["code"], "普通用户触发他人 job 应 403")

	// 普通用户建自己的 job → 成功（creator_id 自动填 99）
	_, resp = doJSON(r, "POST", "/api/jobs", token, map[string]any{
		"job_group": 1, "job_desc": "mine", "schedule_type": 1, "schedule_conf": "0 * * * * ?",
		"executor_handler": "h", "executor_route_strategy": 1, "executor_block_strategy": 0,
	})
	assert.Equal(t, float64(0), resp["code"], "普通用户应能创建 job")
	myJobID := itoa(resp["data"].(map[string]any)["id"])

	// 普通用户触发自己的 job → 成功
	_, resp = doJSON(r, "POST", "/api/jobs/"+myJobID+"/trigger", token, nil)
	assert.Equal(t, float64(0), resp["code"], "普通用户应能触发自己的 job")

	// 普通用户列表只见自己创建的（不含 admin 的 job）
	_, resp = doJSON(r, "GET", "/api/jobs", token, nil)
	assert.Equal(t, float64(0), resp["code"], "普通用户查 job 列表应放行")
	list := resp["data"].(map[string]any)["list"].([]any)
	assert.NotEmpty(t, list, "应至少有自己的 job")
	for _, j := range list {
		creator := j.(map[string]any)["creator_id"]
		assert.Equal(t, float64(99), creator, "普通用户不应看到他人的 job")
	}

	// executor/users/audit 仍仅管理员
	_, resp = doJSON(r, "DELETE", "/api/executors/1", token, nil)
	assert.Equal(t, float64(403), resp["code"], "普通用户删 executor 应被拒绝")
	_, resp = doJSON(r, "GET", "/api/users", token, nil)
	assert.Equal(t, float64(403), resp["code"], "普通用户查用户列表应被拒绝")
	_, resp = doJSON(r, "POST", "/api/users", token, map[string]any{"username": "x", "password": "p", "role": 0})
	assert.Equal(t, float64(403), resp["code"], "普通用户建用户应被拒绝")
	_, resp = doJSON(r, "GET", "/api/audit", token, nil)
	assert.Equal(t, float64(403), resp["code"], "普通用户查审计应被拒绝")
}

// TestRBAC_AdminAllowed 验证 admin 写操作与敏感查询正常通过授权。
func TestRBAC_AdminAllowed(t *testing.T) {
	r, _ := setupTestRouter(t)
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	token := resp["data"].(map[string]any)["accessToken"].(string)
	_, resp = doJSON(r, "GET", "/api/users", token, nil)
	assert.Equal(t, float64(0), resp["code"], "admin 查用户列表应通过")
}

// TestCancelLogPermission 验证：终止日志需 admin 或 job 创建者；他人 403。
func TestCancelLogPermission(t *testing.T) {
	r, _ := setupTestRouter(t)
	token, _ := r.jwt.Sign(99, "normal", auth.RoleUser)

	// admin 建 job + 触发产生 run
	_, resp := doJSON(r, "POST", "/api/auth/sessions", "", map[string]string{"username": "admin", "password": "123456"})
	adminToken := resp["data"].(map[string]any)["accessToken"].(string)
	doJSON(r, "POST", "/api/jobs", adminToken, map[string]any{"job_group": 1, "job_desc": "j", "schedule_type": 1, "schedule_conf": "0 * * * * ?", "executor_handler": "h"})
	// 直接造一条 run 归属 admin 的 job（id=1）
	r.db.Create(&model.TaskRun{JobID: 1, ExecutorHandler: "h"})

	_, resp = doJSON(r, "POST", "/api/logs/1/cancel", token, nil)
	assert.Equal(t, float64(403), resp["code"], "普通用户终止他人 job 的日志应 403")

	_, resp = doJSON(r, "POST", "/api/logs/1/cancel", adminToken, nil)
	assert.Equal(t, float64(0), resp["code"], "admin 应放行")
}
