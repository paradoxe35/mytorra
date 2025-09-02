package ports

import (
	"context"

	"github.com/paradoxe35/torra/packages/events"
)

// EventPublisher is an alias to the shared events interface
// This allows the application layer to use the standardized events interface
type EventPublisher = events.EventPublisher

// EventBus is an alias for EventPublisher for backward compatibility
type EventBus = events.EventPublisher

// EventConsumer defines the interface for consuming domain events
type EventConsumer interface {
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	Unsubscribe(ctx context.Context, topic string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// EventHandler defines the function signature for event handlers
type EventHandler func(ctx context.Context, event []byte) error
