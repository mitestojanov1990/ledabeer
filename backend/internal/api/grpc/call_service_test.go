package grpc_test

import (
	"context"
	"io"
	"testing"

	"ledabeer/backend/internal/api/grpc"
	pb "ledabeer/backend/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestCallService_InitiateCall(t *testing.T) {
	// Should initiate 1:1 call
	ctx := context.Background()
	service := setupTestCallService(t)

	req := &pb.InitiateCallRequest{
		ToPeerId:     "peer2",
		AudioEnabled: true,
		VideoEnabled: true,
	}

	resp, err := service.InitiateCall(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.CallId)
}

func TestCallService_AnswerCall(t *testing.T) {
	// Should answer incoming call
	ctx := context.Background()
	service := setupTestCallService(t)

	req := &pb.AnswerCallRequest{
		CallId: "call123",
		Accept: true,
	}

	resp, err := service.AnswerCall(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, pb.CallStateEnum_CONNECTED, resp.State.State)
}

func TestCallService_StreamSignaling(t *testing.T) {
	// Should stream signaling messages
	service := setupTestCallService(t)

	stream := &mockBidirectionalStream{}

	err := service.StreamSignaling(stream)
	require.NoError(t, err)
}

// Helper functions
func setupTestCallService(t *testing.T) *grpc.CallService {
	// Create a mock call handler
	return grpc.NewCallService(nil)
}

type mockBidirectionalStream struct{}

func (m *mockBidirectionalStream) Send(msg *pb.SignalingMessage) error {
	return nil
}

func (m *mockBidirectionalStream) Recv() (*pb.SignalingMessage, error) {
	return nil, io.EOF
}

func (m *mockBidirectionalStream) Context() context.Context {
	return context.Background()
}

func (m *mockBidirectionalStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockBidirectionalStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockBidirectionalStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockBidirectionalStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockBidirectionalStream) SetTrailer(metadata.MD) {
}
