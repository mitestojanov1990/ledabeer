package storage_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageStore_StoreMessage(t *testing.T) {
	ctx := context.Background()
	store := setupTestMessageStore(t)
	defer store.Close()

	// Store encrypted message
	msg := &storage.Message{
		From:             "peer1",
		To:               "peer2",
		EncryptedContent: []byte("encrypted payload"),
		Timestamp:        time.Now(),
	}

	cid, err := store.StoreMessage(ctx, msg)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
}

func TestMessageStore_RetrieveMessages(t *testing.T) {
	ctx := context.Background()
	store := setupTestMessageStore(t)
	defer store.Close()

	// Store multiple messages for peer
	peerID := "peer2"
	msg1 := &storage.Message{From: "peer1", To: peerID, EncryptedContent: []byte("msg1")}
	msg2 := &storage.Message{From: "peer3", To: peerID, EncryptedContent: []byte("msg2")}

	store.StoreMessage(ctx, msg1)
	store.StoreMessage(ctx, msg2)

	// Retrieve all messages for peer
	messages, err := store.GetMessagesFor(ctx, peerID)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
}

func TestMessageStore_DeleteAfterRetrieval(t *testing.T) {
	// Messages should be deleted after successful retrieval
	ctx := context.Background()
	store := setupTestMessageStore(t)
	defer store.Close()

	peerID := "peer1"
	msg := &storage.Message{From: "peer2", To: peerID, EncryptedContent: []byte("test")}
	store.StoreMessage(ctx, msg)

	// First retrieval succeeds
	messages, _ := store.GetMessagesFor(ctx, peerID)
	assert.Len(t, messages, 1)

	// Second retrieval returns empty
	messages, _ = store.GetMessagesFor(ctx, peerID)
	assert.Empty(t, messages)
}

func TestMessageStore_TTL(t *testing.T) {
	// Messages should expire after TTL
	ctx := context.Background()
	store := setupTestMessageStore(t)
	defer store.Close()

	msg := &storage.Message{
		From:             "peer1",
		To:               "peer2",
		EncryptedContent: []byte("test"),
		TTL:              1 * time.Second,
	}

	store.StoreMessage(ctx, msg)

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	messages, _ := store.GetMessagesFor(ctx, "peer2")
	assert.Empty(t, messages)
}

// Helper function for tests
func setupTestMessageStore(t *testing.T) *storage.MessageStore {
	ctx := context.Background()
	node, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{
		RepoPath: t.TempDir(),
	})
	require.NoError(t, err)
	return storage.NewMessageStore(node)
}
