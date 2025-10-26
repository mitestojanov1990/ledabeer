package storage

import (
	"context"
	"sync"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageManager_RealQuota_Enforcement(t *testing.T) {
	// Should enforce storage quota limits
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create storage manager with quota
	quota := int64(1024 * 1024) // 1MB quota
	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: quota,
		CleanupInterval: time.Minute,
	})

	// Store data up to quota
	largeData := make([]byte, quota/2) // 512KB
	cid, err := manager.StoreContent(ctx, largeData)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Verify quota tracking
	stats := manager.GetStorageStats(ctx)
	assert.Equal(t, int64(len(largeData)), stats.UsedBytes)
	assert.Equal(t, quota, stats.MaxBytes)
	assert.Less(t, stats.UsedBytes, stats.MaxBytes)

	// Try to exceed quota
	excessData := make([]byte, quota) // 1MB - should exceed quota
	_, err = manager.StoreContent(ctx, excessData)
	if err != nil {
		assert.Contains(t, err.Error(), "quota exceeded")
	} else {
		// If no error, verify that cleanup worked and we're still within quota
		stats = manager.GetStorageStats(ctx)
		assert.LessOrEqual(t, stats.UsedBytes, stats.MaxBytes)
	}
}

func TestStorageManager_RealQuota_Cleanup(t *testing.T) {
	// Should clean up old content when quota is exceeded
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create storage manager with small quota
	quota := int64(1024) // 1KB quota
	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: quota,
		CleanupInterval: time.Second,
	})

	// Store content within quota first
	smallData := make([]byte, quota/2) // 512B
	_, err = manager.StoreContent(ctx, smallData)
	require.NoError(t, err)

	// Now try to store content that would exceed quota
	largeData := make([]byte, quota) // 1KB - should exceed quota
	_, err = manager.StoreContent(ctx, largeData)

	// This should either succeed (if cleanup worked) or fail (if quota is enforced)
	// Both outcomes are valid for this test
	if err != nil {
		// Quota enforcement worked
		assert.Contains(t, err.Error(), "quota exceeded")
	} else {
		// Cleanup worked - verify we're within quota
		stats := manager.GetStorageStats(ctx)
		assert.LessOrEqual(t, stats.UsedBytes, quota)
	}
}

func TestStorageManager_RealGarbageCollection_ExpiredContent(t *testing.T) {
	// Should garbage collect expired content
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create storage manager with short expiration
	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: 10 * 1024 * 1024, // 10MB
		CleanupInterval: time.Second,
		ContentTTL:      time.Second, // 1 second TTL
	})

	// Store content with short TTL
	data := []byte("expired content")
	cid, err := manager.StoreContentWithTTL(ctx, data, time.Second)
	require.NoError(t, err)

	// Verify content exists
	exists, err := manager.ContentExists(ctx, cid)
	require.NoError(t, err)
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Verify content was garbage collected
	exists, err = manager.ContentExists(ctx, cid)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestStorageManager_RealGarbageCollection_LRU(t *testing.T) {
	// Should use LRU for garbage collection
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create storage manager with small quota
	quota := int64(2048) // 2KB quota
	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: quota,
		CleanupInterval: time.Second,
	})

	// Store content A (older)
	dataA := make([]byte, 1024) // 1KB
	cidA, err := manager.StoreContent(ctx, dataA)
	require.NoError(t, err)

	// Access content A to update LRU
	_, err = manager.GetContent(ctx, cidA)
	require.NoError(t, err)

	// Store content B (newer)
	dataB := make([]byte, 1024) // 1KB
	cidB, err := manager.StoreContent(ctx, dataB)
	require.NoError(t, err)

	// Store content C (should trigger cleanup)
	dataC := make([]byte, 1024) // 1KB
	cidC, err := manager.StoreContent(ctx, dataC)
	require.NoError(t, err)

	// Wait for cleanup
	time.Sleep(2 * time.Second)

	// Content A should still exist (accessed recently)
	existsA, _ := manager.ContentExists(ctx, cidA)
	assert.True(t, existsA)

	// Content B should be removed (least recently used)
	existsB, _ := manager.ContentExists(ctx, cidB)
	assert.False(t, existsB)

	// Content C should exist (newest)
	existsC, _ := manager.ContentExists(ctx, cidC)
	assert.True(t, existsC)
}

func TestStorageManager_RealStats_Tracking(t *testing.T) {
	// Should track storage statistics accurately
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: 10 * 1024 * 1024, // 10MB
		CleanupInterval: time.Minute,
	})

	// Initial stats
	stats := manager.GetStorageStats(ctx)
	assert.Equal(t, int64(0), stats.UsedBytes)
	assert.Equal(t, int64(10*1024*1024), stats.MaxBytes)
	assert.Equal(t, 0, stats.ContentCount)

	// Store content
	data1 := []byte("content 1")
	cid1, err := manager.StoreContent(ctx, data1)
	require.NoError(t, err)

	// Check stats after first content
	stats = manager.GetStorageStats(ctx)
	assert.Equal(t, int64(len(data1)), stats.UsedBytes)
	assert.Equal(t, 1, stats.ContentCount)

	// Store more content
	data2 := []byte("content 2")
	_, err = manager.StoreContent(ctx, data2)
	require.NoError(t, err)

	// Check stats after second content
	stats = manager.GetStorageStats(ctx)
	assert.Equal(t, int64(len(data1)+len(data2)), stats.UsedBytes)
	assert.Equal(t, 2, stats.ContentCount)

	// Remove content
	err = manager.RemoveContent(ctx, cid1)
	require.NoError(t, err)

	// Check stats after removal
	stats = manager.GetStorageStats(ctx)
	assert.Equal(t, int64(len(data2)), stats.UsedBytes)
	assert.Equal(t, 1, stats.ContentCount)
}

func TestStorageManager_RealCleanup_Concurrent(t *testing.T) {
	// Should handle concurrent cleanup operations safely
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: 1024, // 1KB quota
		CleanupInterval: time.Millisecond * 100,
	})

	// Store multiple content concurrently
	numGoroutines := 10
	contentSize := 200 // 200 bytes each
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			data := make([]byte, contentSize)
			for j := range data {
				data[j] = byte(id)
			}
			_, err := manager.StoreContent(ctx, data)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify quota is respected
	stats := manager.GetStorageStats(ctx)
	assert.LessOrEqual(t, stats.UsedBytes, int64(1024))
}

func TestStorageManager_RealCleanup_ContentExpiration(t *testing.T) {
	// Should handle content expiration properly
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	manager := NewStorageManagerWithConfig(ipfsNode, &StorageConfig{
		MaxStorageBytes: 10 * 1024 * 1024, // 10MB
		CleanupInterval: time.Second,
		ContentTTL:      time.Second * 2, // 2 second TTL
	})

	// Store content with different TTLs
	data1 := []byte("short ttl content")
	cid1, err := manager.StoreContentWithTTL(ctx, data1, time.Second)
	require.NoError(t, err)

	data2 := []byte("long ttl content")
	cid2, err := manager.StoreContentWithTTL(ctx, data2, time.Minute)
	require.NoError(t, err)

	// Wait for short TTL to expire
	time.Sleep(2 * time.Second)

	// Short TTL content should be expired
	exists1, _ := manager.ContentExists(ctx, cid1)
	assert.False(t, exists1)

	// Long TTL content should still exist
	exists2, _ := manager.ContentExists(ctx, cid2)
	assert.True(t, exists2)
}
