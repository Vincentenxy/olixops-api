// Package event 提供领域事件总线抽象。
//
// 初期使用内存实现,后续可替换为 NATS / Kafka / RabbitMQ 而不影响业务代码。
package event

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Event 是一条领域事件。
type Event struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`   // 如 application.deployed
	Source     string         `json:"source"` // 触发来源
	OccurredAt time.Time      `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}

// Handler 是订阅者。
type Handler func(ctx context.Context, ev Event) error

// Bus 是事件总线。
type Bus interface {
	Publish(ctx context.Context, ev Event) error
	Subscribe(eventType string, handler Handler)
	Close() error
}

// InMemoryBus 是最简单的进程内实现,适合单体阶段。
type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	log      *zap.Logger
}

// NewInMemoryBus 构造内存总线。
func NewInMemoryBus(log *zap.Logger) *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[string][]Handler),
		log:      log,
	}
}

// Subscribe 订阅事件。
func (b *InMemoryBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Publish 同步分发,subscriber 失败仅记录日志,不影响发布方。
func (b *InMemoryBus) Publish(ctx context.Context, ev Event) error {
	b.mu.RLock()
	hs := append([]Handler(nil), b.handlers[ev.Type]...)
	b.mu.RUnlock()

	for _, h := range hs {
		if err := h(ctx, ev); err != nil && b.log != nil {
			b.log.Warn("event handler failed",
				zap.String("event_type", ev.Type),
				zap.String("event_id", ev.ID),
				zap.Error(err),
			)
		}
	}
	return nil
}

// Close 内存实现无需关闭。
func (b *InMemoryBus) Close() error { return nil }
