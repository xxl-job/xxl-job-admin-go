package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xxl-job-admin-go/internal/model"
)

// mockTrigger 记录被触发的 jobID（线程安全）。
type mockTrigger struct {
	mu   sync.Mutex
	jobs []uint64
}

func (m *mockTrigger) Trigger(_ context.Context, job *model.Job) {
	m.mu.Lock()
	m.jobs = append(m.jobs, job.ID)
	m.mu.Unlock()
}

func (m *mockTrigger) contains(id uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range m.jobs {
		if j == id {
			return true
		}
	}
	return false
}

func newTestScheduler(t *testing.T) (*Scheduler, *gorm.DB, *mockTrigger) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Job{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	mt := &mockTrigger{}
	pool := NewPool(2, 1, 500*time.Millisecond)
	s := New(db, pool, mt, zap.NewNop())
	return s, db, mt
}

// TestFireSecondTriggersDueJob 过期任务应被触发并更新 next_time。
func TestFireSecondTriggersDueJob(t *testing.T) {
	s, db, mt := newTestScheduler(t)
	past := time.Now().Add(-10 * time.Second).UnixMilli()
	db.Create(&model.Job{ID: 1, JobGroup: 1, ScheduleType: ScheduleCron, ScheduleConf: "0 * * * * ?", TriggerStatus: 1, TriggerNextTime: past})

	s.preload(context.Background())
	s.fireSecond(context.Background(), time.Now().Unix())

	// 等待 pool 异步执行 trigger
	time.Sleep(100 * time.Millisecond)
	if !mt.contains(1) {
		t.Fatal("过期任务应被触发")
	}

	var job model.Job
	db.First(&job, 1)
	if job.TriggerNextTime == past {
		t.Fatal("trigger_next_time 应已更新到未来")
	}
	if job.TriggerLastTime != past {
		t.Fatal("trigger_last_time 应记录上次触发时间")
	}
}

// TestFireSecondSkipsStoppedJob 停止的任务不应被预读触发。
func TestFireSecondSkipsStoppedJob(t *testing.T) {
	s, db, mt := newTestScheduler(t)
	past := time.Now().Add(-10 * time.Second).UnixMilli()
	db.Create(&model.Job{ID: 1, JobGroup: 1, ScheduleType: ScheduleCron, ScheduleConf: "0 * * * * ?", TriggerStatus: 0, TriggerNextTime: past})

	s.preload(context.Background())
	s.fireSecond(context.Background(), time.Now().Unix())
	time.Sleep(50 * time.Millisecond)

	if mt.contains(1) {
		t.Fatal("停止的任务不应被触发")
	}
}

// TestPreloadOnlyReadsDueJobs 预读只读 trigger_next_time 在窗口内的任务。
func TestPreloadOnlyReadsDueJobs(t *testing.T) {
	s, db, _ := newTestScheduler(t)
	past := time.Now().Add(-10 * time.Second).UnixMilli()
	far := time.Now().Add(10 * time.Minute).UnixMilli()
	db.Create(&model.Job{ID: 1, ScheduleType: ScheduleCron, ScheduleConf: "0 * * * * ?", TriggerStatus: 1, TriggerNextTime: past})
	db.Create(&model.Job{ID: 2, ScheduleType: ScheduleCron, ScheduleConf: "0 * * * * ?", TriggerStatus: 1, TriggerNextTime: far})

	s.preload(context.Background())
	s.mu.Lock()
	count := 0
	for _, entries := range s.ring {
		count += len(entries)
	}
	s.mu.Unlock()

	if count != 1 {
		t.Fatalf("预读应只入 1 个到期任务，got %d", count)
	}
}
