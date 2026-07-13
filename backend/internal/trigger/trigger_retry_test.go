package trigger

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
)

func newRetryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Job{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// TestRetryJobExhausted 重试次数耗尽后清除计数。
// 直接测纯计数逻辑 recordRetry，避免 retryJob 的异步 Trigger 污染（竞态）。
func TestRetryJobExhausted(t *testing.T) {
	trig := NewTrigger(nil, registry.NewMemory(), NewBlockTracker(), zap.NewNop())
	const max = int16(2)

	trig.recordRetry(1, max) // 0→1
	trig.recordRetry(1, max) // 1→2
	trig.recordRetry(1, max) // 2>=2 清除

	trig.retryMu.Lock()
	_, ok := trig.retryCount[1]
	trig.retryMu.Unlock()
	assert.False(t, ok, "retry 耗尽后应清除计数")
}

// TestRetryJobNoRetryCount fail_retry_count=0 不重试。
func TestRetryJobNoRetryCount(t *testing.T) {
	db := newRetryTestDB(t)
	db.Create(&model.Job{ID: 1, ExecutorFailRetryCount: 0})
	trig := NewTrigger(db, registry.NewMemory(), NewBlockTracker(), zap.NewNop())

	trig.retryJob(1)

	trig.retryMu.Lock()
	_, ok := trig.retryCount[1]
	trig.retryMu.Unlock()
	assert.False(t, ok, "fail_retry_count=0 不应记录重试")
}

// TestClearRetry clearRetry 清除指定 job 计数。
func TestClearRetry(t *testing.T) {
	trig := NewTrigger(nil, registry.NewMemory(), NewBlockTracker(), zap.NewNop())
	trig.retryCount[1] = 3
	trig.clearRetry(1)
	trig.retryMu.Lock()
	_, ok := trig.retryCount[1]
	trig.retryMu.Unlock()
	assert.False(t, ok, "clearRetry 应清除计数")
}

// TestOnTaskDoneSuccessClearsRetry 成功回调清除重试计数。
func TestOnTaskDoneSuccessClearsRetry(t *testing.T) {
	db := newRetryTestDB(t) // triggerChildren 查无 job，直接 return
	trig := NewTrigger(db, registry.NewMemory(), NewBlockTracker(), zap.NewNop())
	trig.retryCount[1] = 2
	trig.OnTaskDone(1, 100, 200) // 成功应清除
	trig.retryMu.Lock()
	_, ok := trig.retryCount[1]
	trig.retryMu.Unlock()
	assert.False(t, ok, "成功后应清除重试计数")
}
