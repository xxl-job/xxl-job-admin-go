package registry

import "testing"

func TestMemoryRegistry(t *testing.T) {
	r := NewMemory()
	r.OnConnect("app", "id1", &ExecutorConn{ID: "id1", AppName: "app", Address: "addr1"})
	r.OnConnect("app", "id2", &ExecutorConn{ID: "id2", AppName: "app", Address: "addr2"})

	if got := r.Addresses("app"); len(got) != 2 {
		t.Fatalf("addresses: got %d", len(got))
	}
	if got := r.Conns("app"); len(got) != 2 {
		t.Fatalf("conns: got %d", len(got))
	}

	r.OnDisconnect("app", "id1")
	if got := r.Addresses("app"); len(got) != 1 || got[0] != "addr2" {
		t.Fatalf("after disconnect: %+v", got)
	}

	r.UpdateLoad("app", "id2", LoadReport{ActiveTasks: 5})
	c := r.Conns("app")[0]
	if lp := c.Load(); lp == nil || lp.ActiveTasks != 5 {
		t.Fatalf("load not updated: %+v", lp)
	}
}

// TestConnsByCluster 验证流量泳道：按 cluster 标签过滤执行器连接。
func TestConnsByCluster(t *testing.T) {
	r := NewMemory()
	r.OnConnect("app", "g1", &ExecutorConn{ID: "g1", AppName: "app", Address: "a1", Cluster: "gray"})
	r.OnConnect("app", "g2", &ExecutorConn{ID: "g2", AppName: "app", Address: "a2", Cluster: "gray"})
	r.OnConnect("app", "p1", &ExecutorConn{ID: "p1", AppName: "app", Address: "a3", Cluster: "prod"})

	if got := r.Conns("app"); len(got) != 3 {
		t.Fatalf("Conns all: got %d", len(got))
	}
	if got := r.ConnsByCluster("app", "gray"); len(got) != 2 {
		t.Fatalf("ConnsByCluster gray: got %d", len(got))
	}
	if got := r.ConnsByCluster("app", "prod"); len(got) != 1 {
		t.Fatalf("ConnsByCluster prod: got %d", len(got))
	}
	if got := r.ConnsByCluster("app", "nonexist"); len(got) != 0 {
		t.Fatalf("ConnsByCluster nonexist: got %d", len(got))
	}

	// 下线一个 gray 后，gray 仅剩 1
	r.OnDisconnect("app", "g1")
	if got := r.ConnsByCluster("app", "gray"); len(got) != 1 {
		t.Fatalf("after disconnect gray: got %d", len(got))
	}
}
