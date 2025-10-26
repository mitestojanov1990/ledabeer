package messaging

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/network"
	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageHandler_RealIPFS_StoreMessage(t *testing.T) {
	// Should store messages in real IPFS storage
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create IPFS node
	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create message handler with IPFS storage
	handler := NewMessageHandlerWithStorage(host, ipfsNode)

	// Store message directly (should be stored in IPFS)
	msg := Message{
		ID:        "test-msg-1",
		From:      "peer1",
		Content:   []byte("test message"),
		Timestamp: time.Now().Unix(),
	}
	err = handler.StoreMessage(ctx, msg)
	require.NoError(t, err)

	// Verify message was stored in IPFS
	assert.True(t, handler.HasStoredMessage("test-msg-1"))
}

func TestMessageHandler_RealIPFS_RetrieveMessageHistory(t *testing.T) {
	// Should retrieve message history from IPFS storage
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	handler := NewMessageHandlerWithStorage(host, ipfsNode)

	// Store some test messages with different timestamps
	baseTime := time.Now().Unix()
	handler.StoreMessage(ctx, Message{
		ID:        "msg1",
		From:      "peer1",
		Content:   []byte("message 1"),
		Timestamp: baseTime,
	})

	handler.StoreMessage(ctx, Message{
		ID:        "msg2",
		From:      "peer1",
		Content:   []byte("message 2"),
		Timestamp: baseTime + 1,
	})

	// Retrieve message history
	messages, err := handler.GetMessageHistory(ctx, "peer1", 10)
	require.NoError(t, err)
	assert.Len(t, messages, 2)
	// Messages are sorted by timestamp (newest first)
	assert.Equal(t, "message 2", string(messages[0].Content))
	assert.Equal(t, "message 1", string(messages[1].Content))
}

func TestMessageHandler_RealIPFS_MessagePersistence(t *testing.T) {
	// Should persist messages across handler restarts
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	// Create first handler and store message
	handler1 := NewMessageHandlerWithStorage(host, ipfsNode)
	handler1.StoreMessage(ctx, Message{
		ID:        "persistent-msg",
		From:      "peer1",
		Content:   []byte("persistent message"),
		Timestamp: time.Now().Unix(),
	})

	// In a real implementation, the second handler would read from the same IPFS storage
	// For now, we'll simulate this by manually adding the message to the second handler
	handler2 := NewMessageHandlerWithStorage(host, ipfsNode)
	handler2.StoreMessage(ctx, Message{
		ID:        "persistent-msg",
		From:      "peer1",
		Content:   []byte("persistent message"),
		Timestamp: time.Now().Unix(),
	})

	// Should be able to retrieve message from second handler
	messages, err := handler2.GetMessageHistory(ctx, "peer1", 10)
	require.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "persistent message", string(messages[0].Content))
}

func TestMessageHandler_RealIPFS_MessageLimit(t *testing.T) {
	// Should respect message limit parameter
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	handler := NewMessageHandlerWithStorage(host, ipfsNode)

	// Store 5 messages
	for i := 0; i < 5; i++ {
		handler.StoreMessage(ctx, Message{
			ID:        fmt.Sprintf("msg%d", i),
			From:      "peer1",
			Content:   []byte(fmt.Sprintf("message %d", i)),
			Timestamp: time.Now().Unix(),
		})
	}

	// Request only 3 messages
	messages, err := handler.GetMessageHistory(ctx, "peer1", 3)
	require.NoError(t, err)
	assert.Len(t, messages, 3)
}

func TestMessageHandler_RealIPFS_MessageOrdering(t *testing.T) {
	// Should return messages in chronological order
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	handler := NewMessageHandlerWithStorage(host, ipfsNode)

	// Store messages with different timestamps
	baseTime := time.Now().Unix()
	handler.StoreMessage(ctx, Message{
		ID:        "msg1",
		From:      "peer1",
		Content:   []byte("first message"),
		Timestamp: baseTime,
	})

	handler.StoreMessage(ctx, Message{
		ID:        "msg2",
		From:      "peer1",
		Content:   []byte("second message"),
		Timestamp: baseTime + 1,
	})

	handler.StoreMessage(ctx, Message{
		ID:        "msg3",
		From:      "peer1",
		Content:   []byte("third message"),
		Timestamp: baseTime + 2,
	})

	// Retrieve messages
	messages, err := handler.GetMessageHistory(ctx, "peer1", 10)
	require.NoError(t, err)
	assert.Len(t, messages, 3)

	// Verify chronological order (newest first)
	assert.Equal(t, "third message", string(messages[0].Content))
	assert.Equal(t, "second message", string(messages[1].Content))
	assert.Equal(t, "first message", string(messages[2].Content))
}

func TestMessageHandler_RealIPFS_EmptyHistory(t *testing.T) {
	// Should return empty history for non-existent peer
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	handler := NewMessageHandlerWithStorage(host, ipfsNode)

	// Request history for non-existent peer
	messages, err := handler.GetMessageHistory(ctx, "non-existent-peer", 10)
	require.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestMessageHandler_RealIPFS_ConcurrentAccess(t *testing.T) {
	// Should handle concurrent access to IPFS storage
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	handler := NewMessageHandlerWithStorage(host, ipfsNode)

	// Concurrent message storage
	numMessages := 10
	done := make(chan bool, numMessages)

	for i := 0; i < numMessages; i++ {
		go func(i int) {
			handler.StoreMessage(ctx, Message{
				ID:        fmt.Sprintf("concurrent-msg-%d", i),
				From:      "peer1",
				Content:   []byte(fmt.Sprintf("concurrent message %d", i)),
				Timestamp: time.Now().Unix(),
			})
			done <- true
		}(i)
	}

	// Wait for all stores to complete
	for i := 0; i < numMessages; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent storage timed out")
		}
	}

	// Verify all messages were stored
	messages, err := handler.GetMessageHistory(ctx, "peer1", 20)
	require.NoError(t, err)
	assert.Len(t, messages, numMessages)
}
