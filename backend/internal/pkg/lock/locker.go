// Package lock 提供分布式锁抽象，默认 MySQL 租约锁实现。
package lock

import "context"

// Locker 分布式锁接口。
// 默认实现 MySQLLeaseLocker（无 Redis 依赖）；未来可加 RedisLocker。
type Locker interface {
	// TryAcquire 尝试获取 key 的锁，成功返回 true。
	TryAcquire(ctx context.Context, key string) (bool, error)
	// Renew 续约当前持有的 key，租约未过期返回 true。
	Renew(ctx context.Context, key string) (bool, error)
	// Release 主动释放 key（仅当 owner 是自己）。
	Release(ctx context.Context, key string) error
}
