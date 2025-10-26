package storage_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageManager_QuotaEnforcement(t *testing.T) {
	ctx := context.Background()
	manager := setupTestStorageManager(t)
	defer manager.Close()

	// Set quota to 1MB
	manager.SetQuota(1024 * 1024)

	// Store content within quota
	smallData := make([]byte, 512*1024) // 512KB
	cid, err := manager.StoreContent(ctx, smallData)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Try to store content that exceeds quota
	largeData := make([]byte, 2*1024*1024) // 2MB
	_, err = manager.StoreContent(ctx, largeData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota exceeded")
}

func TestStorageManager_GarbageCollection(t *testing.T) {
	ctx := context.Background()
	manager := setupTestStorageManager(t)
	defer manager.Close()

	// Store content
	data := []byte("test content")
	cid, err := manager.StoreContent(ctx, data)
	require.NoError(t, err)

	// Verify content exists
	retrieved, err := manager.GetContent(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Run garbage collection
	removed, err := manager.RunGarbageCollection(ctx)
	require.NoError(t, err)

	// Should not remove pinned content
	assert.Equal(t, 0, removed)

	// Unpin content
	err = manager.UnpinContent(ctx, cid)
	require.NoError(t, err)

	// Run garbage collection again
	removed, err = manager.RunGarbageCollection(ctx)
	require.NoError(t, err)

	// Should remove unpinned content
	assert.Greater(t, removed, 0)
}

func TestStorageManager_StorageStats(t *testing.T) {
	ctx := context.Background()
	manager := setupTestStorageManager(t)
	defer manager.Close()

	// Store some content
	data1 := []byte("content 1")
	data2 := []byte("content 2")

	manager.StoreContent(ctx, data1)
	manager.StoreContent(ctx, data2)

	// Get storage statistics
	stats := manager.GetStorageStats(ctx)

	assert.Greater(t, stats.UsedBytes, int64(0))
	assert.Greater(t, stats.ContentCount, 0)
}

func TestStorageManager_ContentExpiration(t *testing.T) {
	ctx := context.Background()
	manager := setupTestStorageManager(t)
	defer manager.Close()

	// Store content with short TTL
	data := []byte("temporary content")
	cid, err := manager.StoreContentWithTTL(ctx, data, 1*time.Second)
	require.NoError(t, err)

	// Content should exist initially
	_, err = manager.GetContent(ctx, cid)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Content should be expired
	_, err = manager.GetContent(ctx, cid)
	assert.Error(t, err)
}

// Helper function for tests
func setupTestStorageManager(t *testing.T) *storage.StorageManager {
	ctx := context.Background()
	node, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	return storage.NewStorageManager(node)
}
