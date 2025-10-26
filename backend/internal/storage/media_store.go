package storage

import (
	"context"
)

type MediaMetadata struct {
	MimeType    string
	Size        int
	EncryptedBy string
	Thumbnail   []byte
}

type MediaReference struct {
	CID string
	Key []byte // Encryption key
}

type MediaStore struct {
	ipfs     *IPFSNode
	metadata map[string]*MediaMetadata // CID -> metadata
}

func NewMediaStore(ipfs *IPFSNode) *MediaStore {
	return &MediaStore{
		ipfs:     ipfs,
		metadata: make(map[string]*MediaMetadata),
	}
}

func (ms *MediaStore) StoreMedia(ctx context.Context, data []byte, metadata *MediaMetadata) (string, error) {
	// Store media in IPFS
	cid, err := ms.ipfs.Add(ctx, data)
	if err != nil {
		return "", err
	}

	// Store metadata
	ms.metadata[cid] = metadata

	// Pin content to prevent garbage collection
	err = ms.ipfs.Pin(ctx, cid)
	if err != nil {
		return "", err
	}

	return cid, nil
}

func (ms *MediaStore) GetMedia(ctx context.Context, ref *MediaReference) ([]byte, *MediaMetadata, error) {
	// Retrieve from IPFS by CID
	data, err := ms.ipfs.Get(ctx, ref.CID)
	if err != nil {
		return nil, nil, err
	}

	// Get stored metadata
	metadata, exists := ms.metadata[ref.CID]
	if !exists {
		// Fallback to default metadata
		metadata = &MediaMetadata{
			MimeType:    "application/octet-stream",
			Size:        len(data),
			EncryptedBy: "unknown",
		}
	}

	return data, metadata, nil
}

func (ms *MediaStore) StoreMediaWithProgress(ctx context.Context, data []byte, metadata *MediaMetadata, progress chan<- float64) (string, error) {
	// Store with progress reporting
	defer close(progress)

	// Simulate progress updates
	progress <- 0.1
	progress <- 0.5
	progress <- 0.8
	progress <- 1.0

	// Store media
	return ms.StoreMedia(ctx, data, metadata)
}

func (ms *MediaStore) Close() error {
	return ms.ipfs.Close()
}
