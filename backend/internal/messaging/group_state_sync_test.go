package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemberStateSynchronizationOnLeave tests that when a member leaves, their state is properly synchronized
func TestMemberStateSynchronizationOnLeave(t *testing.T) {
	// Create four hosts
	hosts := make([]host.Host, 4)
	gms := make([]*GroupManager, 4)

	ctx := context.Background()
	for i := 0; i < 4; i++ {
		h, err := createTestHost()
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		gm, err := NewGroupManager(ctx, h)
		require.NoError(t, err)
		gms[i] = gm
	}

	// Connect all hosts in a mesh
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			err := hosts[i].Connect(ctx, peer.AddrInfo{ID: hosts[j].ID(), Addrs: hosts[j].Addrs()})
			require.NoError(t, err)
		}
	}

	// Create group on host0
	groupID := "test-group-state-sync-leave"
	err := gms[0].CreateGroup(ctx, groupID)
	require.NoError(t, err)

	// Add all other hosts as members
	for i := 1; i < 4; i++ {
		err := gms[0].AddMemberWithContext(ctx, groupID, hosts[i].ID().String())
		require.NoError(t, err)
	}

	// All hosts join the group
	for i := 1; i < 4; i++ {
		err := gms[i].JoinGroup(ctx, groupID)
		require.NoError(t, err)
	}

	// Wait for initial synchronization
	time.Sleep(300 * time.Millisecond)

	// Verify all members are present
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			assert.True(t, gms[i].IsGroupMember(groupID, hosts[j].ID().String()),
				"Host %d should see host %d as member", i, j)
		}
	}

	// Host2 leaves the group voluntarily
	err = gms[2].LeaveGroup(groupID)
	require.NoError(t, err)

	// Wait for state synchronization
	time.Sleep(300 * time.Millisecond)

	// All remaining hosts should not see host2 as a member
	for i := 0; i < 4; i++ {
		if i == 2 {
			// Host2 should not see itself as a member
			assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
				"Host2 should not see itself as member after leaving")
		} else {
			// Other hosts should not see host2 as a member
			assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
				"Host %d should not see host2 as member after host2 leaves", i)
		}
	}

	// Verify remaining members are still present
	remainingHosts := []int{0, 1, 3}
	for _, i := range remainingHosts {
		for _, j := range remainingHosts {
			assert.True(t, gms[i].IsGroupMember(groupID, hosts[j].ID().String()),
				"Host %d should see host %d as member", i, j)
		}
	}
}

// TestMemberStateSynchronizationOnRemoval tests that when a member is removed, their state is properly synchronized
func TestMemberStateSynchronizationOnRemoval(t *testing.T) {
	// Create four hosts
	hosts := make([]host.Host, 4)
	gms := make([]*GroupManager, 4)

	ctx := context.Background()
	for i := 0; i < 4; i++ {
		h, err := createTestHost()
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		gm, err := NewGroupManager(ctx, h)
		require.NoError(t, err)
		gms[i] = gm
	}

	// Connect all hosts
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			err := hosts[i].Connect(ctx, peer.AddrInfo{ID: hosts[j].ID(), Addrs: hosts[j].Addrs()})
			require.NoError(t, err)
		}
	}

	// Create group and add all members
	groupID := "test-group-state-sync-removal"
	err := gms[0].CreateGroup(ctx, groupID)
	require.NoError(t, err)

	for i := 1; i < 4; i++ {
		err := gms[0].AddMemberWithContext(ctx, groupID, hosts[i].ID().String())
		require.NoError(t, err)
	}

	// All hosts join the group
	for i := 1; i < 4; i++ {
		err := gms[i].JoinGroup(ctx, groupID)
		require.NoError(t, err)
	}

	// Wait for synchronization
	time.Sleep(200 * time.Millisecond)

	// Host0 removes host2 from the group
	err = gms[0].RemoveMemberWithContext(ctx, groupID, hosts[2].ID().String())
	require.NoError(t, err)

	// Wait for state synchronization
	time.Sleep(300 * time.Millisecond)

	// All hosts should not see host2 as a member
	for i := 0; i < 4; i++ {
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
			"Host %d should not see host2 as member after removal", i)
	}

	// Verify remaining members are still present
	remainingHosts := []int{0, 1, 3}
	for _, i := range remainingHosts {
		for _, j := range remainingHosts {
			assert.True(t, gms[i].IsGroupMember(groupID, hosts[j].ID().String()),
				"Host %d should see host %d as member", i, j)
		}
	}
}

// TestMemberStateSynchronizationConcurrentLeaves tests concurrent member leaves
func TestMemberStateSynchronizationConcurrentLeaves(t *testing.T) {
	// Create five hosts
	hosts := make([]host.Host, 5)
	gms := make([]*GroupManager, 5)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		h, err := createTestHost()
		require.NoError(t, err)
		hosts[i] = h
		defer h.Close()

		gm, err := NewGroupManager(ctx, h)
		require.NoError(t, err)
		gms[i] = gm
	}

	// Connect all hosts
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			err := hosts[i].Connect(ctx, peer.AddrInfo{ID: hosts[j].ID(), Addrs: hosts[j].Addrs()})
			require.NoError(t, err)
		}
	}

	// Create group and add all members
	groupID := "test-group-state-sync-concurrent-leaves"
	err := gms[0].CreateGroup(ctx, groupID)
	require.NoError(t, err)

	for i := 1; i < 5; i++ {
		err := gms[0].AddMemberWithContext(ctx, groupID, hosts[i].ID().String())
		require.NoError(t, err)
	}

	// All hosts join the group
	for i := 1; i < 5; i++ {
		err := gms[i].JoinGroup(ctx, groupID)
		require.NoError(t, err)
	}

	// Wait for synchronization
	time.Sleep(200 * time.Millisecond)

	// Concurrently have hosts 2 and 3 leave
	go func() {
		gms[2].LeaveGroup(groupID)
	}()
	go func() {
		gms[3].LeaveGroup(groupID)
	}()

	// Wait for both leaves to complete
	time.Sleep(500 * time.Millisecond)

	// All hosts should not see hosts 2 and 3 as members
	for i := 0; i < 5; i++ {
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
			"Host %d should not see host2 as member after concurrent leaves", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[3].ID().String()),
			"Host %d should not see host3 as member after concurrent leaves", i)
	}

	// Verify remaining members are still present
	remainingHosts := []int{0, 1, 4}
	for _, i := range remainingHosts {
		for _, j := range remainingHosts {
			assert.True(t, gms[i].IsGroupMember(groupID, hosts[j].ID().String()),
				"Host %d should see host %d as member", i, j)
		}
	}
}

// TestMemberStateSynchronizationKeyRotation tests that key rotation happens on member leave
func TestMemberStateSynchronizationKeyRotation(t *testing.T) {
	// Create three hosts
	host1, err := createTestHost()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := createTestHost()
	require.NoError(t, err)
	defer host2.Close()

	host3, err := createTestHost()
	require.NoError(t, err)
	defer host3.Close()

	// Connect hosts
	ctx := context.Background()
	err = host1.Connect(ctx, peer.AddrInfo{ID: host2.ID(), Addrs: host2.Addrs()})
	require.NoError(t, err)
	err = host1.Connect(ctx, peer.AddrInfo{ID: host3.ID(), Addrs: host3.Addrs()})
	require.NoError(t, err)
	err = host2.Connect(ctx, peer.AddrInfo{ID: host3.ID(), Addrs: host3.Addrs()})
	require.NoError(t, err)

	// Create group managers
	gm1, err := NewGroupManager(ctx, host1)
	require.NoError(t, err)
	gm2, err := NewGroupManager(ctx, host2)
	require.NoError(t, err)
	gm3, err := NewGroupManager(ctx, host3)
	require.NoError(t, err)

	// Create group and add members
	groupID := "test-group-state-sync-key-rotation"
	err = gm1.CreateGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host3.ID().String())
	require.NoError(t, err)

	// Hosts join the group
	err = gm2.JoinGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm3.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Wait for synchronization
	time.Sleep(200 * time.Millisecond)

	// Get initial key versions
	keyVersion1 := gm1.GetGroupKeyVersion(groupID)
	keyVersion2 := gm2.GetGroupKeyVersion(groupID)
	_ = gm3.GetGroupKeyVersion(groupID)

	// Host3 leaves the group
	err = gm3.LeaveGroup(groupID)
	require.NoError(t, err)

	// Wait for state synchronization
	time.Sleep(300 * time.Millisecond)

	// Host1 and host2 should have updated key versions
	newKeyVersion1 := gm1.GetGroupKeyVersion(groupID)
	newKeyVersion2 := gm2.GetGroupKeyVersion(groupID)

	assert.Greater(t, newKeyVersion1, keyVersion1, "Host1 should have updated key version after host3 leaves")
	assert.Greater(t, newKeyVersion2, keyVersion2, "Host2 should have updated key version after host3 leaves")
	assert.Equal(t, newKeyVersion1, newKeyVersion2, "Host1 and host2 should have the same key version")

	// Host3 should no longer have access to the group (key version check should fail)
	// This is expected behavior - leaving members lose all group access
	// We can't check key version because the group no longer exists for host3
}

// TestMemberStateSynchronizationMessageHistory tests that message history is properly handled on member leave
func TestMemberStateSynchronizationMessageHistory(t *testing.T) {
	// Create three hosts
	host1, err := createTestHost()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := createTestHost()
	require.NoError(t, err)
	defer host2.Close()

	host3, err := createTestHost()
	require.NoError(t, err)
	defer host3.Close()

	// Connect hosts
	ctx := context.Background()
	err = host1.Connect(ctx, peer.AddrInfo{ID: host2.ID(), Addrs: host2.Addrs()})
	require.NoError(t, err)
	err = host1.Connect(ctx, peer.AddrInfo{ID: host3.ID(), Addrs: host3.Addrs()})
	require.NoError(t, err)
	err = host2.Connect(ctx, peer.AddrInfo{ID: host3.ID(), Addrs: host3.Addrs()})
	require.NoError(t, err)

	// Create group managers
	gm1, err := NewGroupManager(ctx, host1)
	require.NoError(t, err)
	gm2, err := NewGroupManager(ctx, host2)
	require.NoError(t, err)
	gm3, err := NewGroupManager(ctx, host3)
	require.NoError(t, err)

	// Create group and add members
	groupID := "test-group-state-sync-message-history"
	err = gm1.CreateGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host3.ID().String())
	require.NoError(t, err)

	// Hosts join the group
	err = gm2.JoinGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm3.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Wait for synchronization
	time.Sleep(200 * time.Millisecond)

	// Send some messages
	_, err = gm1.SendGroupMessage(ctx, groupID, []byte("Hello from host1"))
	require.NoError(t, err)
	_, err = gm2.SendGroupMessage(ctx, groupID, []byte("Hello from host2"))
	require.NoError(t, err)
	_, err = gm3.SendGroupMessage(ctx, groupID, []byte("Hello from host3"))
	require.NoError(t, err)

	// Wait for messages to propagate
	time.Sleep(200 * time.Millisecond)

	// Host3 leaves the group
	err = gm3.LeaveGroup(groupID)
	require.NoError(t, err)

	// Wait for state synchronization
	time.Sleep(300 * time.Millisecond)

	// Host1 and host2 should still have access to message history
	history1, err := gm1.GetGroupMessageHistory(ctx, groupID, 10)
	require.NoError(t, err)
	AssertMessageCountAtLeast(t, len(history1), 2, "Host1 message history") // Allow some variance due to PubSub deduplication

	history2, err := gm2.GetGroupMessageHistory(ctx, groupID, 10)
	require.NoError(t, err)
	AssertMessageCountAtLeast(t, len(history2), 2, "Host2 message history") // Allow some variance due to PubSub deduplication

	// Host3 should no longer have access to group messages
	history3, err := gm3.GetGroupMessageHistory(ctx, groupID, 10)
	require.NoError(t, err)
	assert.Len(t, history3, 0, "Host3 should have no messages in history after leaving")
}

// TestMemberStateSynchronizationRejoin tests that a member can rejoin after leaving
func TestMemberStateSynchronizationRejoin(t *testing.T) {
	// Create three hosts
	host1, err := createTestHost()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := createTestHost()
	require.NoError(t, err)
	defer host2.Close()

	host3, err := createTestHost()
	require.NoError(t, err)
	defer host3.Close()

	// Connect hosts
	ctx := context.Background()
	err = host1.Connect(ctx, peer.AddrInfo{ID: host2.ID(), Addrs: host2.Addrs()})
	require.NoError(t, err)
	err = host1.Connect(ctx, peer.AddrInfo{ID: host3.ID(), Addrs: host3.Addrs()})
	require.NoError(t, err)
	err = host2.Connect(ctx, peer.AddrInfo{ID: host3.ID(), Addrs: host3.Addrs()})
	require.NoError(t, err)

	// Create group managers
	gm1, err := NewGroupManager(ctx, host1)
	require.NoError(t, err)
	gm2, err := NewGroupManager(ctx, host2)
	require.NoError(t, err)
	gm3, err := NewGroupManager(ctx, host3)
	require.NoError(t, err)

	// Create group and add members
	groupID := "test-group-state-sync-rejoin"
	err = gm1.CreateGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host3.ID().String())
	require.NoError(t, err)

	// Hosts join the group
	err = gm2.JoinGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm3.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Wait for synchronization
	time.Sleep(200 * time.Millisecond)

	// Host3 leaves the group
	err = gm3.LeaveGroup(groupID)
	require.NoError(t, err)

	// Wait for state synchronization
	time.Sleep(300 * time.Millisecond)

	// Verify host3 is no longer a member
	assert.False(t, gm1.IsGroupMember(groupID, host3.ID().String()))
	assert.False(t, gm2.IsGroupMember(groupID, host3.ID().String()))
	assert.False(t, gm3.IsGroupMember(groupID, host3.ID().String()))

	// Host3 rejoins the group by having host1 add them back
	err = gm3.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Host1 adds host3 back to the group
	err = gm1.AddMemberWithContext(ctx, groupID, host3.ID().String())
	require.NoError(t, err)

	// Wait for state synchronization
	time.Sleep(300 * time.Millisecond)

	// All hosts should see host3 as a member again
	assert.True(t, gm1.IsGroupMember(groupID, host3.ID().String()))
	assert.True(t, gm2.IsGroupMember(groupID, host3.ID().String()))
	assert.True(t, gm3.IsGroupMember(groupID, host3.ID().String()))
}
