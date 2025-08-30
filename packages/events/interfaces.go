package events

import "context"

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// Publish publishes an event synchronously
	Publish(ctx context.Context, event interface{}) error

	// PublishAsync publishes an event asynchronously
	PublishAsync(ctx context.Context, event interface{}) error

	// PublishBatch publishes multiple events in a batch
	PublishBatch(ctx context.Context, events []interface{}) error

	// Close closes the event publisher and cleans up resources
	Close() error
}

// Event represents a domain event
type Event interface {
	EventType() string
	EventData() interface{}
}
