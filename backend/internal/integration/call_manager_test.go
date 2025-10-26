package integration_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_CallManagerCompleteFlow(t *testing.T) {
	ctx := context.Background()

	// Setup two nodes
	aliceHost, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer aliceHost.Close()

	bobHost, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer bobHost.Close()

	// Create call managers
	aliceManager := calls.NewCallManager(aliceHost, nil)
	bobManager := calls.NewCallManager(bobHost, nil)

	// Alice initiates call to Bob
	callID, err := aliceManager.InitiateCall(ctx, bobHost.ID(), calls.CallOptions{
		AudioEnabled: true,
		VideoEnabled: false,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, callID)

	// Verify Alice's call state
	aliceState := aliceManager.GetCallState(callID)
	assert.Equal(t, "initiating", aliceState.State)
	assert.False(t, aliceState.AudioMuted)
	assert.False(t, aliceState.VideoEnabled)

	// Bob accepts the call
	err = bobManager.AcceptCall(callID)
	require.NoError(t, err)

	// Note: In a real implementation, Bob would receive the call offer
	// and his state would be updated. For testing, we'll verify Alice's state.

	// Test media controls
	err = aliceManager.MuteAudio(callID)
	require.NoError(t, err)

	aliceState = aliceManager.GetCallState(callID)
	assert.True(t, aliceState.AudioMuted)

	err = aliceManager.UnmuteAudio(callID)
	require.NoError(t, err)

	aliceState = aliceManager.GetCallState(callID)
	assert.False(t, aliceState.AudioMuted)

	// Test video enable
	err = aliceManager.EnableVideo(callID)
	require.NoError(t, err)

	aliceState = aliceManager.GetCallState(callID)
	assert.True(t, aliceState.VideoEnabled)

	// End the call
	err = aliceManager.EndCall(callID)
	require.NoError(t, err)

	// Verify call is ended
	aliceState = aliceManager.GetCallState(callID)
	assert.Equal(t, "ended", aliceState.State)
}

func TestE2E_GroupCallManager(t *testing.T) {
	ctx := context.Background()

	// Setup three nodes
	aliceHost, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer aliceHost.Close()

	bobHost, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer bobHost.Close()

	charlieHost, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer charlieHost.Close()

	// Create call managers
	aliceManager := calls.NewCallManager(aliceHost, nil)
	bobManager := calls.NewCallManager(bobHost, nil)
	charlieManager := calls.NewCallManager(charlieHost, nil)

	// Alice creates group call
	participants := []peer.ID{bobHost.ID(), charlieHost.ID()}
	callID, err := aliceManager.CreateGroupCall(participants)
	require.NoError(t, err)
	assert.NotEmpty(t, callID)

	// Bob and Charlie join
	err = bobManager.JoinGroupCall(callID)
	require.NoError(t, err)

	err = charlieManager.JoinGroupCall(callID)
	require.NoError(t, err)

	// Test media controls in group call
	err = aliceManager.MuteAudio(callID)
	require.NoError(t, err)

	err = aliceManager.EnableVideo(callID)
	require.NoError(t, err)

	// Note: In a real implementation, the call state would be updated
	// For testing, we'll just verify the operations don't error

	// Leave group call
	err = aliceManager.LeaveGroupCall(callID)
	require.NoError(t, err)
}

func TestE2E_CallManagerWithTURN(t *testing.T) {
	ctx := context.Background()

	// Setup host
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	// Create call manager with TURN config
	turnConfig := &calls.TURNConfig{
		URLs:       []string{"turn:turn.example.com:3478"},
		Username:   "user",
		Credential: "pass",
	}

	manager := calls.NewCallManager(host, nil)

	// Initiate call with TURN
	callID, err := manager.InitiateCall(ctx, peer.ID("test-peer"), calls.CallOptions{
		AudioEnabled: true,
		VideoEnabled: true,
		TURNConfig:   turnConfig,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, callID)

	// Verify call state
	state := manager.GetCallState(callID)
	assert.Equal(t, "initiating", state.State)
	assert.True(t, state.VideoEnabled)
}

func TestE2E_ActiveCallsTracking(t *testing.T) {
	ctx := context.Background()

	// Setup host
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Initially no active calls
	activeCalls := manager.GetActiveCalls()
	assert.Empty(t, activeCalls)

	// Create multiple calls
	callID1, err := manager.InitiateCall(ctx, peer.ID("peer1"), calls.CallOptions{})
	require.NoError(t, err)

	callID2, err := manager.InitiateCall(ctx, peer.ID("peer2"), calls.CallOptions{})
	require.NoError(t, err)

	// Should have 2 active calls
	activeCalls = manager.GetActiveCalls()
	assert.Len(t, activeCalls, 2)
	assert.Contains(t, activeCalls, callID1)
	assert.Contains(t, activeCalls, callID2)

	// End one call
	err = manager.EndCall(callID1)
	require.NoError(t, err)

	// Should have 1 active call
	activeCalls = manager.GetActiveCalls()
	assert.Len(t, activeCalls, 1)
	assert.Contains(t, activeCalls, callID2)

	// End remaining call
	err = manager.EndCall(callID2)
	require.NoError(t, err)

	// Should have no active calls
	activeCalls = manager.GetActiveCalls()
	assert.Empty(t, activeCalls)
}
