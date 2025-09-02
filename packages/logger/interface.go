package logger

import (
	"context"
	"time"
)

// Fields is a type alias for map[string]any for logging fields
type Fields = map[string]any

// Interface defines the logger interface for application services
type Interface interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)

	WithContext(ctx context.Context) Interface
	WithField(key string, value any) Interface
	WithFields(fields map[string]any) Interface
	WithError(err error) Interface

	// Additional methods for HTTP and panic logging
	LogRequest(method, path string, statusCode int, duration time.Duration)
	LogPanic(r any)
}

// Ensure Logger implements Interface
var _ Interface = (*Logger)(nil)
