package ports

import (
	"context"
)

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
	PublishAsync(ctx context.Context, event interface{}) error
	PublishBatch(ctx context.Context, events []interface{}) error
}

// EventConsumer defines the interface for consuming domain events
type EventConsumer interface {
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	Unsubscribe(ctx context.Context, topic string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// EventHandler defines the function signature for event handlers
type EventHandler func(ctx context.Context, event []byte) error
