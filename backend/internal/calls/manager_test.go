package calls_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/network"

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

	manager := calls.NewCallManager(host)

	// Create a dummy peer ID for testing
	peerID := "test-peer"

	callID, err := manager.InitiateCall(ctx, peerID, true, false)
	require.NoError(t, err)
	assert.NotEmpty(t, callID)
}

func TestCallManager_AnswerCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host)

	// First initiate a call
	callID, err := manager.InitiateCall(ctx, "test-peer", true, false)
	require.NoError(t, err)

	// Then answer it
	err = manager.AnswerCall(ctx, callID, true)
	require.NoError(t, err)
}

func TestCallManager_EndCall(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host)

	// First initiate a call
	callID, err := manager.InitiateCall(ctx, "test-peer", true, false)
	require.NoError(t, err)

	// Then end it
	err = manager.EndCall(ctx, callID)
	require.NoError(t, err)
}

func TestCallManager_GetCallSession(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host)

	// Initiate a call
	callID, err := manager.InitiateCall(ctx, "test-peer", true, false)
	require.NoError(t, err)

	// Get the call session
	session := manager.GetCallSession(callID)
	assert.NotNil(t, session)
}

func TestCallManager_HandleSignaling(t *testing.T) {
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host)

	// Test signaling message handling
	response, err := manager.HandleSignaling(ctx, "test-call-id", "offer", "test-sdp", "")
	require.NoError(t, err)
	assert.NotNil(t, response)
}
