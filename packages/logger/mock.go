package logger

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of the Interface for testing
type MockLogger struct {
	mock.Mock
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) WithContext(ctx context.Context) Interface {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return m // Return self if no mock return value is set
	}
	return args.Get(0).(Interface)
}

func (m *MockLogger) WithField(key string, value any) Interface {
	args := m.Called(key, value)
	if args.Get(0) == nil {
		return m // Return self if no mock return value is set
	}
	return args.Get(0).(Interface)
}

func (m *MockLogger) WithFields(fields map[string]any) Interface {
	args := m.Called(fields)
	if args.Get(0) == nil {
		return m // Return self if no mock return value is set
	}
	return args.Get(0).(Interface)
}

func (m *MockLogger) WithError(err error) Interface {
	args := m.Called(err)
	if args.Get(0) == nil {
		return m // Return self if no mock return value is set
	}
	return args.Get(0).(Interface)
}

func (m *MockLogger) LogRequest(method, path string, statusCode int, duration time.Duration) {
	m.Called(method, path, statusCode, duration)
}

func (m *MockLogger) LogPanic(r any) {
	m.Called(r)
}

// Ensure MockLogger implements Interface
var _ Interface = (*MockLogger)(nil)
