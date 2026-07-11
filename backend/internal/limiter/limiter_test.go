package limiter

import "testing"

// TestLimiter_AllowWithinBurst burst 内放行。
func TestLimiter_AllowWithinBurst(t *testing.T) {
	l := New(1, 3) // rps=1, burst=3
	for i := 0; i < 3; i++ {
		if !l.Allow("app") {
			t.Fatalf("burst 内第 %d 次应 Allow", i+1)
		}
	}
}

// TestLimiter_BlockOverRate 超 burst 拒绝。
func TestLimiter_BlockOverRate(t *testing.T) {
	l := New(1, 2)
	l.Allow("app")
	l.Allow("app") // 用尽 burst
	if l.Allow("app") {
		t.Fatal("超 burst 应拒绝")
	}
}

// TestLimiter_Isolation 不同 app 独立令牌桶。
func TestLimiter_Isolation(t *testing.T) {
	l := New(1, 1)
	l.Allow("a") // a 用尽
	if l.Allow("a") {
		t.Fatal("a 超限应拒")
	}
	if !l.Allow("b") {
		t.Fatal("a 限流不影响 b")
	}
}

// TestLimiter_Defaults rps/burst<=0 用默认值（不 panic）。
func TestLimiter_Defaults(t *testing.T) {
	l := New(0, 0)
	if !l.Allow("app") {
		t.Fatal("默认参数应 Allow")
	}
}
