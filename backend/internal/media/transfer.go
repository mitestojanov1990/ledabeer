package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"ledabeer/backend/internal/crypto"
	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// MediaTransfer handles encrypted media file transfers
type MediaTransfer struct {
	host             host.Host
	ctx              context.Context
	config           TransferConfig
	mu               sync.RWMutex
	progressCallback func(TransferProgress)
}

// TransferConfig holds configuration for media transfers
type TransferConfig struct {
	MaxFileSize   int64         `json:"max_file_size"`
	ChunkSize     int           `json:"chunk_size"`
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
}

// TransferProgress represents the progress of a media transfer
type TransferProgress struct {
	Status     TransferStatus `json:"status"`
	Percentage float64        `json:"percentage"`
	BytesSent  int64          `json:"bytes_sent"`
	TotalBytes int64          `json:"total_bytes"`
	Data       []byte         `json:"data,omitempty"` // Final data when completed
}

// TransferStatus represents the status of a transfer
type TransferStatus string

const (
	TransferStatusStarted   TransferStatus = "started"
	TransferStatusProgress  TransferStatus = "progress"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
)

// MediaMetadata holds metadata about a media file
type MediaMetadata struct {
	Type string `json:"type"`
	Size int64  `json:"size"`
	Name string `json:"name"`
	Hash string `json:"hash,omitempty"`
}

// NewMediaTransfer creates a new media transfer manager
func NewMediaTransfer(ctx context.Context, h host.Host) (*MediaTransfer, error) {
	config := TransferConfig{
		MaxFileSize:   100 * 1024 * 1024, // 100MB default
		ChunkSize:     64 * 1024,         // 64KB chunks
		Timeout:       5 * time.Minute,   // 5 minute timeout
		RetryAttempts: 3,
	}
	return NewMediaTransferWithConfig(ctx, h, config)
}

// NewMediaTransferWithConfig creates a new media transfer manager with custom config
func NewMediaTransferWithConfig(ctx context.Context, h host.Host, config TransferConfig) (*MediaTransfer, error) {
	return &MediaTransfer{
		host:   h,
		ctx:    ctx,
		config: config,
	}, nil
}

// SendMedia sends a media file to a peer
func (mt *MediaTransfer) SendMedia(ctx context.Context, peerID peer.ID, reader io.Reader, metadata MediaMetadata, sharedSecret []byte) error {
	// Check file size limit
	if metadata.Size > mt.config.MaxFileSize {
		return fmt.Errorf("file size %d exceeds limit %d", metadata.Size, mt.config.MaxFileSize)
	}

	// Create stream to peer
	stream, err := mt.host.NewStream(ctx, peerID, "/ledabeer/media/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Create Double Ratchet for encryption
	ratchet := crypto.NewDoubleRatchet(sharedSecret, true) // sender

	// Send metadata first
	metadataBytes, err := marshalMetadata(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	encryptedMetadata, err := ratchet.Encrypt(metadataBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt metadata: %w", err)
	}

	err = network.WriteMessage(stream, encryptedMetadata)
	if err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	// Chunk and send file data
	chunks, err := ChunkFile(reader, mt.config.ChunkSize)
	if err != nil {
		return fmt.Errorf("failed to chunk file: %w", err)
	}

	// Send chunks
	for i, chunk := range chunks {
		chunkBytes, err := marshalChunk(chunk)
		if err != nil {
			return fmt.Errorf("failed to marshal chunk %d: %w", i, err)
		}

		encryptedChunk, err := ratchet.Encrypt(chunkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt chunk %d: %w", i, err)
		}

		err = network.WriteMessage(stream, encryptedChunk)
		if err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", i, err)
		}

		// Update progress
		mt.updateProgress(TransferProgress{
			Status:     TransferStatusProgress,
			Percentage: float64(i+1) / float64(len(chunks)) * 100,
			BytesSent:  int64(i+1) * int64(mt.config.ChunkSize),
			TotalBytes: metadata.Size,
		})
	}

	// Mark as completed
	mt.updateProgress(TransferProgress{
		Status:     TransferStatusCompleted,
		Percentage: 100.0,
		BytesSent:  metadata.Size,
		TotalBytes: metadata.Size,
	})

	return nil
}

// ReceiveMedia receives a media file from a stream
func (mt *MediaTransfer) ReceiveMedia(stream libp2pnetwork.Stream, sharedSecret []byte) ([]byte, MediaMetadata, error) {
	// Create Double Ratchet for decryption
	ratchet := crypto.NewDoubleRatchet(sharedSecret, false) // receiver

	// Receive metadata
	encryptedMetadata, err := network.ReadMessage(stream)
	if err != nil {
		return nil, MediaMetadata{}, fmt.Errorf("failed to read metadata: %w", err)
	}

	metadataBytes, err := ratchet.Decrypt(encryptedMetadata)
	if err != nil {
		return nil, MediaMetadata{}, fmt.Errorf("failed to decrypt metadata: %w", err)
	}

	metadata, err := unmarshalMetadata(metadataBytes)
	if err != nil {
		return nil, MediaMetadata{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Receive chunks
	var chunks []Chunk
	for {
		encryptedChunk, err := network.ReadMessage(stream)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, MediaMetadata{}, fmt.Errorf("failed to read chunk: %w", err)
		}

		chunkBytes, err := ratchet.Decrypt(encryptedChunk)
		if err != nil {
			return nil, MediaMetadata{}, fmt.Errorf("failed to decrypt chunk: %w", err)
		}

		chunk, err := unmarshalChunk(chunkBytes)
		if err != nil {
			return nil, MediaMetadata{}, fmt.Errorf("failed to unmarshal chunk: %w", err)
		}

		chunks = append(chunks, chunk)
	}

	// Reassemble file
	fileData, err := ReassembleChunks(chunks)
	if err != nil {
		return nil, MediaMetadata{}, fmt.Errorf("failed to reassemble chunks: %w", err)
	}

	return fileData, metadata, nil
}

// SetProgressCallback sets the progress callback function
func (mt *MediaTransfer) SetProgressCallback(callback func(TransferProgress)) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.progressCallback = callback
}

// updateProgress calls the progress callback if set
func (mt *MediaTransfer) updateProgress(progress TransferProgress) {
	mt.mu.RLock()
	callback := mt.progressCallback
	mt.mu.RUnlock()

	if callback != nil {
		callback(progress)
	}
}

// marshalMetadata serializes metadata to bytes
func marshalMetadata(metadata MediaMetadata) ([]byte, error) {
	// Simple JSON marshaling for now
	// In a real implementation, use proper JSON or protobuf
	return []byte(fmt.Sprintf(`{"type":"%s","size":%d,"name":"%s"}`,
		metadata.Type, metadata.Size, metadata.Name)), nil
}

// unmarshalMetadata deserializes metadata from bytes
func unmarshalMetadata(data []byte) (MediaMetadata, error) {
	// Simple parsing for now
	// In a real implementation, use proper JSON or protobuf
	return MediaMetadata{
		Type: "application/octet-stream", // Default type
		Size: int64(len(data)),           // Approximate size
		Name: "received-file",            // Default name
	}, nil
}

// marshalChunk serializes a chunk to bytes
func marshalChunk(chunk Chunk) ([]byte, error) {
	// Simple JSON serialization for now
	// In a real implementation, use proper protobuf
	return json.Marshal(chunk)
}

// unmarshalChunk deserializes a chunk from bytes
func unmarshalChunk(data []byte) (Chunk, error) {
	// Simple JSON deserialization for now
	var chunk Chunk
	err := json.Unmarshal(data, &chunk)
	if err != nil {
		return Chunk{}, fmt.Errorf("failed to unmarshal chunk: %w", err)
	}
	return chunk, nil
}
