// Package logstream 实时日志推送 Hub。
// 执行器回传 LogChunk → Hub.Publish → 前端 WebSocket 订阅者接收。
package logstream

import "sync"

// Hub 按 logID 维护订阅者，发布日志块。
type Hub struct {
	mu   sync.RWMutex
	subs map[int64][]chan []byte // logID → 订阅 channel
}

// New 创建 Hub。
func New() *Hub {
	return &Hub{subs: map[int64][]chan []byte{}}
}

// Subscribe 订阅指定 logID 的日志流。返回接收 channel。
func (h *Hub) Subscribe(logID int64) chan []byte {
	ch := make(chan []byte, 32)
	h.mu.Lock()
	h.subs[logID] = append(h.subs[logID], ch)
	h.mu.Unlock()
	return ch
}

// Unsubscribe 取消订阅并关闭 channel。
func (h *Hub) Unsubscribe(logID int64, ch chan []byte) {
	h.mu.Lock()
	subs := h.subs[logID]
	for i, c := range subs {
		if c == ch {
			h.subs[logID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(h.subs[logID]) == 0 {
		delete(h.subs, logID)
	}
	h.mu.Unlock()
	close(ch)
}

// Publish 向指定 logID 的所有订阅者推送日志字节。满则丢弃（实时日志不阻塞）。
func (h *Hub) Publish(logID int64, data []byte) {
	h.mu.RLock()
	subs := h.subs[logID]
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- data:
		default:
		}
	}
}
