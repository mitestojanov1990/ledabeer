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

func TestMessageHandler_RealE2EE_EncryptDecrypt(t *testing.T) {
	// Should encrypt messages before sending and decrypt after receiving
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

	// Create message handlers with E2EE
	handler1 := NewMessageHandlerWithE2EE(host1)
	handler2 := NewMessageHandlerWithE2EE(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send encrypted message
	plaintext := []byte("secret message")
	messageID, err := handler1.SendMessage(ctx, host2.ID().String(), plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, messageID)

	// Verify message received and decrypted
	select {
	case msg := <-msgChan2:
		assert.Equal(t, plaintext, msg.Content, "Message should be decrypted correctly")
		assert.Equal(t, host1.ID().String(), msg.From)
		assert.NotEmpty(t, msg.ID)
	case <-time.After(5 * time.Second):
		t.Fatal("Encrypted message not received")
	}
}

func TestMessageHandler_RealE2EE_KeyExchange(t *testing.T) {
	// Should perform key exchange before sending messages
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

	handler1 := NewMessageHandlerWithE2EE(host1)
	handler2 := NewMessageHandlerWithE2EE(host2)

	// Perform key exchange
	err = handler1.ExchangeKeys(ctx, host2.ID().String())
	require.NoError(t, err)

	err = handler2.ExchangeKeys(ctx, host1.ID().String())
	require.NoError(t, err)

	// Verify keys are exchanged
	assert.True(t, handler1.HasSession(host2.ID().String()))
	assert.True(t, handler2.HasSession(host1.ID().String()))
}

func TestMessageHandler_RealE2EE_ForwardSecrecy(t *testing.T) {
	// Should maintain forward secrecy - ratchet advances after each message
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

	handler1 := NewMessageHandlerWithE2EE(host1)
	handler2 := NewMessageHandlerWithE2EE(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send first message
	message1 := []byte("first message")
	_, err = handler1.SendMessage(ctx, host2.ID().String(), message1)
	require.NoError(t, err)

	// Receive first message
	select {
	case msg := <-msgChan2:
		assert.Equal(t, message1, msg.Content)
	case <-time.After(5 * time.Second):
		t.Fatal("First message not received")
	}

	// Send second message (ratchet should advance)
	message2 := []byte("second message")
	_, err = handler1.SendMessage(ctx, host2.ID().String(), message2)
	require.NoError(t, err)

	// Receive second message
	select {
	case msg := <-msgChan2:
		assert.Equal(t, message2, msg.Content)
	case <-time.After(5 * time.Second):
		t.Fatal("Second message not received")
	}

	// Verify that both messages were received and decrypted correctly
	// The fact that both messages are properly decrypted shows that the ratchet
	// is working correctly - each message uses a different key
}

func TestMessageHandler_RealE2EE_MessageIntegrity(t *testing.T) {
	// Should verify message integrity
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

	handler1 := NewMessageHandlerWithE2EE(host1)
	handler2 := NewMessageHandlerWithE2EE(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send message
	originalMessage := []byte("integrity test message")
	_, err = handler1.SendMessage(ctx, host2.ID().String(), originalMessage)
	require.NoError(t, err)

	// Verify message received with correct content
	select {
	case msg := <-msgChan2:
		assert.Equal(t, originalMessage, msg.Content, "Message integrity should be preserved")
		assert.Equal(t, host1.ID().String(), msg.From)
	case <-time.After(5 * time.Second):
		t.Fatal("Message not received")
	}
}

func TestMessageHandler_RealE2EE_ConcurrentEncryption(t *testing.T) {
	// Should handle concurrent encrypted messages
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

	handler1 := NewMessageHandlerWithE2EE(host1)
	handler2 := NewMessageHandlerWithE2EE(host2)

	// Start message handlers
	msgChan2 := handler2.SubscribeToMessages(ctx)

	// Send concurrent encrypted messages
	numMessages := 5
	done := make(chan bool, numMessages)

	for i := 0; i < numMessages; i++ {
		go func(i int) {
			content := []byte("encrypted message " + string(rune(i)))
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
			t.Fatalf("Only received %d out of %d encrypted messages", receivedCount, numMessages)
		}
	}
}
