package grpc_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/api/grpc"
	pb "ledabeer/backend/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestMessageService_SendMessage(t *testing.T) {
	// Should send encrypted message to peer
	ctx := context.Background()
	service := setupTestMessageService(t)

	req := &pb.SendMessageRequest{
		ToPeerId: "peer2",
		Content:  []byte("encrypted content"),
	}

	resp, err := service.SendMessage(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.MessageId)
}

func TestMessageService_ReceiveMessages(t *testing.T) {
	// Should stream incoming messages
	service := setupTestMessageService(t)

	req := &pb.ReceiveMessagesRequest{}
	stream := &mockServerStream{}

	err := service.ReceiveMessages(req, stream)
	require.NoError(t, err)
}

func TestMessageService_SendGroupMessage(t *testing.T) {
	// Should send message to group
	ctx := context.Background()
	service := setupTestMessageService(t)

	req := &pb.SendGroupMessageRequest{
		GroupId: "group1",
		Content: []byte("group message"),
	}

	resp, err := service.SendGroupMessage(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.MessageId)
}

// Helper functions
func setupTestMessageService(t *testing.T) *grpc.MessageService {
	// Create a mock message handler and group manager
	// For unit tests, we can pass nil since we're testing the service logic
	return grpc.NewMessageService(nil, nil)
}

type mockServerStream struct{}

func (m *mockServerStream) Send(msg *pb.Message) error {
	return nil
}

func (m *mockServerStream) Context() context.Context {
	return context.Background()
}

func (m *mockServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) SetTrailer(metadata.MD) {
}
