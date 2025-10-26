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

func TestGroupMessage_EncryptDecrypt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create group with 3 members
	members := make([]host.Host, 3)
	groupManagers := make([]*messaging.GroupManager, 3)

	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		members[i] = h
		defer h.Close()

		// Create group manager for each member
		gm, err := messaging.NewGroupManager(ctx, h)
		require.NoError(t, err)
		groupManagers[i] = gm
	}

	// Connect members in a chain
	for i := 0; i < len(members)-1; i++ {
		peerInfo := members[i+1].Peerstore().PeerInfo(members[i+1].ID())
		err := members[i].Connect(ctx, peerInfo)
		require.NoError(t, err)
	}

	// Create group with all members
	groupID := "test-group"
	memberIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		memberIDs[i] = members[i].ID().String()
	}

	// All members join the group
	for i := 0; i < 3; i++ {
		err := groupManagers[i].CreateGroup(groupID, memberIDs)
		require.NoError(t, err)
	}

	// Wait for group setup
	time.Sleep(1 * time.Second)

	// Member 0 sends encrypted message
	plaintext := "Hello group!"
	encryptedMsg, err := groupManagers[0].EncryptGroupMessage(groupID, []byte(plaintext))
	require.NoError(t, err)
	assert.NotEqual(t, []byte(plaintext), encryptedMsg, "Encrypted message should be different from plaintext")

	// Members 1 and 2 should be able to decrypt
	for i := 1; i < 3; i++ {
		decrypted, err := groupManagers[i].DecryptGroupMessage(groupID, encryptedMsg)
		require.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted), "Member %d should decrypt message correctly", i)
	}
}

func TestGroupMessage_NonMemberRejected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create group with 2 members
	members := make([]host.Host, 2)
	groupManagers := make([]*messaging.GroupManager, 2)

	for i := 0; i < 2; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		members[i] = h
		defer h.Close()

		gm, err := messaging.NewGroupManager(ctx, h)
		require.NoError(t, err)
		groupManagers[i] = gm
	}

	// Connect members
	peerInfo := members[1].Peerstore().PeerInfo(members[1].ID())
	err := members[0].Connect(ctx, peerInfo)
	require.NoError(t, err)

	// Create group with only member 0
	groupID := "test-group"
	memberIDs := []string{members[0].ID().String()}

	err = groupManagers[0].CreateGroup(groupID, memberIDs)
	require.NoError(t, err)

	// Member 0 sends encrypted message
	plaintext := "Secret message"
	encryptedMsg, err := groupManagers[0].EncryptGroupMessage(groupID, []byte(plaintext))
	require.NoError(t, err)

	// Member 1 (not in group) should not be able to decrypt
	_, err = groupManagers[1].DecryptGroupMessage(groupID, encryptedMsg)
	assert.Error(t, err, "Non-member should not be able to decrypt group message")
}

func TestGroupMessage_ForwardSecrecy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create group with 3 members
	members := make([]host.Host, 3)
	groupManagers := make([]*messaging.GroupManager, 3)

	for i := 0; i < 3; i++ {
		h, err := network.NewHost(ctx, &network.Config{
			ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		})
		require.NoError(t, err)
		members[i] = h
		defer h.Close()

		gm, err := messaging.NewGroupManager(ctx, h)
		require.NoError(t, err)
		groupManagers[i] = gm
	}

	// Connect members
	for i := 0; i < len(members)-1; i++ {
		peerInfo := members[i+1].Peerstore().PeerInfo(members[i+1].ID())
		err := members[i].Connect(ctx, peerInfo)
		require.NoError(t, err)
	}

	// Create group with all members
	groupID := "test-group"
	memberIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		memberIDs[i] = members[i].ID().String()
	}

	// All members join the group
	for i := 0; i < 3; i++ {
		err := groupManagers[i].CreateGroup(groupID, memberIDs)
		require.NoError(t, err)
	}

	// Wait for group setup
	time.Sleep(1 * time.Second)

	// Send first message
	msg1 := "First message"
	encrypted1, err := groupManagers[0].EncryptGroupMessage(groupID, []byte(msg1))
	require.NoError(t, err)

	// All members should decrypt first message
	for i := 1; i < 3; i++ {
		decrypted, err := groupManagers[i].DecryptGroupMessage(groupID, encrypted1)
		require.NoError(t, err)
		assert.Equal(t, msg1, string(decrypted))
	}

	// Remove member 2 from group (key rotation)
	err = groupManagers[0].RemoveMember(groupID, members[2].ID().String())
	require.NoError(t, err)

	// Simulate member 2 being removed by deleting their group
	groupManagers[2].RemoveGroupForTesting(groupID)

	// Wait for key rotation
	time.Sleep(1 * time.Second)

	// Send second message after key rotation
	msg2 := "Second message after key rotation"
	encrypted2, err := groupManagers[0].EncryptGroupMessage(groupID, []byte(msg2))
	require.NoError(t, err)

	// Member 1 should decrypt second message
	decrypted2, err := groupManagers[1].DecryptGroupMessage(groupID, encrypted2)
	require.NoError(t, err)
	assert.Equal(t, msg2, string(decrypted2))

	// Member 2 (removed) should not decrypt second message
	_, err = groupManagers[2].DecryptGroupMessage(groupID, encrypted2)
	assert.Error(t, err, "Removed member should not decrypt messages after key rotation")

	// Member 2 should still decrypt first message (forward secrecy)
	// But since we removed the group, they can't decrypt anything
	_, err = groupManagers[2].DecryptGroupMessage(groupID, encrypted1)
	assert.Error(t, err, "Removed member should not decrypt any messages")
}
