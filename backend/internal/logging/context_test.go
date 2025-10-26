package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ledabeer/backend/internal/logging"

	"github.com/stretchr/testify/assert"
)

func TestLogger_WithContext(t *testing.T) {
	// Should attach context fields to all logs
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	ctxLogger := logger.WithFields(
		logging.String("request_id", "req-123"),
		logging.String("peer_id", "peer-456"),
	)

	ctxLogger.Info("operation started")
	ctxLogger.Info("operation completed")

	lines := bytes.Split(buf.Bytes(), []byte("\n"))

	// Both logs should have context fields
	var log1, log2 map[string]interface{}
	json.Unmarshal(lines[0], &log1)
	json.Unmarshal(lines[1], &log2)

	assert.Equal(t, "req-123", log1["request_id"])
	assert.Equal(t, "peer-456", log1["peer_id"])
	assert.Equal(t, "req-123", log2["request_id"])
	assert.Equal(t, "peer-456", log2["peer_id"])
}

func TestLogger_FromContext(t *testing.T) {
	// Should extract logger from context
	ctx := context.Background()
	logger := logging.NewLogger(&logging.Config{
		Level: logging.InfoLevel,
	})

	ctx = logging.WithLogger(ctx, logger)

	extracted := logging.FromContext(ctx)
	assert.NotNil(t, extracted)
}

func TestLogger_RequestID(t *testing.T) {
	// Should automatically add request ID
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	ctx := logging.WithLogger(context.Background(), logger)
	ctx = logging.WithRequestID(ctx)
	ctxLogger := logging.FromContext(ctx)

	ctxLogger.Info("test message")

	var logEntry map[string]interface{}
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.NotEmpty(t, logEntry["request_id"])
}
