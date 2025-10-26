package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRUCache_RealCache_ContentAccess(t *testing.T) {
	// Should cache frequently accessed content
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create LRU cache
	cache := NewLRUCache(&LRUConfig{
		MaxSize:         5,
		TTL:             time.Minute,
		CleanupInterval: time.Second,
	})

	// Store content in IPFS
	content1 := []byte("frequently accessed content 1")
	cid1, err := ipfsNode.Add(ctx, content1)
	require.NoError(t, err)

	// Store second content for testing
	content2 := []byte("frequently accessed content 2")
	_, err = ipfsNode.Add(ctx, content2)
	require.NoError(t, err)

	// First access - cache miss, store in cache
	cached, err := cache.Get(ctx, cid1)
	require.NoError(t, err)
	if cached == nil {
		// Cache miss - fetch from IPFS and store in cache
		content, err := ipfsNode.Get(ctx, cid1)
		require.NoError(t, err)
		err = cache.Set(ctx, cid1, content)
		require.NoError(t, err)
		cached = content
	}
	assert.Equal(t, content1, cached)

	// Subsequent accesses - cache hits
	for i := 0; i < 2; i++ {
		cached, err := cache.Get(ctx, cid1)
		require.NoError(t, err)
		assert.Equal(t, content1, cached)
	}

	// Verify cache stats
	stats := cache.GetStats()
	assert.Equal(t, 2, stats.Hits)   // 2 subsequent accesses
	assert.Equal(t, 1, stats.Misses) // 1 initial miss
	assert.Equal(t, 1, stats.Size)
}

func TestLRUCache_RealCache_Eviction(t *testing.T) {
	// Should evict least recently used content when cache is full
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create small LRU cache
	cache := NewLRUCache(&LRUConfig{
		MaxSize:         2,
		TTL:             time.Minute,
		CleanupInterval: time.Second,
	})

	// Store multiple content items
	contents := make([]string, 3)
	cids := make([]string, 3)

	for i := 0; i < 3; i++ {
		content := []byte(fmt.Sprintf("content %d", i))
		cid, err := ipfsNode.Add(ctx, content)
		require.NoError(t, err)
		contents[i] = string(content)
		cids[i] = cid
	}

	// Access first two items (cache miss, then store)
	for i := 0; i < 2; i++ {
		cached, err := cache.Get(ctx, cids[i])
		require.NoError(t, err)
		if cached == nil {
			content, err := ipfsNode.Get(ctx, cids[i])
			require.NoError(t, err)
			err = cache.Set(ctx, cids[i], content)
			require.NoError(t, err)
		}
	}

	// Access third item (should evict first item)
	cached, err := cache.Get(ctx, cids[2])
	require.NoError(t, err)
	if cached == nil {
		content, err := ipfsNode.Get(ctx, cids[2])
		require.NoError(t, err)
		err = cache.Set(ctx, cids[2], content)
		require.NoError(t, err)
	}

	// First item should be evicted (cache miss)
	cached, err = cache.Get(ctx, cids[0])
	require.NoError(t, err)
	assert.Nil(t, cached) // Should be cache miss due to eviction

	// Verify cache stats
	stats := cache.GetStats()
	assert.Equal(t, 0, stats.Hits)      // No hits due to eviction
	assert.Equal(t, 4, stats.Misses)    // All accesses are misses due to eviction
	assert.Equal(t, 1, stats.Evictions) // One eviction occurred
	assert.Equal(t, 2, stats.Size)      // Cache should be at max size
}

func TestLRUCache_RealCache_TTLExpiration(t *testing.T) {
	// Should expire content after TTL
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create cache with short TTL
	cache := NewLRUCache(&LRUConfig{
		MaxSize:         10,
		TTL:             time.Millisecond * 100, // Very short TTL
		CleanupInterval: time.Millisecond * 50,  // Frequent cleanup
	})

	// Store content
	content := []byte("expiring content")
	cid, err := ipfsNode.Add(ctx, content)
	require.NoError(t, err)

	// Access content (cache miss, then store)
	cached, err := cache.Get(ctx, cid)
	require.NoError(t, err)
	if cached == nil {
		content, err := ipfsNode.Get(ctx, cid)
		require.NoError(t, err)
		err = cache.Set(ctx, cid, content)
		require.NoError(t, err)
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Access again (should be cache miss due to expiration)
	cached, err = cache.Get(ctx, cid)
	require.NoError(t, err)
	assert.Nil(t, cached) // Should be cache miss due to expiration

	// Verify cache stats
	stats := cache.GetStats()
	assert.Equal(t, 0, stats.Hits)   // No hits due to expiration
	assert.Equal(t, 2, stats.Misses) // First access miss, second access miss after expiration
}

func TestLRUCache_RealCache_ConcurrentAccess(t *testing.T) {
	// Should handle concurrent access safely
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create cache
	cache := NewLRUCache(&LRUConfig{
		MaxSize:         10,
		TTL:             time.Minute,
		CleanupInterval: time.Second,
	})

	// Store content
	content := []byte("concurrent access content")
	cid, err := ipfsNode.Add(ctx, content)
	require.NoError(t, err)

	// Concurrent access
	numGoroutines := 10
	results := make(chan []byte, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			cached, err := cache.Get(ctx, cid)
			if err != nil {
				errors <- err
				return
			}
			if cached == nil {
				// Cache miss - fetch from IPFS and store
				content, err := ipfsNode.Get(ctx, cid)
				if err != nil {
					errors <- err
					return
				}
				err = cache.Set(ctx, cid, content)
				if err != nil {
					errors <- err
					return
				}
				cached = content
			}
			results <- cached
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			assert.Equal(t, content, result)
			successCount++
		case err := <-errors:
			t.Logf("Cache access failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for cache access")
		}
	}

	// Verify all accesses succeeded
	assert.Equal(t, numGoroutines, successCount)
}

func TestLRUCache_RealCache_Performance(t *testing.T) {
	// Should improve performance for repeated access
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create cache
	cache := NewLRUCache(&LRUConfig{
		MaxSize:         10,
		TTL:             time.Minute,
		CleanupInterval: time.Second,
	})

	// Store content
	content := []byte("performance test content")
	cid, err := ipfsNode.Add(ctx, content)
	require.NoError(t, err)

	// First access (cache miss)
	start := time.Now()
	cached, err := cache.Get(ctx, cid)
	require.NoError(t, err)
	if cached == nil {
		content, err := ipfsNode.Get(ctx, cid)
		require.NoError(t, err)
		err = cache.Set(ctx, cid, content)
		require.NoError(t, err)
	}
	firstAccessTime := time.Since(start)

	// Second access (cache hit)
	start = time.Now()
	cached, err = cache.Get(ctx, cid)
	require.NoError(t, err)
	assert.NotNil(t, cached) // Should be cache hit
	secondAccessTime := time.Since(start)

	// Cache hit should be faster
	assert.Less(t, secondAccessTime, firstAccessTime)

	// Verify cache stats
	stats := cache.GetStats()
	assert.Equal(t, 1, stats.Hits)
	assert.Equal(t, 1, stats.Misses)
}

func TestLRUCache_RealCache_MemoryUsage(t *testing.T) {
	// Should respect memory limits
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := NewIPFSNode(ctx, &IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create cache with small max size
	cache := NewLRUCache(&LRUConfig{
		MaxSize:         2,
		TTL:             time.Minute,
		CleanupInterval: time.Second,
	})

	// Store more content than cache can hold
	for i := 0; i < 5; i++ {
		content := []byte(fmt.Sprintf("content %d", i))
		cid, err := ipfsNode.Add(ctx, content)
		require.NoError(t, err)

		cached, err := cache.Get(ctx, cid)
		require.NoError(t, err)
		if cached == nil {
			content, err := ipfsNode.Get(ctx, cid)
			require.NoError(t, err)
			err = cache.Set(ctx, cid, content)
			require.NoError(t, err)
		}
	}

	// Verify cache size doesn't exceed limit
	stats := cache.GetStats()
	assert.LessOrEqual(t, stats.Size, 2)
	assert.Equal(t, 2, stats.MaxSize)
}
