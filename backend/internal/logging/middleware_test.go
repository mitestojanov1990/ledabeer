package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledabeer/backend/internal/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMiddleware_gRPCLogging(t *testing.T) {
	// Should log all gRPC requests
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	interceptor := logging.UnaryServerInterceptor(logger)

	// Call mock gRPC handler
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/ledabeer.MessageService/SendMessage",
	}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	})

	require.NoError(t, err)

	var logEntry map[string]interface{}
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.Contains(t, logEntry, "grpc.method")
	assert.Contains(t, logEntry, "grpc.duration_ms")
}

func TestMiddleware_WebSocketLogging(t *testing.T) {
	// Should log WebSocket connections and messages
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
	})

	handler := logging.WebSocketHandler(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock WebSocket handler
	}))

	// Simulate WebSocket connection
	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	assert.Contains(t, output, "websocket_connection")
}

func TestMiddleware_ErrorLogging(t *testing.T) {
	// Should automatically log errors
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.ErrorLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	interceptor := logging.UnaryServerInterceptor(logger)

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/ledabeer.MessageService/SendMessage",
	}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.Internal, "test error")
	})

	assert.Error(t, err)

	var logEntry map[string]interface{}
	json.Unmarshal(buf.Bytes(), &logEntry)

	assert.Equal(t, "ERROR", logEntry["level"])
	assert.Contains(t, logEntry, "error")
}
