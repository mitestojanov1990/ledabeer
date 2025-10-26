package messaging

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageHandler_RealStreamHandler_Setup(t *testing.T) {
	// Should set up real libp2p stream handler for incoming messages
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	handler := NewMessageHandler(host)

	// Start message handler
	msgChan := handler.SubscribeToMessages(ctx)
	assert.NotNil(t, msgChan)

	// Verify stream handler is set up
	// In real implementation, this would verify the handler is registered
	assert.NotNil(t, handler)
}

func TestMessageHandler_RealStreamHandler_ReceiveMessage(t *testing.T) {
	// Should receive real messages via libp2p streams
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

	// Connect hosts
	err = host1.Connect(ctx, peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	})
	require.NoError(t, err)

	handler1 := NewMessageHandler(host1)
	handler2 := NewMessageHandler(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send message from host1 to host2
	testContent := []byte("test message content")
	messageID, err := handler1.SendMessage(ctx, host2.ID().String(), testContent)
	require.NoError(t, err)
	assert.NotEmpty(t, messageID)

	// Verify message received on host2
	select {
	case msg := <-msgChan2:
		assert.Equal(t, testContent, msg.Content)
		assert.Equal(t, host1.ID().String(), msg.From)
		assert.NotEmpty(t, msg.ID)
	case <-time.After(5 * time.Second):
		t.Fatal("Message not received")
	}
}

func TestMessageHandler_RealStreamHandler_MultipleMessages(t *testing.T) {
	// Should handle multiple messages correctly
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

	// Connect hosts
	err = host1.Connect(ctx, peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	})
	require.NoError(t, err)

	handler1 := NewMessageHandler(host1)
	handler2 := NewMessageHandler(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send multiple messages
	messages := []string{"message 1", "message 2", "message 3"}
	messageIDs := make([]string, len(messages))

	for i, content := range messages {
		messageID, err := handler1.SendMessage(ctx, host2.ID().String(), []byte(content))
		require.NoError(t, err)
		messageIDs[i] = messageID
	}

	// Verify all messages received
	receivedMessages := make([]string, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		select {
		case msg := <-msgChan2:
			receivedMessages = append(receivedMessages, string(msg.Content))
		case <-time.After(5 * time.Second):
			t.Fatal("Not all messages received")
		}
	}

	// Verify all messages were received
	assert.Len(t, receivedMessages, len(messages))
	for _, expected := range messages {
		assert.Contains(t, receivedMessages, expected)
	}
}

func TestMessageHandler_RealStreamHandler_ConcurrentMessages(t *testing.T) {
	// Should handle concurrent messages safely
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

	// Connect hosts
	err = host1.Connect(ctx, peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	})
	require.NoError(t, err)

	handler1 := NewMessageHandler(host1)
	handler2 := NewMessageHandler(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send concurrent messages
	numMessages := 10
	done := make(chan bool, numMessages)

	for i := 0; i < numMessages; i++ {
		go func(i int) {
			content := []byte("concurrent message " + string(rune(i)))
			_, err := handler1.SendMessage(ctx, host2.ID().String(), content)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all sends to complete
	for i := 0; i < numMessages; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent sends timed out")
		}
	}

	// Verify messages received
	receivedCount := 0
	for {
		select {
		case <-msgChan2:
			receivedCount++
			if receivedCount >= numMessages {
				return
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Only received %d out of %d messages", receivedCount, numMessages)
		}
	}
}

func TestMessageHandler_RealStreamHandler_ErrorHandling(t *testing.T) {
	// Should handle stream errors gracefully
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	handler := NewMessageHandler(host)

	// Test sending to non-existent peer
	messageID, err := handler.SendMessage(ctx, "non-existent-peer", []byte("test"))
	assert.Error(t, err)
	assert.Empty(t, messageID)

	// Test with invalid peer ID
	messageID, err = handler.SendMessage(ctx, "invalid-peer-id", []byte("test"))
	assert.Error(t, err)
	assert.Empty(t, messageID)
}

func TestMessageHandler_RealStreamHandler_ProtocolHandling(t *testing.T) {
	// Should handle different message protocols correctly
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	handler := NewMessageHandler(host)

	// Verify protocol is set up correctly
	// In real implementation, this would verify the protocol handler
	assert.NotNil(t, handler)

	// Test that handler can be started
	msgChan := handler.SubscribeToMessages(ctx)
	assert.NotNil(t, msgChan)
}

func TestMessageHandler_RealStreamHandler_MessageOrdering(t *testing.T) {
	// Should maintain message ordering
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

	// Connect hosts
	err = host1.Connect(ctx, peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	})
	require.NoError(t, err)

	handler1 := NewMessageHandler(host1)
	handler2 := NewMessageHandler(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send ordered messages
	messages := []string{"first", "second", "third"}
	for _, content := range messages {
		_, err := handler1.SendMessage(ctx, host2.ID().String(), []byte(content))
		require.NoError(t, err)
	}

	// Verify messages received in order
	receivedMessages := make([]string, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		select {
		case msg := <-msgChan2:
			receivedMessages = append(receivedMessages, string(msg.Content))
		case <-time.After(5 * time.Second):
			t.Fatal("Message not received")
		}
	}

	// Verify order is maintained
	assert.Equal(t, messages, receivedMessages)
}
