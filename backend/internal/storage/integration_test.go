package storage_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_OfflineMessageStorage(t *testing.T) {
	ctx := context.Background()

	// Setup IPFS node
	ipfs, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer ipfs.Close()

	// Create message store
	messageStore := storage.NewMessageStore(ipfs)
	defer messageStore.Close()

	// Store a message for offline delivery
	msg := &storage.Message{
		From:             "alice",
		To:               "bob",
		EncryptedContent: []byte("Hello Bob!"),
		Timestamp:        time.Now(),
	}

	cid, err := messageStore.StoreMessage(ctx, msg)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Retrieve messages for Bob
	messages, err := messageStore.GetMessagesFor(ctx, "bob")
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Hello Bob!", string(messages[0].EncryptedContent))
}

func TestIntegration_MediaSharing(t *testing.T) {
	ctx := context.Background()

	// Setup IPFS node
	ipfs, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer ipfs.Close()

	// Create media store
	mediaStore := storage.NewMediaStore(ipfs)
	defer mediaStore.Close()

	// Store media file
	imageData := []byte("test image data")
	metadata := &storage.MediaMetadata{
		MimeType:    "image/jpeg",
		Size:        len(imageData),
		EncryptedBy: "alice",
	}

	cid, err := mediaStore.StoreMedia(ctx, imageData, metadata)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Create media reference
	ref := &storage.MediaReference{
		CID: cid,
		Key: []byte("decryption-key"),
	}

	// Retrieve media
	retrievedData, retrievedMeta, err := mediaStore.GetMedia(ctx, ref)
	require.NoError(t, err)
	assert.Equal(t, imageData, retrievedData)
	assert.Equal(t, "image/jpeg", retrievedMeta.MimeType)
}

func TestIntegration_MessageTTL(t *testing.T) {
	ctx := context.Background()

	// Setup IPFS node
	ipfs, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer ipfs.Close()

	// Create message store
	messageStore := storage.NewMessageStore(ipfs)
	defer messageStore.Close()

	// Store message with short TTL
	msg := &storage.Message{
		From:             "alice",
		To:               "bob",
		EncryptedContent: []byte("Temporary message"),
		Timestamp:        time.Now(),
		TTL:              1 * time.Second,
	}

	_, err = messageStore.StoreMessage(ctx, msg)
	require.NoError(t, err)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Message should be expired
	messages, err := messageStore.GetMessagesFor(ctx, "bob")
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestIntegration_MediaProgressTracking(t *testing.T) {
	ctx := context.Background()

	// Setup IPFS node
	ipfs, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer ipfs.Close()

	// Create media store
	mediaStore := storage.NewMediaStore(ipfs)
	defer mediaStore.Close()

	// Test progress tracking
	data := make([]byte, 1024*1024) // 1MB
	metadata := &storage.MediaMetadata{MimeType: "video/mp4"}

	progress := make(chan float64, 10)
	_, err = mediaStore.StoreMediaWithProgress(ctx, data, metadata, progress)
	require.NoError(t, err)

	// Should have received progress updates
	var progressValues []float64
	for p := range progress {
		progressValues = append(progressValues, p)
	}

	assert.Greater(t, len(progressValues), 0)
	assert.Equal(t, 1.0, progressValues[len(progressValues)-1])
}
