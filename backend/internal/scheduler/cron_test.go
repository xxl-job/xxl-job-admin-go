package scheduler

import (
	"testing"
	"time"
)

func TestNextTriggerTimeFixRate(t *testing.T) {
	base := time.Unix(1000, 0)
	next, err := NextTriggerTime(ScheduleFixRate, "5", base)
	if err != nil {
		t.Fatalf("fix rate: %v", err)
	}
	want := base.Add(5 * time.Second).UnixMilli()
	if next != want {
		t.Fatalf("fix rate next = %d, want %d", next, want)
	}
}

func TestNextTriggerTimeCron(t *testing.T) {
	next, err := NextTriggerTime(ScheduleCron, "0 * * * * ?", time.Now())
	if err != nil {
		t.Fatalf("cron parse: %v", err)
	}
	if next <= 0 {
		t.Fatal("next should be positive")
	}
}

func TestNextTriggerTimeInvalid(t *testing.T) {
	if _, err := NextTriggerTime(ScheduleCron, "not-a-cron", time.Now()); err == nil {
		t.Fatal("expected error for invalid cron")
	}
	if _, err := NextTriggerTime(ScheduleFixRate, "0", time.Now()); err == nil {
		t.Fatal("expected error for zero interval")
	}
	if _, err := NextTriggerTime(99, "x", time.Now()); err == nil {
		t.Fatal("expected error for unknown type")
	}
}
