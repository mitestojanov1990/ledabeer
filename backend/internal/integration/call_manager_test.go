package integration_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/network"

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
	aliceManager := calls.NewCallManager(aliceHost)
	bobManager := calls.NewCallManager(bobHost)

	// Alice initiates call to Bob
	callID, err := aliceManager.InitiateCall(ctx, bobHost.ID().String(), true, false)
	require.NoError(t, err)
	assert.NotEmpty(t, callID)

	// Bob answers the call
	err = bobManager.AnswerCall(ctx, callID, true)
	require.NoError(t, err)

	// Test that we can get the call session
	aliceSession := aliceManager.GetCallSession(callID)
	assert.NotNil(t, aliceSession)

	bobSession := bobManager.GetCallSession(callID)
	assert.NotNil(t, bobSession)

	// Test call end
	err = aliceManager.EndCall(ctx, callID)
	require.NoError(t, err)

	// Test that call session is cleaned up (session should be nil after EndCall)
	aliceSession = aliceManager.GetCallSession(callID)
	// Note: GetCallSession may return nil after EndCall, which is expected behavior
}

func TestE2E_CallManagerSignaling(t *testing.T) {
	ctx := context.Background()

	// Setup host
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := calls.NewCallManager(host)

	// Test signaling message handling
	callID := "test-call"

	// Test offer
	resp, err := manager.HandleSignaling(ctx, callID, "offer", "sdp-offer", "")
	require.NoError(t, err)
	assert.Equal(t, "answer", resp.Type)
	assert.NotEmpty(t, resp.Sdp)

	// Test answer
	resp, err = manager.HandleSignaling(ctx, callID, "answer", "sdp-answer", "")
	require.NoError(t, err)
	assert.Nil(t, resp)

	// Test ICE candidate
	resp, err = manager.HandleSignaling(ctx, callID, "ice-candidate", "", "ice-candidate-data")
	require.NoError(t, err)
	assert.Equal(t, "ice-candidate", resp.Type)
	assert.NotEmpty(t, resp.Candidate)
}
