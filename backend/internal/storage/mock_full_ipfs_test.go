package storage_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockFullIPFS_NodeInitialization(t *testing.T) {
	ctx := context.Background()
	node, err := storage.NewMockFullIPFSNode(ctx, &storage.MockFullIPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer node.Close()

	// Should have a valid node ID
	nodeID := node.ID()
	assert.NotEmpty(t, nodeID)
	assert.Contains(t, nodeID, "12D3KooW") // Should be a valid libp2p peer ID format
}

func TestMockFullIPFS_ContentStorage(t *testing.T) {
	ctx := context.Background()
	node := setupTestMockFullIPFSNode(t)
	defer node.Close()

	// Store content
	content := []byte("test content for mock full IPFS")
	cid, err := node.Add(ctx, content)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.Contains(t, cid, "Qm") // Should be a valid IPFS CID

	// Retrieve content
	retrieved, err := node.Get(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, content, retrieved)
}

func TestMockFullIPFS_PinOperations(t *testing.T) {
	ctx := context.Background()
	node := setupTestMockFullIPFSNode(t)
	defer node.Close()

	// Store content
	content := []byte("pinned content")
	cid, _ := node.Add(ctx, content)

	// Pin content
	err := node.Pin(ctx, cid)
	require.NoError(t, err)

	// Check if pinned
	assert.True(t, node.IsPinned(ctx, cid))

	// Unpin content
	err = node.Unpin(ctx, cid)
	require.NoError(t, err)

	// Check if unpinned
	assert.False(t, node.IsPinned(ctx, cid))
}

func TestMockFullIPFS_EnhancedFeatures(t *testing.T) {
	ctx := context.Background()
	node := setupTestMockFullIPFSNode(t)
	defer node.Close()

	// Test stats
	stats, err := node.GetStats(ctx)
	require.NoError(t, err)
	assert.Contains(t, stats, "repo_size")
	assert.Contains(t, stats, "pinned_count")
	assert.Contains(t, stats, "peer_count")
	assert.Equal(t, "full-ipfs-mock", stats["mode"])

	// Test peer connection
	err = node.ConnectToPeer(ctx, "12D3KooWTestPeer")
	require.NoError(t, err)

	// Test pubsub publishing
	err = node.PublishToPubSub(ctx, "test-topic", []byte("test message"))
	require.NoError(t, err)

	// Test pubsub subscription
	ch, err := node.SubscribeToPubSub(ctx, "test-topic")
	require.NoError(t, err)
	assert.NotNil(t, ch)
}

func TestMockFullIPFS_IntegrationWithStorage(t *testing.T) {
	ctx := context.Background()
	node := setupTestMockFullIPFSNode(t)
	defer node.Close()

	// Test direct storage and retrieval
	content := []byte("Hello from mock full IPFS!")
	cid, err := node.Add(ctx, content)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Retrieve content
	retrieved, err := node.Get(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, content, retrieved)

	// Test pinning
	err = node.Pin(ctx, cid)
	require.NoError(t, err)
	assert.True(t, node.IsPinned(ctx, cid))
}

// Helper function for tests
func setupTestMockFullIPFSNode(t *testing.T) *storage.MockFullIPFSNode {
	ctx := context.Background()
	node, err := storage.NewMockFullIPFSNode(ctx, &storage.MockFullIPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	return node
}
