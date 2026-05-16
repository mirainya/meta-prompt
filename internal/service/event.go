package service

import "sync"

type Event struct {
	Step     int    `json:"step"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Progress string `json:"progress,omitempty"`
	Error    string `json:"error,omitempty"`
}

type EventBus struct {
	mu   sync.RWMutex
	subs map[int64][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{subs: make(map[int64][]chan Event)}
}

func (b *EventBus) Subscribe(historyID int64) chan Event {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.subs[historyID] = append(b.subs[historyID], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBus) Unsubscribe(historyID int64, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[historyID]
	for i, s := range subs {
		if s == ch {
			b.subs[historyID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(b.subs[historyID]) == 0 {
		delete(b.subs, historyID)
	}
	close(ch)
}

func (b *EventBus) Publish(historyID int64, evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs[historyID] {
		select {
		case ch <- evt:
		default:
		}
	}
}
