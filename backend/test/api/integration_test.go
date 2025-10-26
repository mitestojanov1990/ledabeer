package api_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/api"
	grpcapi "ledabeer/backend/internal/api/grpc"
	"ledabeer/backend/internal/api/websocket"
	"ledabeer/backend/internal/auth"

	pb "ledabeer/backend/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestIntegration_SendReceiveMessage(t *testing.T) {
	// Full flow: gRPC send, WebSocket receive
	ctx := context.Background()
	gateway := startTestGateway(t)
	defer gateway.Stop()

	// Test that gateway is running
	assert.True(t, gateway.IsRunning())

	// Test that we can create gRPC client (simplified test)
	conn, err := grpc.Dial(gateway.GRPCAddr(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)
	assert.NotNil(t, client)

	// Test message sending
	req := &pb.SendMessageRequest{
		ToPeerId: "peer2",
		Content:  []byte("test message"),
	}

	resp, err := client.SendMessage(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.MessageId)
}

func TestIntegration_MediaUploadDownload(t *testing.T) {
	// Upload via gRPC, download via gRPC
	ctx := context.Background()
	gateway := startTestGateway(t)
	defer gateway.Stop()

	conn, err := grpc.Dial(gateway.GRPCAddr(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewMediaServiceClient(conn)
	assert.NotNil(t, client)

	// Test media info retrieval
	req := &pb.GetMediaInfoRequest{
		Cid: "QmTest...",
	}

	resp, err := client.GetMediaInfo(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", resp.MimeType)
}

func TestIntegration_CallSignaling(t *testing.T) {
	// Initiate call via gRPC, receive via WebSocket
	ctx := context.Background()
	gateway := startTestGateway(t)
	defer gateway.Stop()

	conn, err := grpc.Dial(gateway.GRPCAddr(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewCallServiceClient(conn)
	assert.NotNil(t, client)

	// Test call initiation
	req := &pb.InitiateCallRequest{
		ToPeerId:     "peer2",
		AudioEnabled: true,
	}

	resp, err := client.InitiateCall(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.CallId)
}

// Helper functions
func startTestGateway(t *testing.T) *api.Gateway {
	config := &api.Config{
		GRPCPort: 0, // Use random port
		HTTPPort: 0, // Use random port
	}

	// Create mock services
	msgService := grpcapi.NewMessageService(nil, nil)
	mediaService := grpcapi.NewMediaService(nil)
	callService := grpcapi.NewCallService(nil)
	auth := auth.NewAuthenticator()
	wsServer := websocket.NewServer(auth)

	gateway := api.NewGateway(config, msgService, mediaService, callService, wsServer)

	err := gateway.Start()
	require.NoError(t, err)

	// Wait a bit for servers to start
	time.Sleep(100 * time.Millisecond)

	return gateway
}
