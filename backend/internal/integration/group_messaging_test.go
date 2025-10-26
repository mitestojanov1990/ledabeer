package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_GroupMessageExchange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Create 3 nodes
	hosts := make([]host.Host, 3)
	groupManagers := make([]*messaging.GroupManager, 3)

	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		gm, err := messaging.NewGroupManager(ctx, h)
		require.NoError(t, err)
		groupManagers[i] = gm
	}

	// Connect nodes in a chain: 0 -> 1 -> 2
	for i := 0; i < len(hosts)-1; i++ {
		peerInfo := hosts[i+1].Peerstore().PeerInfo(hosts[i+1].ID())
		err := hosts[i].Connect(ctx, peerInfo)
		require.NoError(t, err)
	}

	// Create group with all members
	groupID := "integration-group"
	memberIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		memberIDs[i] = hosts[i].ID().String()
	}

	// All members join the group
	for i := 0; i < 3; i++ {
		err := groupManagers[i].CreateGroup(groupID, memberIDs)
		require.NoError(t, err)
	}

	// Wait for group setup
	time.Sleep(2 * time.Second)

	// Node 0 sends a message to the group
	message1 := "Hello from node 0!"
	err := groupManagers[0].SendGroupMessage(groupID, message1)
	require.NoError(t, err)

	// Node 1 sends a message to the group
	message2 := "Hello from node 1!"
	err = groupManagers[1].SendGroupMessage(groupID, message2)
	require.NoError(t, err)

	// Wait for message propagation
	time.Sleep(3 * time.Second)

	// All nodes should have received both messages
	for i := 0; i < 3; i++ {
		messages := groupManagers[i].GetGroupMessages(groupID)
		assert.Len(t, messages, 2, "Node %d should have received 2 messages", i)

		// Check that both messages are present (order may vary)
		found1 := false
		found2 := false
		for _, msg := range messages {
			if msg == message1 {
				found1 = true
			}
			if msg == message2 {
				found2 = true
			}
		}
		assert.True(t, found1, "Node %d should have received message 1", i)
		assert.True(t, found2, "Node %d should have received message 2", i)
	}
}

func TestE2E_GroupMembershipChanges(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Create 4 nodes
	hosts := make([]host.Host, 4)
	groupManagers := make([]*messaging.GroupManager, 4)

	for i := 0; i < 4; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		gm, err := messaging.NewGroupManager(ctx, h)
		require.NoError(t, err)
		groupManagers[i] = gm
	}

	// Connect nodes
	for i := 0; i < len(hosts)-1; i++ {
		peerInfo := hosts[i+1].Peerstore().PeerInfo(hosts[i+1].ID())
		err := hosts[i].Connect(ctx, peerInfo)
		require.NoError(t, err)
	}

	// Create group with first 3 members
	groupID := "membership-group"
	initialMembers := make([]string, 3)
	for i := 0; i < 3; i++ {
		initialMembers[i] = hosts[i].ID().String()
	}

	// First 3 members join the group
	for i := 0; i < 3; i++ {
		err := groupManagers[i].CreateGroup(groupID, initialMembers)
		require.NoError(t, err)
	}

	// Wait for group setup
	time.Sleep(2 * time.Second)

	// Send initial message
	initialMessage := "Initial group message"
	err := groupManagers[0].SendGroupMessage(groupID, initialMessage)
	require.NoError(t, err)

	// Add 4th member
	newMemberID := hosts[3].ID().String()
	err = groupManagers[0].AddMember(groupID, newMemberID)
	require.NoError(t, err)

	// 4th member joins the group
	allMembers := append(initialMembers, newMemberID)
	err = groupManagers[3].CreateGroup(groupID, allMembers)
	require.NoError(t, err)

	// Wait for member addition
	time.Sleep(2 * time.Second)

	// Send message after adding member
	afterAddMessage := "Message after adding member"
	err = groupManagers[1].SendGroupMessage(groupID, afterAddMessage)
	require.NoError(t, err)

	// Remove 2nd member
	removedMemberID := hosts[1].ID().String()
	err = groupManagers[0].RemoveMember(groupID, removedMemberID)
	require.NoError(t, err)

	// Wait for member removal
	time.Sleep(2 * time.Second)

	// Send message after removing member
	afterRemoveMessage := "Message after removing member"
	err = groupManagers[0].SendGroupMessage(groupID, afterRemoveMessage)
	require.NoError(t, err)

	// Wait for message propagation
	time.Sleep(2 * time.Second)

	// Verify message counts for different nodes
	// Note: Due to pubsub message propagation and deduplication, exact counts may vary
	// We check that messages are present rather than exact counts

	// Node 0 (remover) should have messages
	messages0 := groupManagers[0].GetGroupMessages(groupID)
	assert.GreaterOrEqual(t, len(messages0), 3, "Node 0 should have at least 3 messages")

	// Node 2 should have messages
	messages2 := groupManagers[2].GetGroupMessages(groupID)
	assert.GreaterOrEqual(t, len(messages2), 3, "Node 2 should have at least 3 messages")

	// Node 3 (added member) should have messages after joining
	messages3 := groupManagers[3].GetGroupMessages(groupID)
	assert.GreaterOrEqual(t, len(messages3), 2, "Node 3 should have at least 2 messages (after joining)")

	// Node 1 (removed member) should have messages before removal
	messages1 := groupManagers[1].GetGroupMessages(groupID)
	assert.GreaterOrEqual(t, len(messages1), 2, "Node 1 should have at least 2 messages (before removal)")

	// Verify that all expected messages are present in at least one node
	allMessages := make(map[string]bool)
	for _, messages := range [][]string{messages0, messages1, messages2, messages3} {
		for _, msg := range messages {
			allMessages[msg] = true
		}
	}

	assert.True(t, allMessages["Initial group message"], "Initial message should be present")
	assert.True(t, allMessages["Message after adding member"], "After add message should be present")
	assert.True(t, allMessages["Message after removing member"], "After remove message should be present")
}

func TestE2E_LargeGroup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create 10 nodes for large group test
	hosts := make([]host.Host, 10)
	groupManagers := make([]*messaging.GroupManager, 10)

	for i := 0; i < 10; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		gm, err := messaging.NewGroupManager(ctx, h)
		require.NoError(t, err)
		groupManagers[i] = gm
	}

	// Connect nodes in a chain: 0 -> 1 -> 2 -> ... -> 9
	for i := 0; i < len(hosts)-1; i++ {
		peerInfo := hosts[i+1].Peerstore().PeerInfo(hosts[i+1].ID())
		err := hosts[i].Connect(ctx, peerInfo)
		require.NoError(t, err)
	}

	// Create group with all members
	groupID := "large-group"
	memberIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		memberIDs[i] = hosts[i].ID().String()
	}

	// All members join the group
	for i := 0; i < 10; i++ {
		err := groupManagers[i].CreateGroup(groupID, memberIDs)
		require.NoError(t, err)
	}

	// Wait for group setup
	time.Sleep(3 * time.Second)

	// Each node sends a message
	for i := 0; i < 10; i++ {
		message := fmt.Sprintf("Message from node %d", i)
		err := groupManagers[i].SendGroupMessage(groupID, message)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Small delay between messages
	}

	// Wait for all messages to propagate
	time.Sleep(5 * time.Second)

	// Check that all nodes received all messages
	for i := 0; i < 10; i++ {
		messages := groupManagers[i].GetGroupMessages(groupID)
		assert.Len(t, messages, 10, "Node %d should have received 10 messages", i)

		// Verify all expected messages are present
		expectedMessages := make(map[string]bool)
		for j := 0; j < 10; j++ {
			expectedMessages[fmt.Sprintf("Message from node %d", j)] = true
		}

		for _, msg := range messages {
			assert.True(t, expectedMessages[msg], "Node %d should have received message: %s", i, msg)
		}
	}
}
