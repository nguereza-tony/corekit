package common

import (
	"context"
	"sync"
)

type Event struct {
	Name string
	Data map[string]interface{}
}

type EventHandler func(ctx context.Context, event Event) error

type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

var globalBus = &EventBus{
	handlers: make(map[string][]EventHandler),
}

func GetEventBus() *EventBus {
	return globalBus
}

func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *EventBus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := b.handlers[event.Name]
	b.mu.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			_ = h(ctx, event)
		}(handler)
	}
}
