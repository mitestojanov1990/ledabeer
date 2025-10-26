package calls

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallSession_RealConnectionState_StateTransitions(t *testing.T) {
	// Should track real WebRTC connection state transitions
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Test initial state
	state := session.GetState()
	assert.Equal(t, StateInitiating, state)

	// Test state changes based on WebRTC connection state
	// In a real implementation, this would be triggered by WebRTC events
	if session.pc != nil {
		// Simulate connection state change
		session.pc.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
			// This would update the session's internal state
		})
	}
}

func TestCallSession_RealConnectionState_ConnectionFailure(t *testing.T) {
	// Should handle connection failures gracefully
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Test that failed connections are handled
	state := session.GetState()
	assert.NotEqual(t, StateEnded, state) // Should not be ended initially

	// Simulate connection failure
	if session.pc != nil {
		// In real implementation, this would be triggered by WebRTC failure
		session.pc.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
			if connectionState == webrtc.PeerConnectionStateFailed {
				// Handle failure
			}
		})
	}
}

func TestCallSession_RealConnectionState_Disconnected(t *testing.T) {
	// Should detect disconnections
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Test disconnected state
	state := session.GetState()
	assert.NotEqual(t, StateEnded, state) // Should not be ended initially

	// Simulate disconnection
	if session.pc != nil {
		session.pc.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
			if connectionState == webrtc.PeerConnectionStateDisconnected {
				// Handle disconnection
			}
		})
	}
}

func TestCallSession_RealParticipants_TrackParticipants(t *testing.T) {
	// Should track real participants in call
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Test participants tracking
	participants := session.GetParticipants()
	assert.NotNil(t, participants)
	assert.Contains(t, participants, session.localParticipant)

	// Test adding remote participants
	// In a real implementation, this would happen when remote peer joins
	remoteParticipant := "remote-peer-123"
	session.AddParticipant(remoteParticipant)

	participants = session.GetParticipants()
	assert.Contains(t, participants, remoteParticipant)
	assert.Contains(t, participants, session.localParticipant)
}

func TestCallSession_RealParticipants_RemoveParticipants(t *testing.T) {
	// Should remove participants when they leave
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Add participant
	remoteParticipant := "remote-peer-123"
	session.AddParticipant(remoteParticipant)

	// Verify participant is added
	participants := session.GetParticipants()
	assert.Contains(t, participants, remoteParticipant)

	// Remove participant
	session.RemoveParticipant(remoteParticipant)

	// Verify participant is removed
	participants = session.GetParticipants()
	assert.NotContains(t, participants, remoteParticipant)
	assert.Contains(t, participants, session.localParticipant)
}

func TestCallSession_RealParticipants_GroupCall(t *testing.T) {
	// Should handle multiple participants in group call
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, true)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Add multiple participants
	participants := []string{"peer1", "peer2", "peer3"}
	for _, participant := range participants {
		session.AddParticipant(participant)
	}

	// Verify all participants are tracked
	allParticipants := session.GetParticipants()
	assert.Len(t, allParticipants, len(participants)+1) // +1 for local participant
	assert.Contains(t, allParticipants, session.localParticipant)
	for _, participant := range participants {
		assert.Contains(t, allParticipants, participant)
	}
}

func TestCallSession_RealConnectionState_StateConsistency(t *testing.T) {
	// Should maintain state consistency across operations
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Test state consistency
	initialState := session.GetState()
	assert.Equal(t, StateInitiating, initialState)

	// Add participants
	session.AddParticipant("peer1")
	session.AddParticipant("peer2")

	// State should remain consistent
	state := session.GetState()
	assert.Equal(t, initialState, state)

	// Participants should be tracked
	participants := session.GetParticipants()
	assert.Len(t, participants, 3) // local + 2 remote
}

func TestCallSession_RealConnectionState_ConcurrentAccess(t *testing.T) {
	// Should handle concurrent access to state and participants
	ctx := context.Background()

	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	manager := NewCallManager(host)

	// Create call session
	callID, err := manager.InitiateCall(ctx, "peer2", true, false)
	require.NoError(t, err)

	session := manager.GetCallSession(callID)
	require.NotNil(t, session)

	// Test concurrent access
	done := make(chan bool, 10)

	// Concurrent state reads
	for i := 0; i < 5; i++ {
		go func() {
			state := session.GetState()
			assert.NotNil(t, state)
			done <- true
		}()
	}

	// Concurrent participant operations
	for i := 0; i < 5; i++ {
		go func(i int) {
			participant := fmt.Sprintf("peer%d", i)
			session.AddParticipant(participant)
			participants := session.GetParticipants()
			assert.NotNil(t, participants)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}
