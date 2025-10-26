package messaging

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/network"
	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupManager_RealPubSub_SendGroupMessage(t *testing.T) {
	// Should send group message via real PubSub publishing
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-1"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Send group message via PubSub
	messageID, err := groupManager.SendGroupMessage(ctx, groupID, []byte("group message content"))
	require.NoError(t, err)
	assert.NotEmpty(t, messageID)

	// Verify message was published to PubSub
	assert.True(t, groupManager.HasPublishedToGroup(groupID, messageID))
}

func TestGroupManager_RealPubSub_ReceiveGroupMessage(t *testing.T) {
	// Should receive group messages via PubSub subscription
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-2"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Subscribe to group messages
	msgChan := groupManager.SubscribeToGroupMessages(ctx, groupID)

	// Send message to group
	messageID, err := groupManager.SendGroupMessage(ctx, groupID, []byte("received group message"))
	require.NoError(t, err)

	// Verify message received
	select {
	case msg := <-msgChan:
		assert.Equal(t, messageID, msg.ID)
		assert.Equal(t, []byte("received group message"), msg.Content)
		assert.Equal(t, groupID, msg.GroupID)
	case <-time.After(5 * time.Second):
		t.Fatal("Group message not received")
	}
}

func TestGroupManager_RealPubSub_MultipleSubscribers(t *testing.T) {
	// Should deliver message to multiple subscribers
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-3"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Create multiple subscribers
	subscriber1 := groupManager.SubscribeToGroupMessages(ctx, groupID)
	subscriber2 := groupManager.SubscribeToGroupMessages(ctx, groupID)

	// Send message
	messageID, err := groupManager.SendGroupMessage(ctx, groupID, []byte("multi-subscriber message"))
	require.NoError(t, err)

	// Verify both subscribers receive the message
	received1 := false
	received2 := false

	for i := 0; i < 2; i++ {
		select {
		case msg := <-subscriber1:
			assert.Equal(t, messageID, msg.ID)
			received1 = true
		case msg := <-subscriber2:
			assert.Equal(t, messageID, msg.ID)
			received2 = true
		case <-time.After(5 * time.Second):
			t.Fatal("Not all subscribers received message")
		}
	}

	assert.True(t, received1, "Subscriber 1 should receive message")
	assert.True(t, received2, "Subscriber 2 should receive message")
}

func TestGroupManager_RealPubSub_GroupMessageHistory(t *testing.T) {
	// Should store group messages in IPFS and retrieve history
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	ipfsNode, err := storage.NewIPFSNode(ctx, &storage.IPFSConfig{})
	require.NoError(t, err)
	defer ipfsNode.Close()

	groupManager, err := NewGroupManagerWithStorage(ctx, host, ipfsNode)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-4"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Send multiple messages
	_, err = groupManager.SendGroupMessage(ctx, groupID, []byte("first group message"))
	require.NoError(t, err)

	_, err = groupManager.SendGroupMessage(ctx, groupID, []byte("second group message"))
	require.NoError(t, err)

	// Retrieve group message history
	messages, err := groupManager.GetGroupMessageHistory(ctx, groupID, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(messages), 2, "Should have at least 2 messages")

	// Verify messages contain the expected content
	messageContents := make([]string, len(messages))
	for i, msg := range messages {
		messageContents[i] = string(msg.Content)
	}
	assert.Contains(t, messageContents, "first group message")
	assert.Contains(t, messageContents, "second group message")
}

func TestGroupManager_RealPubSub_MessageEncryption(t *testing.T) {
	// Should encrypt group messages before publishing
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-5"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Send encrypted message
	plaintext := []byte("sensitive group message")
	messageID, err := groupManager.SendGroupMessage(ctx, groupID, plaintext)
	require.NoError(t, err)

	// Verify message was encrypted before publishing
	publishedData := groupManager.GetPublishedMessage(groupID, messageID)
	assert.NotEqual(t, plaintext, publishedData) // Should be encrypted
	assert.NotEmpty(t, publishedData)            // Should contain encrypted data
}

func TestGroupManager_RealPubSub_GroupMemberManagement(t *testing.T) {
	// Should manage group members and their PubSub subscriptions
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-6"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Add members
	member1 := "peer1"
	member2 := "peer2"
	err = groupManager.AddMemberWithContext(ctx, groupID, member1)
	require.NoError(t, err)
	err = groupManager.AddMemberWithContext(ctx, groupID, member2)
	require.NoError(t, err)

	// Verify members are added (excluding host ID)
	members, err := groupManager.GetGroupMembers(ctx, groupID)
	require.NoError(t, err)

	// Filter out the host ID to get only the added members
	addedMembers := make([]string, 0)
	for _, member := range members {
		if member != host.ID().String() {
			addedMembers = append(addedMembers, member)
		}
	}

	assert.Len(t, addedMembers, 2)
	assert.Contains(t, addedMembers, member1)
	assert.Contains(t, addedMembers, member2)

	// Remove member
	err = groupManager.RemoveMemberWithContext(ctx, groupID, member1)
	require.NoError(t, err)

	// Verify member removed (excluding host ID)
	members, err = groupManager.GetGroupMembers(ctx, groupID)
	require.NoError(t, err)

	// Filter out the host ID to get only the added members
	addedMembers = make([]string, 0)
	for _, member := range members {
		if member != host.ID().String() {
			addedMembers = append(addedMembers, member)
		}
	}

	assert.Len(t, addedMembers, 1)
	assert.NotContains(t, addedMembers, member1)
	assert.Contains(t, addedMembers, member2)
}

func TestGroupManager_RealPubSub_MessageOrdering(t *testing.T) {
	// Should maintain message ordering within group
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create group
	groupID := "test-group-7"
	err = groupManager.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Subscribe to messages
	msgChan := groupManager.SubscribeToGroupMessages(ctx, groupID)

	// Send multiple messages in sequence
	messages := []string{"first", "second", "third"}
	messageIDs := make([]string, len(messages))

	for i, content := range messages {
		messageID, err := groupManager.SendGroupMessage(ctx, groupID, []byte(content))
		require.NoError(t, err)
		messageIDs[i] = messageID
	}

	// Verify messages received in order
	receivedMessages := make([]string, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		select {
		case msg := <-msgChan:
			receivedMessages = append(receivedMessages, string(msg.Content))
		case <-time.After(5 * time.Second):
			t.Fatalf("Only received %d out of %d messages", len(receivedMessages), len(messages))
		}
	}

	assert.Equal(t, messages, receivedMessages, "Messages should be received in order")
}
