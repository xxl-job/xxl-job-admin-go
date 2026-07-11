package circuit

import (
	"testing"
	"time"
)

// TestBreaker_AllowNormal 无失败时放行。
func TestBreaker_AllowNormal(t *testing.T) {
	b := New(3, time.Second)
	if !b.Allow("app") {
		t.Fatal("无失败应 Allow")
	}
}

// TestBreaker_TripOnFailures 连续失败达阈值触发熔断。
func TestBreaker_TripOnFailures(t *testing.T) {
	b := New(3, time.Second)
	b.OnResult("app", false)
	b.OnResult("app", false)
	if !b.Allow("app") {
		t.Fatal("2 次失败未达阈值 3，应仍 Allow")
	}
	b.OnResult("app", false) // 第 3 次 → open
	if b.Allow("app") {
		t.Fatal("3 次失败应熔断")
	}
}

// TestBreaker_RecoverAfterCooldown open 过冷却后 half-open 放行，探测成功 closed。
func TestBreaker_RecoverAfterCooldown(t *testing.T) {
	b := New(2, 100*time.Millisecond)
	b.OnResult("app", false)
	b.OnResult("app", false) // open
	if !b.IsOpen("app") {
		t.Fatal("应处于 open")
	}
	time.Sleep(150 * time.Millisecond)
	if !b.Allow("app") {
		t.Fatal("冷却后应 half-open 放行探测")
	}
	b.OnResult("app", true) // 探测成功 → closed
	if b.IsOpen("app") {
		t.Fatal("探测成功应 closed")
	}
}

// TestBreaker_HalfOpenFailReopen half-open 探测失败 → 重新 open。
func TestBreaker_HalfOpenFailReopen(t *testing.T) {
	b := New(2, 100*time.Millisecond)
	b.OnResult("app", false)
	b.OnResult("app", false) // open
	time.Sleep(150 * time.Millisecond)
	_ = b.Allow("app")       // half-open
	b.OnResult("app", false) // 探测失败 → reopen
	if !b.IsOpen("app") {
		t.Fatal("探测失败应重新 open")
	}
}

// TestBreaker_SuccessResets 成功重置连续失败计数。
func TestBreaker_SuccessResets(t *testing.T) {
	b := New(3, time.Second)
	b.OnResult("app", false)
	b.OnResult("app", false)
	b.OnResult("app", true) // 重置
	b.OnResult("app", false)
	b.OnResult("app", false) // 重置后仅 2 次
	if !b.Allow("app") {
		t.Fatal("重置后 2 次未达阈值 3，应 Allow")
	}
}

// TestBreaker_Isolation 不同 app 互不影响。
func TestBreaker_Isolation(t *testing.T) {
	b := New(2, time.Second)
	b.OnResult("a", false)
	b.OnResult("a", false) // a 熔断
	if !b.Allow("b") {
		t.Fatal("app a 熔断不影响 app b")
	}
}
