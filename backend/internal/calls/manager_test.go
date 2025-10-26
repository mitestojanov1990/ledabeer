package calls_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/network"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallManager_InitiateCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Create a dummy peer ID for testing
	peerID := peer.ID("test-peer")

	callID, err := manager.InitiateCall(ctx, peerID, calls.CallOptions{})
	require.NoError(t, err)
	assert.NotEmpty(t, callID)
}

func TestCallManager_AcceptCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Accept a call
	err = manager.AcceptCall("test-call-id")
	require.NoError(t, err)
}

func TestCallManager_EndCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// End a call
	err = manager.EndCall("test-call-id")
	require.NoError(t, err)
}

func TestCallManager_CreateGroupCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Create group call with participants
	participants := []peer.ID{
		peer.ID("peer1"),
		peer.ID("peer2"),
		peer.ID("peer3"),
	}

	callID, err := manager.CreateGroupCall(participants)
	require.NoError(t, err)
	assert.NotEmpty(t, callID)
}

func TestCallManager_JoinGroupCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Join group call
	err = manager.JoinGroupCall("test-group-call-id")
	require.NoError(t, err)
}

func TestCallManager_MediaControls(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Test mute/unmute
	err = manager.MuteAudio("test-call-id")
	require.NoError(t, err)

	err = manager.UnmuteAudio("test-call-id")
	require.NoError(t, err)

	// Test video enable/disable
	err = manager.EnableVideo("test-call-id")
	require.NoError(t, err)

	err = manager.DisableVideo("test-call-id")
	require.NoError(t, err)
}

func TestCallManager_GetCallState(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Get call state
	state := manager.GetCallState("test-call-id")
	assert.NotEmpty(t, state)
}

func TestCallManager_GetActiveCalls(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host, nil)

	// Get active calls
	calls := manager.GetActiveCalls()
	assert.NotNil(t, calls)
}
