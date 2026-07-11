package logstream

import "testing"

// TestHub_SubscribePublish 订阅后能收到推送。
func TestHub_SubscribePublish(t *testing.T) {
	h := New()
	ch := h.Subscribe(1)
	defer h.Unsubscribe(1, ch)

	h.Publish(1, []byte("hello"))
	select {
	case got := <-ch:
		if string(got) != "hello" {
			t.Errorf("收到 %q，want hello", got)
		}
	default:
		t.Fatal("未收到推送")
	}
}

// TestHub_MultipleSubscribers 同 logID 多订阅者都收到。
func TestHub_MultipleSubscribers(t *testing.T) {
	h := New()
	a := h.Subscribe(7)
	b := h.Subscribe(7)
	defer h.Unsubscribe(7, a)
	defer h.Unsubscribe(7, b)

	h.Publish(7, []byte("msg"))
	for i, ch := range []chan []byte{a, b} {
		select {
		case <-ch:
		default:
			t.Fatalf("订阅者 %d 未收到", i)
		}
	}
}

// TestHub_Isolation 不同 logID 互不干扰。
func TestHub_Isolation(t *testing.T) {
	h := New()
	a := h.Subscribe(1)
	b := h.Subscribe(2)
	defer h.Unsubscribe(1, a)
	defer h.Unsubscribe(2, b)

	h.Publish(1, []byte("only-a"))
	select {
	case <-b:
		t.Fatal("logID 2 不应收到 logID 1 的推送")
	default:
	}
}

// TestHub_Unsubscribe 取消后 channel 关闭、再发布不 panic。
func TestHub_Unsubscribe(t *testing.T) {
	h := New()
	ch := h.Subscribe(3)
	h.Unsubscribe(3, ch)

	if _, ok := <-ch; ok {
		t.Fatal("Unsubscribe 后 channel 应已关闭")
	}
	h.Publish(3, []byte("x")) // 不应 panic
}

// TestHub_PublishNoSubscriber 无订阅者时 Publish 不阻塞。
func TestHub_PublishNoSubscriber(t *testing.T) {
	h := New()
	h.Publish(999, []byte("orphan"))
}

// TestHub_BufferFullDrops 缓冲(32)满后丢弃，不阻塞。
func TestHub_BufferFullDrops(t *testing.T) {
	h := New()
	ch := h.Subscribe(5)
	defer h.Unsubscribe(5, ch)

	for i := 0; i < 40; i++ {
		h.Publish(5, []byte("x"))
	}
	n := 0
	for {
		select {
		case <-ch:
			n++
		default:
			return
		}
	}
}
