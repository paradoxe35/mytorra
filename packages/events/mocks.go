package events

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockEventPublisher is a mock implementation of the EventPublisher interface
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishAsync(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBatch(ctx context.Context, events []interface{}) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockEvent is a mock implementation of the Event interface
type MockEvent struct {
	mock.Mock
}

func (m *MockEvent) EventType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEvent) EventData() interface{} {
	args := m.Called()
	return args.Get(0)
}
