package lock

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MySQLLeaseLocker 基于 scheduler_lease 表的租约锁（leader 选举）。
//
// 抢锁/续约（原子）：
//
//	UPDATE scheduler_lease
//	SET owner=:node, lease_until=NOW()+TTL
//	WHERE id=1 AND (lease_until < NOW() OR owner=:node)
//
// leader 停止续约 → TTL 后 lease 过期 → 备节点抢到。会话无关，连接池安全。
type MySQLLeaseLocker struct {
	db     *gorm.DB
	nodeID string
	ttl    time.Duration
}

// NewMySQLLeaseLocker 创建租约锁。
func NewMySQLLeaseLocker(db *gorm.DB, nodeID string, ttl time.Duration) *MySQLLeaseLocker {
	return &MySQLLeaseLocker{db: db, nodeID: nodeID, ttl: ttl}
}

// TryAcquire 抢租约。
func (l *MySQLLeaseLocker) TryAcquire(ctx context.Context, key string) (bool, error) {
	return l.update(ctx)
}

// Renew 续约（与 TryAcquire 同语义：仅 owner 匹配或已过期时成功）。
func (l *MySQLLeaseLocker) Renew(ctx context.Context, key string) (bool, error) {
	return l.update(ctx)
}

// Release 主动释放（仅当 owner 是自己）。
func (l *MySQLLeaseLocker) Release(ctx context.Context, key string) error {
	return l.db.WithContext(ctx).Exec(
		"UPDATE scheduler_lease SET owner='', lease_until='1970-01-01 00:00:00' WHERE id=1 AND owner=?",
		l.nodeID,
	).Error
}

func (l *MySQLLeaseLocker) update(ctx context.Context) (bool, error) {
	ttlSec := int(l.ttl.Seconds())
	res := l.db.WithContext(ctx).Exec(
		"UPDATE scheduler_lease SET owner=?, lease_until=DATE_ADD(NOW(3), INTERVAL ? SECOND) WHERE id=1 AND (lease_until < NOW(3) OR owner=?)",
		l.nodeID, ttlSec, l.nodeID,
	)
	if res.Error != nil {
		return false, fmt.Errorf("lease update: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}
