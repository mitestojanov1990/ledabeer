package messaging_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub_Subscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create host with pubsub
	h, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h.Close()

	pubsub, err := messaging.NewPubSub(ctx, h)
	require.NoError(t, err)

	// Subscribe to topic
	topic := "test-topic"
	subscription, err := pubsub.Subscribe(topic)
	require.NoError(t, err)
	assert.NotNil(t, subscription, "Should return valid subscription")

	// Verify subscription is active
	assert.True(t, pubsub.IsSubscribed(topic), "Should be subscribed to topic")
}

func TestPubSub_PublishReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create two hosts
	h1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h1.Close()

	h2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h2.Close()

	// Connect hosts
	h2Addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()
	peerInfo, err := network.ParseAddrInfo(h2Addr)
	require.NoError(t, err)
	err = h1.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	// Create pubsub instances
	pubsub1, err := messaging.NewPubSub(ctx, h1)
	require.NoError(t, err)

	pubsub2, err := messaging.NewPubSub(ctx, h2)
	require.NoError(t, err)

	// Host2 subscribes to topic
	topic := "test-topic"
	_, err = pubsub2.Subscribe(topic)
	require.NoError(t, err)

	// Channel to receive messages
	received := make(chan []byte, 1)
	pubsub2.SetMessageHandler(func(topicName string, data []byte) {
		if topicName == topic {
			received <- data
		}
	})

	// Wait for mesh formation
	time.Sleep(2 * time.Second)

	// Host1 publishes message
	testMessage := []byte("Hello from host1!")
	err = pubsub1.Publish(topic, testMessage)
	require.NoError(t, err)

	// Verify host2 receives message
	select {
	case msg := <-received:
		assert.Equal(t, testMessage, msg, "Should receive published message")
	case <-time.After(5 * time.Second):
		t.Fatal("Message not received within timeout")
	}
}

func TestPubSub_MultipleSubscribers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create three hosts
	hosts := make([]host.Host, 3)
	pubsubs := make([]*messaging.PubSub, 3)

	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		ps, err := messaging.NewPubSub(ctx, h)
		require.NoError(t, err)
		pubsubs[i] = ps
	}

	// Connect hosts in a chain: 0 -> 1 -> 2
	for i := 0; i < len(hosts)-1; i++ {
		peerInfo := hosts[i+1].Peerstore().PeerInfo(hosts[i+1].ID())
		err := hosts[i].Connect(ctx, peerInfo)
		require.NoError(t, err)
	}

	// All hosts subscribe to same topic
	topic := "broadcast-topic"
	for i := 0; i < 3; i++ {
		_, err := pubsubs[i].Subscribe(topic)
		require.NoError(t, err)
	}

	// Wait for mesh formation
	time.Sleep(2 * time.Second)

	// Setup message handlers
	receivedCount := make(chan int, 3)
	for i := 1; i < 3; i++ { // Hosts 1 and 2 listen
		pubsubs[i].SetMessageHandler(func(topicName string, data []byte) {
			if topicName == topic {
				receivedCount <- 1
			}
		})
	}

	// Host 0 publishes message
	testMessage := []byte("Broadcast message!")
	err := pubsubs[0].Publish(topic, testMessage)
	require.NoError(t, err)

	// Wait for both subscribers to receive
	timeout := time.After(10 * time.Second)
	received := 0
	for received < 2 {
		select {
		case <-receivedCount:
			received++
		case <-timeout:
			t.Fatalf("Only %d/2 subscribers received message", received)
		}
	}
}

func TestPubSub_Unsubscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create host with pubsub
	h, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h.Close()

	pubsub, err := messaging.NewPubSub(ctx, h)
	require.NoError(t, err)

	// Subscribe to topic
	topic := "test-topic"
	_, err = pubsub.Subscribe(topic)
	require.NoError(t, err)
	assert.True(t, pubsub.IsSubscribed(topic), "Should be subscribed")

	// Unsubscribe
	err = pubsub.Unsubscribe(topic)
	require.NoError(t, err)
	assert.False(t, pubsub.IsSubscribed(topic), "Should not be subscribed after unsubscribe")
}

func TestPubSub_TopicValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create host with pubsub
	h, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h.Close()

	pubsub, err := messaging.NewPubSub(ctx, h)
	require.NoError(t, err)

	// Test invalid topic names
	invalidTopics := []string{
		"",                   // empty
		"topic with spaces",  // spaces
		"topic@invalid",      // special chars
		"topic/with/slashes", // slashes
		"topic.with.dots",    // dots
	}

	for _, topic := range invalidTopics {
		_, err := pubsub.Subscribe(topic)
		assert.Error(t, err, "Should reject invalid topic: %s", topic)
	}

	// Test valid topic names
	validTopics := []string{
		"valid-topic",
		"valid_topic",
		"valid123",
		"topic-name_123",
	}

	for _, topic := range validTopics {
		_, err := pubsub.Subscribe(topic)
		assert.NoError(t, err, "Should accept valid topic: %s", topic)
	}
}

func TestPubSub_MessageDeduplication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create two hosts
	h1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h1.Close()

	h2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h2.Close()

	// Connect hosts
	h2Addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()
	peerInfo, err := network.ParseAddrInfo(h2Addr)
	require.NoError(t, err)
	err = h1.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	// Create pubsub instances
	pubsub1, err := messaging.NewPubSub(ctx, h1)
	require.NoError(t, err)

	pubsub2, err := messaging.NewPubSub(ctx, h2)
	require.NoError(t, err)

	// Host2 subscribes to topic
	topic := "dedup-test"
	_, err = pubsub2.Subscribe(topic)
	require.NoError(t, err)

	// Count received messages
	receivedCount := 0
	pubsub2.SetMessageHandler(func(topicName string, data []byte) {
		if topicName == topic {
			receivedCount++
		}
	})

	// Wait for mesh formation
	time.Sleep(2 * time.Second)

	// Host1 publishes same message multiple times
	testMessage := []byte("Duplicate test message")
	for i := 0; i < 3; i++ {
		err = pubsub1.Publish(topic, testMessage)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Should only receive message once due to deduplication
	assert.Equal(t, 1, receivedCount, "Should receive message only once due to deduplication")
}
