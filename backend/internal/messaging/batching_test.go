package messaging

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageBatching_RealBatching_MultipleMessages(t *testing.T) {
	// Should batch multiple messages into single network call
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	host2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host2.Close()

	// Set up stream handler
	host2.SetStreamHandler("/ledabeer/messaging", func(stream libp2pnetwork.Stream) {
		defer stream.Close()
		// Echo handler for testing
		buf := make([]byte, 1024)
		stream.Read(buf)
		stream.Write(buf)
	})

	// Connect hosts
	err = network.Connect(ctx, host1, host2)
	require.NoError(t, err)

	// Create message handler with batching
	msgHandler := NewMessageHandlerWithBatching(host1, &BatchingConfig{
		BatchSize:    5,
		BatchTimeout: time.Millisecond * 100,
		MaxBatchSize: 10,
	})

	// Send multiple messages quickly
	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}
	messageIDs := make([]string, len(messages))

	for i, content := range messages {
		messageID, err := msgHandler.SendMessage(ctx, host2.ID().String(), []byte(content))
		require.NoError(t, err)
		messageIDs[i] = messageID
	}

	// Wait for batching to complete
	time.Sleep(200 * time.Millisecond)

	// Verify messages were batched
	stats := msgHandler.GetBatchingStats()
	assert.Equal(t, 1, stats.BatchesSent) // Should be 1 batch for 5 messages
	assert.Equal(t, 5, stats.MessagesBatched)
	assert.Equal(t, 0, stats.MessagesPending)
}

func TestMessageBatching_RealBatching_TimeoutFlush(t *testing.T) {
	// Should flush batch when timeout is reached
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	host2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host2.Close()

	// Set up stream handler
	host2.SetStreamHandler("/ledabeer/messaging", func(stream libp2pnetwork.Stream) {
		defer stream.Close()
		buf := make([]byte, 1024)
		stream.Read(buf)
		stream.Write(buf)
	})

	// Connect hosts
	err = network.Connect(ctx, host1, host2)
	require.NoError(t, err)

	// Create message handler with short timeout
	msgHandler := NewMessageHandlerWithBatching(host1, &BatchingConfig{
		BatchSize:    10,                    // Large batch size
		BatchTimeout: time.Millisecond * 50, // Short timeout
		MaxBatchSize: 20,
	})

	// Send one message
	_, err = msgHandler.SendMessage(ctx, host2.ID().String(), []byte("single message"))
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Verify batch was flushed due to timeout
	stats := msgHandler.GetBatchingStats()
	assert.Equal(t, 1, stats.BatchesSent)
	assert.Equal(t, 1, stats.MessagesBatched)
	assert.Equal(t, 0, stats.MessagesPending)
}

func TestMessageBatching_RealBatching_MaxBatchSize(t *testing.T) {
	// Should respect maximum batch size
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	host2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host2.Close()

	// Set up stream handler
	host2.SetStreamHandler("/ledabeer/messaging", func(stream libp2pnetwork.Stream) {
		defer stream.Close()
		buf := make([]byte, 1024)
		stream.Read(buf)
		stream.Write(buf)
	})

	// Connect hosts
	err = network.Connect(ctx, host1, host2)
	require.NoError(t, err)

	// Create message handler with small max batch size
	msgHandler := NewMessageHandlerWithBatching(host1, &BatchingConfig{
		BatchSize:    3,
		BatchTimeout: time.Millisecond * 100, // Short timeout
		MaxBatchSize: 3,                      // Small max batch size
	})

	// Send more messages than max batch size
	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}
	for _, content := range messages {
		_, err := msgHandler.SendMessage(ctx, host2.ID().String(), []byte(content))
		require.NoError(t, err)
	}

	// Wait for batching to complete (including timeout flush)
	time.Sleep(300 * time.Millisecond)

	// Verify multiple batches were created
	stats := msgHandler.GetBatchingStats()
	assert.GreaterOrEqual(t, stats.BatchesSent, 2) // Should be at least 2 batches
	assert.Equal(t, 5, stats.MessagesBatched)
	assert.Equal(t, 0, stats.MessagesPending)
}

func TestMessageBatching_RealBatching_ConcurrentBatches(t *testing.T) {
	// Should handle concurrent batches to different peers
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	// Create multiple target hosts
	targets := make([]host.Host, 3)
	for i := 0; i < 3; i++ {
		target, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		defer target.Close()

		// Set up stream handler
		target.SetStreamHandler("/ledabeer/messaging", func(stream libp2pnetwork.Stream) {
			defer stream.Close()
			buf := make([]byte, 1024)
			stream.Read(buf)
			stream.Write(buf)
		})

		// Connect to host1
		err = network.Connect(ctx, host1, target)
		require.NoError(t, err)

		targets[i] = target
	}

	// Create message handler with batching
	msgHandler := NewMessageHandlerWithBatching(host1, &BatchingConfig{
		BatchSize:    2,
		BatchTimeout: time.Millisecond * 100,
		MaxBatchSize: 5,
	})

	// Send messages to different peers concurrently
	for i, target := range targets {
		for j := 0; j < 3; j++ {
			_, err := msgHandler.SendMessage(ctx, target.ID().String(), []byte(fmt.Sprintf("msg%d-%d", i, j)))
			require.NoError(t, err)
		}
	}

	// Wait for batching to complete
	time.Sleep(200 * time.Millisecond)

	// Verify all messages were batched
	stats := msgHandler.GetBatchingStats()
	assert.GreaterOrEqual(t, stats.BatchesSent, 3) // At least 3 batches (one per peer)
	assert.Equal(t, 9, stats.MessagesBatched)      // 3 messages * 3 peers
	assert.Equal(t, 0, stats.MessagesPending)
}

func TestMessageBatching_RealBatching_ErrorHandling(t *testing.T) {
	// Should handle batching errors gracefully
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	// Create message handler with batching
	msgHandler := NewMessageHandlerWithBatching(host1, &BatchingConfig{
		BatchSize:    2,
		BatchTimeout: time.Millisecond * 100,
		MaxBatchSize: 5,
	})

	// Try to send message to non-existent peer
	_, err = msgHandler.SendMessage(ctx, "non-existent-peer", []byte("test message"))
	assert.Error(t, err)

	// Verify error handling doesn't break batching
	stats := msgHandler.GetBatchingStats()
	assert.Equal(t, 0, stats.BatchesSent)
	assert.Equal(t, 0, stats.MessagesBatched)
}
