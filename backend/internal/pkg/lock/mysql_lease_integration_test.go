//go:build integration

// 真集成测试：用 testcontainers 启 MySQL 容器，验证租约锁在真实 DB 下的抢锁/续约/故障转移/释放。
// 默认 `go test` 不执行（build tag=integration），需 `go test -tags=integration ./...`。
// 无 Docker 环境自动 t.Skip 跳过，不阻塞单测流程。

package lock

import (
	"context"
	"testing"
	"time"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupMySQL 启 MySQL 容器并建 scheduler_lease 表。无 Docker 则跳过。
func setupMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()
	container, err := tcmysql.Run(ctx, "mysql:8.0")
	if err != nil {
		t.Skipf("跳过：无法启动 MySQL 容器（无 Docker？）: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	dsn, cerr := container.ConnectionString(ctx)
	if cerr != nil {
		t.Fatalf("connection string: %v", cerr)
	}
	dsn += "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS scheduler_lease (
		id TINYINT NOT NULL DEFAULT 1,
		owner VARCHAR(64) NOT NULL DEFAULT '',
		lease_until DATETIME(3) NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT chk_single CHECK (id = 1)
	) ENGINE=InnoDB`)
	db.Exec("INSERT INTO scheduler_lease (id, owner, lease_until) VALUES (1, '', '1970-01-01 00:00:00')")
	return db
}

// TestLease_AcquireRenew 抢锁与续约的互斥性：首节点抢到，次节点抢不到；owner 续约成功，非 owner 失败。
func TestLease_AcquireRenew(t *testing.T) {
	db := setupMySQL(t)
	l1 := NewMySQLLeaseLocker(db, "node1", 30*time.Second)
	l2 := NewMySQLLeaseLocker(db, "node2", 30*time.Second)
	ctx := context.Background()

	if ok, _ := l1.TryAcquire(ctx, "k"); !ok {
		t.Fatal("node1 首次应抢到租约")
	}
	if ok, _ := l2.TryAcquire(ctx, "k"); ok {
		t.Fatal("node2 在 node1 持锁期间不应抢到")
	}
	if ok, _ := l1.Renew(ctx, "k"); !ok {
		t.Fatal("owner node1 续约应成功")
	}
	if ok, _ := l2.Renew(ctx, "k"); ok {
		t.Fatal("非 owner node2 续约应失败")
	}
}

// TestLease_Failover 故障转移：owner 停续约 → TTL 过期 → 备节点抢到（核心可靠性场景）。
func TestLease_Failover(t *testing.T) {
	db := setupMySQL(t)
	l1 := NewMySQLLeaseLocker(db, "node1", 2*time.Second) // 短 TTL 加速测试
	l2 := NewMySQLLeaseLocker(db, "node2", 2*time.Second)
	ctx := context.Background()

	if ok, _ := l1.TryAcquire(ctx, "k"); !ok {
		t.Fatal("node1 应抢到")
	}
	// node1 停止续约，等待租约过期
	time.Sleep(3 * time.Second)
	if ok, _ := l2.TryAcquire(ctx, "k"); !ok {
		t.Fatal("租约过期后 node2 应抢到（故障转移）")
	}
	// node1 恢复续约应失败（已被 node2 抢占）
	if ok, _ := l1.Renew(ctx, "k"); ok {
		t.Fatal("node1 恢复后续约应失败（已被抢占）")
	}
}

// TestLease_Release 主动释放后其他节点可抢。
func TestLease_Release(t *testing.T) {
	db := setupMySQL(t)
	l1 := NewMySQLLeaseLocker(db, "node1", 30*time.Second)
	l2 := NewMySQLLeaseLocker(db, "node2", 30*time.Second)
	ctx := context.Background()

	l1.TryAcquire(ctx, "k")
	if err := l1.Release(ctx, "k"); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if ok, _ := l2.TryAcquire(ctx, "k"); !ok {
		t.Fatal("Release 后 node2 应抢到")
	}
	// 非 owner Release 无效（node2 持锁，node1 Release 不应影响）
	l1.Release(ctx, "k")
	if ok, _ := l2.Renew(ctx, "k"); !ok {
		t.Fatal("非 owner Release 不应影响真实 owner 续约")
	}
}
