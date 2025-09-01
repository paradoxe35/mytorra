package logger

import (
	"context"
	"log/slog"
	"time"
)

// SlogAdapter adapts Go's standard slog.Logger to implement our Interface
// This allows us to use slog directly while maintaining interface compatibility
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new adapter for slog.Logger
func NewSlogAdapter(logger *slog.Logger) Interface {
	return &SlogAdapter{
		logger: logger,
	}
}

// NewSlogAdapterWithOptions creates a new slog adapter with custom options
func NewSlogAdapterWithOptions(serviceName, environment string, handler slog.Handler) Interface {
	logger := slog.New(handler).With(
		slog.String("service", serviceName),
		slog.String("environment", environment),
		slog.String("version", getVersion()),
	)

	return &SlogAdapter{
		logger: logger,
	}
}

func (a *SlogAdapter) Debug(msg string, fields ...any) {
	a.logger.Debug(msg, fields...)
}

func (a *SlogAdapter) Info(msg string, fields ...any) {
	a.logger.Info(msg, fields...)
}

func (a *SlogAdapter) Warn(msg string, fields ...any) {
	a.logger.Warn(msg, fields...)
}

func (a *SlogAdapter) Error(msg string, fields ...any) {
	a.logger.Error(msg, fields...)
}

func (a *SlogAdapter) Fatal(msg string, fields ...any) {
	a.logger.Error(msg, fields...)
	// Note: slog doesn't have Fatal, so we just log as Error
	// Fatal behavior (os.Exit) is handled by the caller if needed
}

func (a *SlogAdapter) WithContext(ctx context.Context) Interface {
	if requestID := ctx.Value("request_id"); requestID != nil {
		return &SlogAdapter{
			logger: a.logger.With(slog.String("request_id", requestID.(string))),
		}
	}
	return a
}

func (a *SlogAdapter) WithField(key string, value any) Interface {
	return &SlogAdapter{
		logger: a.logger.With(slog.Any(key, value)),
	}
}

func (a *SlogAdapter) WithFields(fields map[string]any) Interface {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &SlogAdapter{
		logger: a.logger.With(attrs...),
	}
}

func (a *SlogAdapter) WithError(err error) Interface {
	return &SlogAdapter{
		logger: a.logger.With(slog.String("error", err.Error())),
	}
}

func (a *SlogAdapter) LogRequest(method, path string, statusCode int, duration time.Duration) {
	a.logger.Info("http request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	)
}

func (a *SlogAdapter) LogPanic(r any) {
	a.logger.Error("panic recovered",
		slog.Any("panic", r),
	)
}

// Ensure SlogAdapter implements Interface
var _ Interface = (*SlogAdapter)(nil)
