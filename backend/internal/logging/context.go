package logging

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return NewLogger(&Config{Level: InfoLevel})
}

func WithRequestID(ctx context.Context) context.Context {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)

	logger := FromContext(ctx)
	logger = logger.WithFields(String("request_id", requestID))

	return WithLogger(ctx, logger)
}
