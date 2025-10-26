package storage_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPFS_StoreContent(t *testing.T) {
	ctx := context.Background()
	node := setupTestIPFSNode(t)
	defer node.Close()

	// Store encrypted content
	content := []byte("test encrypted message")
	cid, err := node.Add(ctx, content)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
}

func TestIPFS_RetrieveContent(t *testing.T) {
	ctx := context.Background()
	node := setupTestIPFSNode(t)
	defer node.Close()

	// Store and retrieve
	original := []byte("test content")
	cid, _ := node.Add(ctx, original)

	retrieved, err := node.Get(ctx, cid)
	require.NoError(t, err)
	assert.Equal(t, original, retrieved)
}

func TestIPFS_PinContent(t *testing.T) {
	// Should pin content to prevent garbage collection
	ctx := context.Background()
	node := setupTestIPFSNode(t)
	defer node.Close()

	content := []byte("important content")
	cid, _ := node.Add(ctx, content)

	err := node.Pin(ctx, cid)
	require.NoError(t, err)

	// Verify pinned
	assert.True(t, node.IsPinned(ctx, cid))
}

func TestIPFS_UnpinContent(t *testing.T) {
	// Should unpin to allow garbage collection
	ctx := context.Background()
	node := setupTestIPFSNode(t)
	defer node.Close()

	content := []byte("temporary content")
	cid, _ := node.Add(ctx, content)
	node.Pin(ctx, cid)

	err := node.Unpin(ctx, cid)
	require.NoError(t, err)

	assert.False(t, node.IsPinned(ctx, cid))
}

// Helper function for tests
func setupTestIPFSNode(t *testing.T) *storage.IPFSNode {
	ctx := context.Background()
	node, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	return node
}
