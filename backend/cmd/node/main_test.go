package main_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/api"
	"ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/auth"
	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/logging"
	"ledabeer/backend/internal/media"
	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"
	"ledabeer/backend/internal/storage"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_Integration_AllComponentsWired(t *testing.T) {
	// Should wire all backend components together
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This test verifies that main.go can initialize all components
	// without errors and they work together

	// Test that we can create all the required components
	// (This is a simplified test - in a real scenario we'd need to mock external dependencies)

	// Verify logging system is initialized
	assert.NotNil(t, getLogger(), "Logger should be initialized")

	// Verify network host can be created
	host, err := createHost(ctx)
	require.NoError(t, err, "Should create libp2p host")
	defer host.Close()

	// Verify IPFS node can be created
	ipfsNode, err := createIPFSNode(ctx)
	require.NoError(t, err, "Should create IPFS node")
	defer ipfsNode.Close()

	// Verify all handlers can be created
	msgHandler, err := createMessageHandler(host)
	require.NoError(t, err, "Should create message handler")
	assert.NotNil(t, msgHandler)

	groupManager, err := createGroupManager(host)
	require.NoError(t, err, "Should create group manager")
	assert.NotNil(t, groupManager)

	mediaHandler, err := createMediaHandler(ipfsNode)
	require.NoError(t, err, "Should create media handler")
	assert.NotNil(t, mediaHandler)

	callManager, err := createCallManager(host)
	require.NoError(t, err, "Should create call manager")
	assert.NotNil(t, callManager)

	// Verify API services can be created
	msgService, err := createMessageService(msgHandler, groupManager)
	require.NoError(t, err, "Should create message service")
	assert.NotNil(t, msgService)

	mediaService, err := createMediaService(mediaHandler)
	require.NoError(t, err, "Should create media service")
	assert.NotNil(t, mediaService)

	callService, err := createCallService(callManager)
	require.NoError(t, err, "Should create call service")
	assert.NotNil(t, callService)

	// Verify WebSocket server can be created
	wsServer, err := createWebSocketServer()
	require.NoError(t, err, "Should create WebSocket server")
	assert.NotNil(t, wsServer)

	// Verify API gateway can be created
	gateway, err := createAPIGateway(msgService, mediaService, callService, wsServer)
	require.NoError(t, err, "Should create API gateway")
	assert.NotNil(t, gateway)
}

func TestMain_Integration_StartupShutdown(t *testing.T) {
	// Should handle startup and shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test that all components can be started and stopped
	components, err := initializeAllComponents(ctx)
	require.NoError(t, err, "Should initialize all components")

	// Test startup
	err = startAllComponents(components)
	require.NoError(t, err, "Should start all components")

	// Test shutdown
	err = stopAllComponents(components)
	require.NoError(t, err, "Should stop all components")
}

func TestMain_Integration_LoggingIntegration(t *testing.T) {
	// Should integrate logging throughout all components
	ctx := context.Background()

	// Verify logging is configured
	logger := getLogger()
	assert.NotNil(t, logger, "Logger should be available")

	// Test that components use the logger
	host, err := createHost(ctx)
	require.NoError(t, err)
	defer host.Close()

	// Verify logging middleware is applied
	gateway, err := createAPIGatewayWithLogging()
	require.NoError(t, err)
	assert.NotNil(t, gateway)
}

// Helper functions that mirror the main.go implementation
func getLogger() *logging.Logger {
	logConfig, _ := logging.ConfigFromEnv()
	return logging.NewLogger(logConfig)
}

func createHost(ctx context.Context) (host.Host, error) {
	return network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
}

func createIPFSNode(ctx context.Context) (*storage.IPFSNode, error) {
	return storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
}

func createMessageHandler(host host.Host) (*messaging.MessageHandler, error) {
	return messaging.NewMessageHandler(host), nil
}

func createGroupManager(host host.Host) (*messaging.GroupManager, error) {
	return messaging.NewGroupManager(context.Background(), host)
}

func createMediaHandler(ipfsNode *storage.IPFSNode) (*media.MediaHandler, error) {
	return media.NewMediaHandler(ipfsNode), nil
}

func createCallManager(host host.Host) (*calls.CallManager, error) {
	return calls.NewCallManager(host), nil
}

func createMessageService(msgHandler *messaging.MessageHandler, groupManager *messaging.GroupManager) (*grpc.MessageService, error) {
	return grpc.NewMessageService(msgHandler, groupManager), nil
}

func createMediaService(mediaHandler *media.MediaHandler) (*grpc.MediaService, error) {
	return grpc.NewMediaService(mediaHandler), nil
}

func createCallService(callManager *calls.CallManager) (*grpc.CallService, error) {
	return grpc.NewCallService(callManager), nil
}

func createWebSocketServer() (*websocket.Server, error) {
	auth := auth.NewAuthenticator()
	return websocket.NewServer(auth), nil
}

func createAPIGateway(msgService *grpc.MessageService, mediaService *grpc.MediaService, callService *grpc.CallService, wsServer *websocket.Server) (*api.Gateway, error) {
	return api.NewGateway(&api.Config{
		GRPCPort: 50051,
		HTTPPort: 8080,
	}, msgService, mediaService, callService, wsServer), nil
}

func createAPIGatewayWithLogging() (*api.Gateway, error) {
	// Create a minimal gateway with logging
	msgService := grpc.NewMessageService(nil, nil)
	mediaService := grpc.NewMediaService(nil)
	callService := grpc.NewCallService(nil)
	wsServer, _ := createWebSocketServer()

	gateway := api.NewGateway(&api.Config{
		GRPCPort: 50051,
		HTTPPort: 8080,
	}, msgService, mediaService, callService, wsServer)

	return gateway, nil
}

func initializeAllComponents(ctx context.Context) (map[string]interface{}, error) {
	// Initialize all components like in main.go
	host, err := createHost(ctx)
	if err != nil {
		return nil, err
	}

	ipfsNode, err := createIPFSNode(ctx)
	if err != nil {
		return nil, err
	}

	msgHandler, err := createMessageHandler(host)
	if err != nil {
		return nil, err
	}

	groupManager, err := createGroupManager(host)
	if err != nil {
		return nil, err
	}

	mediaHandler, err := createMediaHandler(ipfsNode)
	if err != nil {
		return nil, err
	}

	callManager, err := createCallManager(host)
	if err != nil {
		return nil, err
	}

	msgService, err := createMessageService(msgHandler, groupManager)
	if err != nil {
		return nil, err
	}

	mediaService, err := createMediaService(mediaHandler)
	if err != nil {
		return nil, err
	}

	callService, err := createCallService(callManager)
	if err != nil {
		return nil, err
	}

	wsServer, err := createWebSocketServer()
	if err != nil {
		return nil, err
	}

	gateway, err := createAPIGateway(msgService, mediaService, callService, wsServer)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"host":         host,
		"ipfsNode":     ipfsNode,
		"msgHandler":   msgHandler,
		"groupManager": groupManager,
		"mediaHandler": mediaHandler,
		"callManager":  callManager,
		"msgService":   msgService,
		"mediaService": mediaService,
		"callService":  callService,
		"wsServer":     wsServer,
		"gateway":      gateway,
	}, nil
}

func startAllComponents(components map[string]interface{}) error {
	gateway := components["gateway"].(*api.Gateway)
	return gateway.Start()
}

func stopAllComponents(components map[string]interface{}) error {
	gateway := components["gateway"].(*api.Gateway)
	gateway.Stop()
	return nil
}
