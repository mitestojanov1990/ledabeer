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

func TestMediaService_UploadMedia(t *testing.T) {
	// Should upload media via streaming
	service := setupTestMediaService(t)

	stream := &mockClientStream{
		chunks: [][]byte{
			make([]byte, 1024),
			make([]byte, 1024),
		},
	}

	err := service.UploadMedia(stream)
	require.NoError(t, err)
}

func TestMediaService_DownloadMedia(t *testing.T) {
	// Should stream media download
	service := setupTestMediaService(t)

	req := &pb.DownloadMediaRequest{
		Cid: "QmTest...",
	}
	stream := &mockMediaServerStream{}

	err := service.DownloadMedia(req, stream)
	require.NoError(t, err)
}

func TestMediaService_GetMediaInfo(t *testing.T) {
	// Should retrieve media metadata
	ctx := context.Background()
	service := setupTestMediaService(t)

	req := &pb.GetMediaInfoRequest{
		Cid: "QmTest...",
	}

	resp, err := service.GetMediaInfo(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", resp.MimeType)
}

// Helper functions
func setupTestMediaService(t *testing.T) *grpc.MediaService {
	// Create a mock media handler
	return grpc.NewMediaService(nil)
}

type mockClientStream struct {
	chunks [][]byte
	index  int
}

func (m *mockClientStream) Recv() (*pb.MediaChunk, error) {
	if m.index >= len(m.chunks) {
		return nil, io.EOF
	}

	chunk := &pb.MediaChunk{
		Data:        m.chunks[m.index],
		ChunkIndex:  int32(m.index),
		TotalChunks: int32(len(m.chunks)),
	}
	m.index++
	return chunk, nil
}

func (m *mockClientStream) Context() context.Context {
	return context.Background()
}

func (m *mockClientStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockClientStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockClientStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockClientStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockClientStream) SetTrailer(metadata.MD) {
}

func (m *mockClientStream) SendAndClose(resp *pb.UploadMediaResponse) error {
	return nil
}

type mockMediaServerStream struct{}

func (m *mockMediaServerStream) Send(chunk *pb.MediaChunk) error {
	return nil
}

func (m *mockMediaServerStream) Context() context.Context {
	return context.Background()
}

func (m *mockMediaServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockMediaServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockMediaServerStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockMediaServerStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockMediaServerStream) SetTrailer(metadata.MD) {
}
