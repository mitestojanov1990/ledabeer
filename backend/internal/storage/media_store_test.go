package storage_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaStore_StoreImage(t *testing.T) {
	ctx := context.Background()
	store := setupTestMediaStore(t)
	defer store.Close()

	// Store encrypted image
	imageData := []byte("encrypted image data")
	metadata := &storage.MediaMetadata{
		MimeType:    "image/jpeg",
		Size:        len(imageData),
		EncryptedBy: "peer1",
	}

	cid, err := store.StoreMedia(ctx, imageData, metadata)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
}

func TestMediaStore_RetrieveByReference(t *testing.T) {
	ctx := context.Background()
	store := setupTestMediaStore(t)
	defer store.Close()

	original := []byte("test media")
	metadata := &storage.MediaMetadata{MimeType: "video/mp4"}
	cid, _ := store.StoreMedia(ctx, original, metadata)

	// Create reference
	ref := &storage.MediaReference{
		CID: cid,
		Key: []byte("decryption key"),
	}

	// Retrieve using reference
	data, meta, err := store.GetMedia(ctx, ref)
	require.NoError(t, err)
	assert.Equal(t, original, data)
	assert.Equal(t, "video/mp4", meta.MimeType)
}

func TestMediaStore_ChunkedUpload(t *testing.T) {
	// Large media should be uploaded in chunks
	ctx := context.Background()
	store := setupTestMediaStore(t)
	defer store.Close()

	// 10MB file
	largeData := make([]byte, 10*1024*1024)
	metadata := &storage.MediaMetadata{MimeType: "video/mp4", Size: len(largeData)}

	cid, err := store.StoreMedia(ctx, largeData, metadata)
	require.NoError(t, err)

	// Should retrieve correctly
	retrieved, _, err := store.GetMedia(ctx, &storage.MediaReference{CID: cid})
	require.NoError(t, err)
	assert.Equal(t, len(largeData), len(retrieved))
}

func TestMediaStore_ProgressTracking(t *testing.T) {
	// Should report upload progress
	ctx := context.Background()
	store := setupTestMediaStore(t)
	defer store.Close()

	data := make([]byte, 5*1024*1024)
	metadata := &storage.MediaMetadata{MimeType: "image/png"}

	progress := make(chan float64, 10)
	_, err := store.StoreMediaWithProgress(ctx, data, metadata, progress)
	require.NoError(t, err)

	// Should have received progress updates
	var progressValues []float64
	for p := range progress {
		progressValues = append(progressValues, p)
	}

	assert.Greater(t, len(progressValues), 0)
	assert.Equal(t, 1.0, progressValues[len(progressValues)-1])
}

// Helper function for tests
func setupTestMediaStore(t *testing.T) *storage.MediaStore {
	ctx := context.Background()
	node, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	return storage.NewMediaStore(node)
}
