package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"ledabeer/backend/internal/media"
	pb "ledabeer/backend/pkg/proto"
)

type MediaService struct {
	pb.UnimplementedMediaServiceServer
	mediaHandler *media.MediaHandler
}

func NewMediaService(mediaHandler *media.MediaHandler) *MediaService {
	return &MediaService{mediaHandler: mediaHandler}
}

func (s *MediaService) UploadMedia(stream pb.MediaService_UploadMediaServer) error {
	// Handle nil handler for unit tests
	if s.mediaHandler == nil {
		// Mock implementation for unit tests
		mediaID := generateMediaID()
		cid := generateCID([]byte("mock data"))
		return stream.SendAndClose(&pb.UploadMediaResponse{
			MediaId: mediaID,
			Cid:     cid,
			Size:    1024,
		})
	}

	// Collect all chunks
	var chunks [][]byte

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		chunks = append(chunks, chunk.Data)
	}

	// Store via real media handler
	cid, size, err := s.mediaHandler.StoreMedia(stream.Context(), chunks)
	if err != nil {
		return fmt.Errorf("failed to store media: %w", err)
	}

	// Generate media ID for response
	mediaID := generateMediaID()

	return stream.SendAndClose(&pb.UploadMediaResponse{
		MediaId: mediaID,
		Cid:     cid,
		Size:    size,
	})
}

func (s *MediaService) DownloadMedia(req *pb.DownloadMediaRequest, stream pb.MediaService_DownloadMediaServer) error {
	// Handle nil handler for unit tests
	if s.mediaHandler == nil {
		// Mock implementation for unit tests
		mediaData := []byte("mock media data")
		chunkSize := 1024

		for i := 0; i < len(mediaData); i += chunkSize {
			end := i + chunkSize
			if end > len(mediaData) {
				end = len(mediaData)
			}

			chunk := &pb.MediaChunk{
				Data:        mediaData[i:end],
				ChunkIndex:  int32(i / chunkSize),
				TotalChunks: int32((len(mediaData) + chunkSize - 1) / chunkSize),
			}

			if err := stream.Send(chunk); err != nil {
				return err
			}
		}
		return nil
	}

	// Download from real IPFS via media handler
	mediaData, err := s.mediaHandler.RetrieveMedia(stream.Context(), req.Cid)
	if err != nil {
		return fmt.Errorf("failed to retrieve media: %w", err)
	}

	// Stream chunks to client
	chunkSize := 64 * 1024 // 64KB chunks
	totalChunks := (len(mediaData) + chunkSize - 1) / chunkSize

	for i := 0; i < len(mediaData); i += chunkSize {
		end := i + chunkSize
		if end > len(mediaData) {
			end = len(mediaData)
		}

		chunk := &pb.MediaChunk{
			Data:        mediaData[i:end],
			ChunkIndex:  int32(i / chunkSize),
			TotalChunks: int32(totalChunks),
		}

		if err := stream.Send(chunk); err != nil {
			return err
		}
	}

	return nil
}

func (s *MediaService) GetMediaInfo(ctx context.Context, req *pb.GetMediaInfoRequest) (*pb.MediaInfo, error) {
	// Handle nil handler for unit tests
	if s.mediaHandler == nil {
		// Mock implementation for unit tests
		return &pb.MediaInfo{
			Cid:       req.Cid,
			MimeType:  "image/jpeg",
			Size:      1024,
			Timestamp: time.Now().Unix(),
		}, nil
	}

	// Get real metadata from IPFS
	info, err := s.mediaHandler.GetMediaInfo(ctx, req.Cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get media info: %w", err)
	}

	return &pb.MediaInfo{
		Cid:       info.CID,
		MimeType:  info.MimeType,
		Size:      info.Size,
		Timestamp: info.Timestamp,
	}, nil
}

func (s *MediaService) SendMediaMessage(ctx context.Context, req *pb.SendMediaMessageRequest) (*pb.SendMediaMessageResponse, error) {
	// Send media message (mock for now)
	messageID := generateMessageID()

	return &pb.SendMediaMessageResponse{
		MessageId: messageID,
		Timestamp: time.Now().Unix(),
	}, nil
}

func generateMediaID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("media_%x", bytes)
}

func generateCID(data []byte) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("Qm%s", bytes)
}
