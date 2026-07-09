package trigger

import "testing"

func TestBlockTrackerDiscardLater(t *testing.T) {
	b := NewBlockTracker()
	b.OnStart(1, 100)
	allow, cancel := b.Allow(1, BlockDiscardLater)
	if allow {
		t.Fatal("discard-later should block when running")
	}
	if len(cancel) != 0 {
		t.Fatal("discard-later should not return cancel ids")
	}
	b.OnDone(1, 100)
	allow2, _ := b.Allow(1, BlockDiscardLater)
	if !allow2 {
		t.Fatal("should allow after done")
	}
}

func TestBlockTrackerCoverEarly(t *testing.T) {
	b := NewBlockTracker()
	b.OnStart(1, 100)
	b.OnStart(1, 101)
	allow, cancel := b.Allow(1, BlockCoverEarly)
	if !allow {
		t.Fatal("cover-early should allow")
	}
	if len(cancel) != 2 {
		t.Fatalf("cover-early should cancel 2, got %d", len(cancel))
	}
}

func TestBlockTrackerSerial(t *testing.T) {
	b := NewBlockTracker()
	b.OnStart(1, 100)
	allow, _ := b.Allow(1, BlockSerial)
	if !allow {
		t.Fatal("serial should allow (executor queues)")
	}
}
