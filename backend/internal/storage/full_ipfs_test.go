package storage_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullIPFS_NodeInitialization(t *testing.T) {
	// Skip if IPFS is not installed
	if !storage.IsIPFSInstalled() {
		t.Skip("IPFS not installed, skipping full IPFS tests")
	}

	ctx := context.Background()
	node, err := storage.NewFullIPFSNode(ctx, &storage.FullIPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	defer node.Close()

	// Should have a valid node ID
	nodeID := node.ID()
	assert.NotEmpty(t, nodeID)
	assert.NotEqual(t, "unknown-ipfs-node-id", nodeID)
}

func TestFullIPFS_ContentStorage(t *testing.T) {
	// Skip if IPFS is not installed
	if !storage.IsIPFSInstalled() {
		t.Skip("IPFS not installed, skipping full IPFS tests")
	}

	ctx := context.Background()
	node := setupTestFullIPFSNode(t)
	defer node.Close()

	// Store content
	content := []byte("test content for full IPFS")
	cid, err := node.Add(ctx, content)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Retrieve content
	retrieved, err := node.Get(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, content, retrieved)
}

func TestFullIPFS_PinOperations(t *testing.T) {
	// Skip if IPFS is not installed
	if !storage.IsIPFSInstalled() {
		t.Skip("IPFS not installed, skipping full IPFS tests")
	}

	ctx := context.Background()
	node := setupTestFullIPFSNode(t)
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

func TestFullIPFS_IntegrationWithStorage(t *testing.T) {
	// Skip if IPFS is not installed
	if !storage.IsIPFSInstalled() {
		t.Skip("IPFS not installed, skipping full IPFS tests")
	}

	ctx := context.Background()
	node := setupTestFullIPFSNode(t)
	defer node.Close()

	// Test direct storage and retrieval
	content := []byte("Hello from full IPFS!")
	cid, err := node.Add(ctx, content)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)

	// Retrieve content
	retrieved, err := node.Get(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, content, retrieved)
}

// Helper function for tests
func setupTestFullIPFSNode(t *testing.T) *storage.FullIPFSNode {
	ctx := context.Background()
	node, err := storage.NewFullIPFSNode(ctx, &storage.FullIPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	return node
}
