package messaging

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnhancedMemberRemovalBroadcasting tests enhanced member removal with acknowledgments
func TestEnhancedMemberRemovalBroadcasting(t *testing.T) {
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

	// Connect all hosts
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

	// Create group and add all members
	groupID := "test-group-enhanced-removal"
	err = gm1.CreateGroup(ctx, groupID)
	require.NoError(t, err)

	err = gm1.AddMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host3.ID().String())
	require.NoError(t, err)

	// All hosts join the group
	err = gm2.JoinGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm3.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Wait for initial synchronization
	time.Sleep(200 * time.Millisecond)

	// Verify all members are present
	assert.True(t, gm1.IsGroupMember(groupID, host1.ID().String()))
	assert.True(t, gm1.IsGroupMember(groupID, host2.ID().String()))
	assert.True(t, gm1.IsGroupMember(groupID, host3.ID().String()))

	// Remove host3 from the group via host1 (with acknowledgment request)
	err = gm1.RemoveMemberWithContext(ctx, groupID, host3.ID().String())
	require.NoError(t, err)

	// Wait for message propagation and acknowledgments
	time.Sleep(500 * time.Millisecond)

	// Host1 should no longer see host3 as a member
	assert.False(t, gm1.IsGroupMember(groupID, host3.ID().String()), "Host1 should not see host3 as member after removal")

	// Host2 should process the removal message and no longer see host3
	removalMsg := map[string]interface{}{
		"type":        "member_removed",
		"group_id":    groupID,
		"member_id":   host3.ID().String(),
		"timestamp":   time.Now().Unix(),
		"request_ack": true,
		"from":        host1.ID().String(),
	}
	gm2.handleMemberRemovedMessage(groupID, removalMsg)
	assert.False(t, gm2.IsGroupMember(groupID, host3.ID().String()), "Host2 should not see host3 as member after processing removal")

	// Host3 should process the removal message and leave the group
	gm3.handleMemberRemovedMessage(groupID, removalMsg)
	assert.False(t, gm3.IsGroupMember(groupID, host3.ID().String()), "Host3 should not see itself as member after processing removal")
}

// TestRemovalAcknowledgmentSystem tests the acknowledgment system for removals
func TestRemovalAcknowledgmentSystem(t *testing.T) {
	// Create two hosts
	host1, err := createTestHost()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := createTestHost()
	require.NoError(t, err)
	defer host2.Close()

	// Connect hosts
	ctx := context.Background()
	err = host1.Connect(ctx, peer.AddrInfo{ID: host2.ID(), Addrs: host2.Addrs()})
	require.NoError(t, err)

	// Create group managers
	gm1, err := NewGroupManager(ctx, host1)
	require.NoError(t, err)
	gm2, err := NewGroupManager(ctx, host2)
	require.NoError(t, err)

	// Create group and add member
	groupID := "test-group-removal-ack"
	err = gm1.CreateGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)
	err = gm2.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Wait for synchronization
	time.Sleep(100 * time.Millisecond)

	// Remove host2 from the group
	err = gm1.RemoveMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)

	// Wait for removal message
	time.Sleep(100 * time.Millisecond)

	// Simulate host2 processing the removal and sending acknowledgment
	removalMsg := map[string]interface{}{
		"type":        "member_removed",
		"group_id":    groupID,
		"member_id":   host2.ID().String(),
		"timestamp":   time.Now().Unix(),
		"request_ack": true,
		"from":        host1.ID().String(),
	}
	gm2.handleMemberRemovedMessage(groupID, removalMsg)

	// Simulate acknowledgment message
	ackMsg := map[string]interface{}{
		"type":      "removal_ack",
		"group_id":  groupID,
		"member_id": host2.ID().String(),
		"from":      host2.ID().String(),
		"timestamp": time.Now().Unix(),
	}
	gm1.handleRemovalAckMessage(groupID, ackMsg)

	// Check that acknowledgment was recorded
	gm1.removalMu.RLock()
	hasAck := gm1.removalAcks[groupID][host2.ID().String()][host2.ID().String()]
	gm1.removalMu.RUnlock()

	assert.True(t, hasAck, "Acknowledgment should be recorded")
}

// TestRemovalRetryMechanism tests the retry mechanism for failed removals
func TestRemovalRetryMechanism(t *testing.T) {
	// Create two hosts
	host1, err := createTestHost()
	require.NoError(t, err)
	defer host1.Close()

	host2, err := createTestHost()
	require.NoError(t, err)
	defer host2.Close()

	// Connect hosts
	ctx := context.Background()
	err = host1.Connect(ctx, peer.AddrInfo{ID: host2.ID(), Addrs: host2.Addrs()})
	require.NoError(t, err)

	// Create group managers
	gm1, err := NewGroupManager(ctx, host1)
	require.NoError(t, err)
	gm2, err := NewGroupManager(ctx, host2)
	require.NoError(t, err)

	// Create group and add member
	groupID := "test-group-removal-retry"
	err = gm1.CreateGroup(ctx, groupID)
	require.NoError(t, err)
	err = gm1.AddMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)
	err = gm2.JoinGroup(ctx, groupID)
	require.NoError(t, err)

	// Wait for synchronization
	time.Sleep(100 * time.Millisecond)

	// Remove host2 from the group
	err = gm1.RemoveMemberWithContext(ctx, groupID, host2.ID().String())
	require.NoError(t, err)

	// Wait for removal message
	time.Sleep(100 * time.Millisecond)

	// Simulate retry message
	retryMsg := map[string]interface{}{
		"type":      "removal_retry",
		"group_id":  groupID,
		"member_id": host2.ID().String(),
		"from":      host1.ID().String(),
		"timestamp": time.Now().Unix(),
	}
	gm2.handleRemovalRetryMessage(groupID, retryMsg)

	// Host2 should no longer see itself as a member
	assert.False(t, gm2.IsGroupMember(groupID, host2.ID().String()), "Host2 should not see itself as member after retry")
}

// TestRemovalMessageOrdering tests that removal messages are processed in correct order
func TestRemovalMessageOrdering(t *testing.T) {
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
	groupID := "test-group-removal-ordering"
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

	// Remove hosts 2 and 3 in sequence
	err = gms[0].RemoveMemberWithContext(ctx, groupID, hosts[2].ID().String())
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	err = gms[0].RemoveMemberWithContext(ctx, groupID, hosts[3].ID().String())
	require.NoError(t, err)

	// Wait for message propagation
	time.Sleep(200 * time.Millisecond)

	// Process removal messages in order
	removalMsg2 := map[string]interface{}{
		"type":        "member_removed",
		"group_id":    groupID,
		"member_id":   hosts[2].ID().String(),
		"timestamp":   time.Now().Unix() - 1,
		"request_ack": true,
		"from":        hosts[0].ID().String(),
	}
	removalMsg3 := map[string]interface{}{
		"type":        "member_removed",
		"group_id":    groupID,
		"member_id":   hosts[3].ID().String(),
		"timestamp":   time.Now().Unix(),
		"request_ack": true,
		"from":        hosts[0].ID().String(),
	}

	// Process messages on all hosts
	for _, gm := range gms {
		gm.handleMemberRemovedMessage(groupID, removalMsg2)
		gm.handleMemberRemovedMessage(groupID, removalMsg3)
	}

	// Only hosts 0 and 1 should remain as members
	// Hosts 0 and 1 should see each other and not see removed hosts
	for i := 0; i < 2; i++ {
		assert.True(t, gms[i].IsGroupMember(groupID, hosts[0].ID().String()),
			"Host %d should see host0 as member", i)
		assert.True(t, gms[i].IsGroupMember(groupID, hosts[1].ID().String()),
			"Host %d should see host1 as member", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
			"Host %d should not see host2 as member", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[3].ID().String()),
			"Host %d should not see host3 as member", i)
	}

	// Hosts 2 and 3 should have lost access to the group entirely
	for i := 2; i < 4; i++ {
		// They should not see any members because they're no longer in the group
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[0].ID().String()),
			"Host %d should not see host0 as member (no group access)", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[1].ID().String()),
			"Host %d should not see host1 as member (no group access)", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
			"Host %d should not see host2 as member", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[3].ID().String()),
			"Host %d should not see host3 as member", i)
	}
}

// TestRemovalConcurrentWithAcks tests concurrent removals with acknowledgment system
func TestRemovalConcurrentWithAcks(t *testing.T) {
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
	groupID := "test-group-removal-concurrent-acks"
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

	// Concurrently remove hosts 2 and 3
	go func() {
		gms[0].RemoveMemberWithContext(ctx, groupID, hosts[2].ID().String())
	}()
	go func() {
		gms[1].RemoveMemberWithContext(ctx, groupID, hosts[3].ID().String())
	}()

	// Wait for both removals to complete
	time.Sleep(500 * time.Millisecond)

	// Process removal messages for all hosts
	for i := 0; i < 4; i++ {
		// Process host2 removal
		removalMsg2 := map[string]interface{}{
			"type":        "member_removed",
			"group_id":    groupID,
			"member_id":   hosts[2].ID().String(),
			"timestamp":   time.Now().Unix(),
			"request_ack": true,
			"from":        hosts[0].ID().String(),
		}
		gms[i].handleMemberRemovedMessage(groupID, removalMsg2)

		// Process host3 removal
		removalMsg3 := map[string]interface{}{
			"type":        "member_removed",
			"group_id":    groupID,
			"member_id":   hosts[3].ID().String(),
			"timestamp":   time.Now().Unix(),
			"request_ack": true,
			"from":        hosts[1].ID().String(),
		}
		gms[i].handleMemberRemovedMessage(groupID, removalMsg3)
	}

	// Only hosts 0 and 1 should remain as members
	// Hosts 0 and 1 should see each other and not see removed hosts
	for i := 0; i < 2; i++ {
		assert.True(t, gms[i].IsGroupMember(groupID, hosts[0].ID().String()),
			"Host %d should see host0 as member", i)
		assert.True(t, gms[i].IsGroupMember(groupID, hosts[1].ID().String()),
			"Host %d should see host1 as member", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
			"Host %d should not see host2 as member", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[3].ID().String()),
			"Host %d should not see host3 as member", i)
	}

	// Hosts 2 and 3 should have lost access to the group entirely
	for i := 2; i < 4; i++ {
		// They should not see any members because they're no longer in the group
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[0].ID().String()),
			"Host %d should not see host0 as member (no group access)", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[1].ID().String()),
			"Host %d should not see host1 as member (no group access)", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[2].ID().String()),
			"Host %d should not see host2 as member", i)
		assert.False(t, gms[i].IsGroupMember(groupID, hosts[3].ID().String()),
			"Host %d should not see host3 as member", i)
	}
}

// Helper function to create a test host
func createTestHost() (host.Host, error) {
	// Create a basic libp2p host for testing
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.Identity(createTestIdentity()),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// Helper function to create a test identity
func createTestIdentity() crypto.PrivKey {
	// Generate a random Ed25519 private key for testing
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}
	return priv
}
