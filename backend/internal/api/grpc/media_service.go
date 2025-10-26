package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	pb "ledabeer/backend/pkg/proto"
)

type MediaService struct {
	pb.UnimplementedMediaServiceServer
	mediaHandler interface{} // Will be replaced with actual media handler
}

func NewMediaService(handler interface{}) *MediaService {
	return &MediaService{mediaHandler: handler}
}

func (s *MediaService) UploadMedia(stream pb.MediaService_UploadMediaServer) error {
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

	// Reassemble media
	var mediaData []byte
	for _, chunk := range chunks {
		mediaData = append(mediaData, chunk...)
	}

	// Generate media ID and CID
	mediaID := generateMediaID()
	cid := generateCID(mediaData)

	// Store in IPFS (mock for now)

	return stream.SendAndClose(&pb.UploadMediaResponse{
		MediaId: mediaID,
		Cid:     cid,
		Size:    int64(len(mediaData)),
	})
}

func (s *MediaService) DownloadMedia(req *pb.DownloadMediaRequest, stream pb.MediaService_DownloadMediaServer) error {
	// Download media from IPFS (mock for now)
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

func (s *MediaService) GetMediaInfo(ctx context.Context, req *pb.GetMediaInfoRequest) (*pb.MediaInfo, error) {
	// Get media info from IPFS (mock for now)
	return &pb.MediaInfo{
		Cid:       req.Cid,
		MimeType:  "image/jpeg",
		Size:      1024,
		Timestamp: time.Now().Unix(),
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
