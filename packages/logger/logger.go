package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	*slog.Logger
}

func New(serviceName, environment string) *Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     getLogLevel(environment),
		AddSource: environment == "development",
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", time.Now().UTC().Format(time.RFC3339))
			}
			return a
		},
	}

	if environment == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
		slog.String("environment", environment),
		slog.String("version", getVersion()),
	)

	return &Logger{logger}
}

func getLogLevel(environment string) slog.Level {
	switch environment {
	case "production":
		return slog.LevelInfo
	case "staging":
		return slog.LevelDebug
	default:
		return slog.LevelDebug
	}
}

func getVersion() string {
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}
	return "unknown"
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if requestID := ctx.Value("request_id"); requestID != nil {
		return &Logger{l.With(slog.String("request_id", requestID.(string)))}
	}
	return l
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.With(slog.String("error", err.Error()))}
}

func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{l.With(slog.Any(key, value))}
}

func (l *Logger) WithFields(fields map[string]any) *Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &Logger{l.With(attrs...)}
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

func (l *Logger) LogRequest(method, path string, statusCode int, duration time.Duration) {
	l.Info("http request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	)
}

func (l *Logger) LogPanic(r interface{}) {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]

	l.Error("panic recovered",
		slog.Any("panic", r),
		slog.String("stack", string(buf)),
	)
}
