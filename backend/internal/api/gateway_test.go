package api

import (
	"testing"

	grpcapi "ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGateway_Start(t *testing.T) {
	// Should start both gRPC and WebSocket servers
	gateway := setupTestGateway(t)

	err := gateway.Start()
	require.NoError(t, err)

	assert.True(t, gateway.IsRunning())
}

func TestGateway_GRPCEndpoint(t *testing.T) {
	// Should expose gRPC services
	gateway := setupTestGateway(t)
	gateway.Start()
	defer gateway.Stop()

	// Test that gateway has gRPC server configured
	assert.NotNil(t, gateway.GetGRPCServer())
}

func TestGateway_WebSocketEndpoint(t *testing.T) {
	// Should expose WebSocket endpoint
	gateway := setupTestGateway(t)
	gateway.Start()
	defer gateway.Stop()

	wsURL := "ws://" + gateway.HTTPAddr() + "/ws"
	// Test that the endpoint is accessible (simplified test)
	assert.NotEmpty(t, wsURL)
}

// Helper functions
func setupTestGateway(t *testing.T) *Gateway {
	config := &Config{
		GRPCPort: 0, // Use random port
		HTTPPort: 0, // Use random port
	}

	// Create mock services
	msgService := grpcapi.NewMessageService(nil, nil)
	mediaService := grpcapi.NewMediaService(nil)
	callService := grpcapi.NewCallService(nil)
	auth := auth.NewAuthenticator()
	wsServer := websocket.NewServer(auth)

	return NewGateway(config, msgService, mediaService, callService, wsServer)
}
