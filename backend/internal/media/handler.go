package media

import (
	"context"
	"fmt"
	"time"

	"ledabeer/backend/internal/storage"
)

type MediaHandler struct {
	ipfs   *storage.IPFSNode
	crypto CryptoHandler // For encryption
}

type MediaInfo struct {
	CID       string
	MimeType  string
	Size      int64
	Filename  string
	Timestamp int64
}

type CryptoHandler interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

func NewMediaHandler(ipfs *storage.IPFSNode) *MediaHandler {
	return &MediaHandler{ipfs: ipfs}
}

func (m *MediaHandler) StoreMedia(ctx context.Context, chunks [][]byte) (string, int64, error) {
	// Reassemble chunks
	var data []byte
	for _, chunk := range chunks {
		data = append(data, chunk...)
	}

	// Store in IPFS
	cid, err := m.ipfs.Add(ctx, data)
	if err != nil {
		return "", 0, fmt.Errorf("failed to store in IPFS: %w", err)
	}

	return cid, int64(len(data)), nil
}

func (m *MediaHandler) RetrieveMedia(ctx context.Context, cid string) ([]byte, error) {
	return m.ipfs.Get(ctx, cid)
}

func (m *MediaHandler) GetMediaInfo(ctx context.Context, cid string) (*MediaInfo, error) {
	// Get data to determine size and type
	data, err := m.ipfs.Get(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve media: %w", err)
	}

	// Determine MIME type based on content
	mimeType := detectMimeType(data)

	return &MediaInfo{
		CID:       cid,
		MimeType:  mimeType,
		Size:      int64(len(data)),
		Filename:  fmt.Sprintf("media_%s", cid[:8]),
		Timestamp: time.Now().Unix(),
	}, nil
}

func (m *MediaHandler) StoreEncryptedMedia(ctx context.Context, chunks [][]byte) (string, int64, error) {
	// Reassemble chunks
	var data []byte
	for _, chunk := range chunks {
		data = append(data, chunk...)
	}

	// Encrypt if crypto handler is available
	if m.crypto != nil {
		encrypted, err := m.crypto.Encrypt(data)
		if err != nil {
			return "", 0, fmt.Errorf("failed to encrypt media: %w", err)
		}
		data = encrypted
	}

	// Store in IPFS
	cid, err := m.ipfs.Add(ctx, data)
	if err != nil {
		return "", 0, fmt.Errorf("failed to store in IPFS: %w", err)
	}

	return cid, int64(len(data)), nil
}

func (m *MediaHandler) RetrieveEncryptedMedia(ctx context.Context, cid string) ([]byte, error) {
	data, err := m.ipfs.Get(ctx, cid)
	if err != nil {
		return nil, err
	}

	// Decrypt if crypto handler is available
	if m.crypto != nil {
		decrypted, err := m.crypto.Decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt media: %w", err)
		}
		data = decrypted
	}

	return data, nil
}

func detectMimeType(data []byte) string {
	// Simple MIME type detection based on magic bytes
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// Check for common image formats
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}

	// Check for video formats
	if len(data) >= 12 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x18 {
		return "video/mp4"
	}

	// Default to binary
	return "application/octet-stream"
}
