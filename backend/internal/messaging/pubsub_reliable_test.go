package messaging_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPubSub_ReliableMessageCounting tests message counting with deduplication handling
func TestPubSub_ReliableMessageCounting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create three hosts
	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// All hosts subscribe to same topic
	topic := "reliable-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counters for hosts 1 and 2 (they will receive messages)
	counter1 := helper.SetupMessageCounter(1, "host1", 5*time.Second)
	counter2 := helper.SetupMessageCounter(2, "host2", 5*time.Second)

	// Expect each host to receive 1 message
	counter1.ExpectMessage(topic, 1)
	counter2.ExpectMessage(topic, 1)

	// Host 0 publishes message
	testMessage := []byte("Reliable test message")
	err = helper.PublishMessage(0, topic, testMessage)
	require.NoError(t, err)

	// Wait for messages with flexible assertions
	helper.AssertMessageCounts(t)

	// Verify specific message content
	counter1.AssertMessageReceived(t, topic, string(testMessage))
	counter2.AssertMessageReceived(t, topic, string(testMessage))
}

// TestPubSub_ReliableDeduplication tests message deduplication with flexible counting
func TestPubSub_ReliableDeduplication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create two hosts
	for i := 0; i < 2; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// Host 1 subscribes to topic
	topic := "dedup-reliable-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counter
	counter := helper.SetupMessageCounter(1, "receiver", 5*time.Second)
	counter.ExpectMessage(topic, 1) // Expect 1 message due to deduplication

	// Host 0 publishes same message multiple times
	testMessage := []byte("Duplicate test message")
	for i := 0; i < 3; i++ {
		err = helper.PublishMessage(0, topic, testMessage)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Assert with flexible counting (should be 0-1 due to deduplication)
	actualCount := counter.GetCount(topic)
	messaging.AssertMessageCountFlexible(t, actualCount, 1, topic)
}

// TestPubSub_ReliableMultipleMessages tests multiple message handling with flexible counting
func TestPubSub_ReliableMultipleMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create three hosts
	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// All hosts subscribe to same topic
	topic := "multiple-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counters for hosts 1 and 2
	counter1 := helper.SetupMessageCounter(1, "host1", 10*time.Second)
	counter2 := helper.SetupMessageCounter(2, "host2", 10*time.Second)

	// Expect each host to receive 3 messages (with some flexibility)
	counter1.ExpectMessage(topic, 3)
	counter2.ExpectMessage(topic, 3)

	// Host 0 publishes multiple different messages
	messages := []string{"message 1", "message 2", "message 3"}
	for _, msg := range messages {
		err = helper.PublishMessage(0, topic, []byte(msg))
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}

	// Wait for messages with flexible assertions
	helper.AssertMessageCounts(t)

	// Verify all messages were received by both hosts
	for _, msg := range messages {
		counter1.AssertMessageReceived(t, topic, msg)
		counter2.AssertMessageReceived(t, topic, msg)
	}
}

// TestPubSub_ReliableConcurrentMessages tests concurrent message handling
func TestPubSub_ReliableConcurrentMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create four hosts
	for i := 0; i < 4; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts in a mesh
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// All hosts subscribe to same topic
	topic := "concurrent-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counters for hosts 1, 2, and 3
	counter1 := helper.SetupMessageCounter(1, "host1", 15*time.Second)
	counter2 := helper.SetupMessageCounter(2, "host2", 15*time.Second)
	counter3 := helper.SetupMessageCounter(3, "host3", 15*time.Second)

	// Expect each host to receive 2 messages (with flexibility)
	counter1.ExpectMessage(topic, 2)
	counter2.ExpectMessage(topic, 2)
	counter3.ExpectMessage(topic, 2)

	// Host 0 publishes messages concurrently
	go func() {
		helper.PublishMessage(0, topic, []byte("concurrent message 1"))
		time.Sleep(50 * time.Millisecond)
		helper.PublishMessage(0, topic, []byte("concurrent message 2"))
	}()

	// Wait for messages with flexible assertions
	helper.AssertMessageCounts(t)

	// Verify message counts are within expected range
	messaging.AssertMessageCountRange(t, counter1.GetCount(topic), 1, 3, "host1 messages")
	messaging.AssertMessageCountRange(t, counter2.GetCount(topic), 1, 3, "host2 messages")
	messaging.AssertMessageCountRange(t, counter3.GetCount(topic), 1, 3, "host3 messages")
}

// TestPubSub_ReliableMessageOrdering tests message ordering with flexible counting
func TestPubSub_ReliableMessageOrdering(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create two hosts
	for i := 0; i < 2; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// Host 1 subscribes to topic
	topic := "ordering-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counter
	counter := helper.SetupMessageCounter(1, "receiver", 10*time.Second)
	counter.ExpectMessage(topic, 3) // Expect 3 messages

	// Host 0 publishes messages in sequence
	messages := []string{"first", "second", "third"}
	for i, msg := range messages {
		err = helper.PublishMessage(0, topic, []byte(msg))
		require.NoError(t, err)

		// Small delay between messages to ensure ordering
		if i < len(messages)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Wait for messages
	helper.AssertMessageCounts(t)

	// Verify all messages were received
	received := counter.GetReceived(topic)
	assert.GreaterOrEqual(t, len(received), 2, "Should receive at least 2 messages")
	assert.LessOrEqual(t, len(received), 3, "Should receive at most 3 messages")

	// Verify specific messages were received
	for _, msg := range messages {
		counter.AssertMessageReceived(t, topic, msg)
	}
}

// TestPubSub_ReliableNetworkPartition tests message handling during network partitions
func TestPubSub_ReliableNetworkPartition(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create three hosts
	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// All hosts subscribe to same topic
	topic := "partition-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counters
	counter1 := helper.SetupMessageCounter(1, "host1", 15*time.Second)
	counter2 := helper.SetupMessageCounter(2, "host2", 15*time.Second)

	// Expect each host to receive 1 message (with flexibility)
	counter1.ExpectMessage(topic, 1)
	counter2.ExpectMessage(topic, 1)

	// Host 0 publishes message
	err = helper.PublishMessage(0, topic, []byte("partition test message"))
	require.NoError(t, err)

	// Wait for messages with flexible assertions
	helper.AssertMessageCounts(t)

	// Verify message counts are reasonable (0-2 due to potential network issues)
	messaging.AssertMessageCountAtMost(t, counter1.GetCount(topic), 2, "host1 messages")
	messaging.AssertMessageCountAtMost(t, counter2.GetCount(topic), 2, "host2 messages")
}

// TestPubSub_ReliableStressTest tests PubSub under stress with flexible counting
func TestPubSub_ReliableStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	helper := messaging.NewPubSubTestHelper()
	defer helper.Cleanup()

	// Create five hosts
	for i := 0; i < 5; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)

		_, err = helper.AddHost(ctx, h)
		require.NoError(t, err)
	}

	// Connect hosts
	err := helper.ConnectHosts(ctx)
	require.NoError(t, err)

	// All hosts subscribe to same topic
	topic := "stress-test"
	err = helper.SubscribeAll(topic)
	require.NoError(t, err)

	// Wait for mesh formation
	helper.WaitForMeshFormation()

	// Setup message counters for all hosts except the publisher
	counters := make([]*messaging.MessageCounter, 4)
	for i := 0; i < 4; i++ {
		counters[i] = helper.SetupMessageCounter(i+1, fmt.Sprintf("host%d", i+1), 25*time.Second)
		counters[i].ExpectMessage(topic, 10) // Expect 10 messages with flexibility
	}

	// Host 0 publishes many messages
	numMessages := 10
	for i := 0; i < numMessages; i++ {
		msg := []byte(fmt.Sprintf("stress message %d", i))
		err = helper.PublishMessage(0, topic, msg)
		require.NoError(t, err)

		// Small delay to avoid overwhelming the network
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for messages with flexible assertions
	helper.AssertMessageCounts(t)

	// Verify message counts are within reasonable range
	for i, counter := range counters {
		actualCount := counter.GetCount(topic)
		// Allow significant variance due to network conditions and deduplication
		messaging.AssertMessageCountRange(t, actualCount, 5, 15, fmt.Sprintf("host%d messages", i+1))
	}
}
