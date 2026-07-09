package trigger

import (
	"testing"

	"xxl-job-admin-go/internal/model"
	"xxl-job-admin-go/internal/registry"
)

func mkConns(n int) []*registry.ExecutorConn {
	conns := make([]*registry.ExecutorConn, n)
	for i := range conns {
		id := string(rune('a' + i))
		conns[i] = &registry.ExecutorConn{ID: id, Address: "addr" + id}
	}
	return conns
}

func TestFirstLastRouter(t *testing.T) {
	conns := mkConns(3)
	if got := (&firstRouter{}).Route(&model.Job{}, conns); len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("first: %+v", got)
	}
	if got := (&lastRouter{}).Route(&model.Job{}, conns); len(got) != 1 || got[0].ID != "c" {
		t.Fatalf("last: %+v", got)
	}
	if got := (&firstRouter{}).Route(&model.Job{}, nil); got != nil {
		t.Fatalf("first empty should be nil: %+v", got)
	}
}

func TestBroadcastRouter(t *testing.T) {
	conns := mkConns(3)
	got := (&broadcastRouter{}).Route(&model.Job{}, conns)
	if len(got) != 3 {
		t.Fatalf("broadcast: got %d", len(got))
	}
}

func TestRoutersRegistry(t *testing.T) {
	rs := NewRouters()
	if rs.Get(RouteFirst) == nil {
		t.Fatal("RouteFirst missing")
	}
	if rs.Get(RouteShardingBroadcast) == nil {
		t.Fatal("RouteShardingBroadcast missing")
	}
	if rs.Get(99) != nil {
		t.Fatal("unknown route should be nil")
	}
}
