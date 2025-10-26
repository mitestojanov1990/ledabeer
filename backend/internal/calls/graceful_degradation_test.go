package calls

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulDegradation_RealGroupCalls_PartialFailure(t *testing.T) {
	// Should handle partial failures in group calls gracefully
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	// Create call manager with graceful degradation
	callManager := NewCallManagerWithDegradation(host1, &DegradationConfig{
		MaxParticipants:  5,
		MinParticipants:  2,
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 100,
		GracefulDegrade:  true,
		FallbackMode:     true,
	})

	// Create group call
	callID, err := callManager.InitiateGroupCall(ctx, "group1", []string{"peer1", "peer2", "peer3", "peer4"})
	require.NoError(t, err)

	// Simulate partial participant failures
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer2")
	require.NoError(t, err)
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer3")
	require.NoError(t, err)

	// Wait for graceful degradation
	time.Sleep(50 * time.Millisecond)

	// Verify call continues with remaining participants
	session := callManager.GetCallSession(callID)
	require.NotNil(t, session)

	participants := session.GetParticipants()
	assert.Contains(t, participants, "peer1")
	assert.Contains(t, participants, "peer4")
	assert.NotContains(t, participants, "peer2")
	assert.NotContains(t, participants, "peer3")

	// Verify degradation stats
	stats := callManager.GetDegradationStats()
	assert.Equal(t, 2, stats.FailedParticipants)
	assert.Equal(t, 2, stats.DegradationEvents)  // One event per failure
	assert.Equal(t, 3, stats.ActiveParticipants) // 2 remaining + 1 local
}

func TestGracefulDegradation_RealGroupCalls_MinimumParticipants(t *testing.T) {
	// Should maintain minimum participants for call quality
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	callManager := NewCallManagerWithDegradation(host1, &DegradationConfig{
		MaxParticipants:  5,
		MinParticipants:  2,
		FailureThreshold: 1,
		RecoveryTimeout:  time.Millisecond * 100,
		GracefulDegrade:  true,
		FallbackMode:     true,
	})

	// Create group call
	callID, err := callManager.InitiateGroupCall(ctx, "group1", []string{"peer1", "peer2", "peer3"})
	require.NoError(t, err)

	// Simulate failures that would go below minimum
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer1")
	require.NoError(t, err)
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer2")
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Verify call is still active (above minimum)
	session := callManager.GetCallSession(callID)
	require.NotNil(t, session)

	participants := session.GetParticipants()
	assert.Len(t, participants, 2) // peer3 + localParticipant
	assert.Contains(t, participants, "peer3")

	// Verify call state is maintained
	assert.Equal(t, StateConnected, session.GetState())

	// Verify degradation stats
	stats := callManager.GetDegradationStats()
	assert.Equal(t, 2, stats.FailedParticipants)
	assert.Equal(t, 2, stats.ActiveParticipants) // 1 remaining + 1 local
}

func TestGracefulDegradation_RealGroupCalls_Recovery(t *testing.T) {
	// Should recover from partial failures when participants come back
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	callManager := NewCallManagerWithDegradation(host1, &DegradationConfig{
		MaxParticipants:  5,
		MinParticipants:  1,
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 50,
		GracefulDegrade:  true,
		FallbackMode:     true,
	})

	// Create group call
	callID, err := callManager.InitiateGroupCall(ctx, "group1", []string{"peer1", "peer2", "peer3"})
	require.NoError(t, err)

	// Simulate participant failure
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer2")
	require.NoError(t, err)

	// Wait for degradation
	time.Sleep(30 * time.Millisecond)

	// Simulate participant recovery
	err = callManager.SimulateParticipantRecovery(ctx, callID, "peer2")
	require.NoError(t, err)

	// Wait for recovery
	time.Sleep(100 * time.Millisecond)

	// Verify participant is back
	session := callManager.GetCallSession(callID)
	require.NotNil(t, session)

	participants := session.GetParticipants()
	assert.Contains(t, participants, "peer1")
	assert.Contains(t, participants, "peer2")
	assert.Contains(t, participants, "peer3")

	// Verify recovery stats
	stats := callManager.GetDegradationStats()
	assert.Equal(t, 1, stats.RecoveryEvents)
	assert.Equal(t, 4, stats.ActiveParticipants) // 3 participants + 1 local
}

func TestGracefulDegradation_RealGroupCalls_FallbackMode(t *testing.T) {
	// Should switch to fallback mode when too many failures
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	callManager := NewCallManagerWithDegradation(host1, &DegradationConfig{
		MaxParticipants:  5,
		MinParticipants:  1,
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 100,
		GracefulDegrade:  true,
		FallbackMode:     true,
	})

	// Create group call
	callID, err := callManager.InitiateGroupCall(ctx, "group1", []string{"peer1", "peer2", "peer3", "peer4"})
	require.NoError(t, err)

	// Simulate multiple failures to trigger fallback
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer2")
	require.NoError(t, err)
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer3")
	require.NoError(t, err)
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer4")
	require.NoError(t, err)

	// Wait for fallback mode
	time.Sleep(100 * time.Millisecond)

	// Verify fallback mode is active
	session := callManager.GetCallSession(callID)
	require.NotNil(t, session)

	// Verify fallback mode stats
	stats := callManager.GetDegradationStats()
	assert.True(t, stats.FallbackModeActive)
	assert.Equal(t, 3, stats.FailedParticipants)
	assert.Equal(t, 2, stats.ActiveParticipants) // 1 remaining + 1 local
}

func TestGracefulDegradation_RealGroupCalls_QualityAdaptation(t *testing.T) {
	// Should adapt call quality based on participant count
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	callManager := NewCallManagerWithDegradation(host1, &DegradationConfig{
		MaxParticipants:   5,
		MinParticipants:   1,
		FailureThreshold:  2,
		RecoveryTimeout:   time.Millisecond * 100,
		GracefulDegrade:   true,
		FallbackMode:      true,
		QualityAdaptation: true,
	})

	// Create group call
	callID, err := callManager.InitiateGroupCall(ctx, "group1", []string{"peer1", "peer2", "peer3"})
	require.NoError(t, err)

	// Get initial quality
	session := callManager.GetCallSession(callID)
	require.NotNil(t, session)

	initialQuality := session.GetCallQuality()
	assert.Equal(t, "high", initialQuality)

	// Simulate participant failure
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer2")
	require.NoError(t, err)

	// Wait for quality adaptation
	time.Sleep(50 * time.Millisecond)

	// Verify quality has adapted
	adaptedQuality := session.GetCallQuality()
	assert.Equal(t, "medium", adaptedQuality)

	// Simulate another failure
	err = callManager.SimulateParticipantFailure(ctx, callID, "peer3")
	require.NoError(t, err)

	// Wait for further adaptation
	time.Sleep(50 * time.Millisecond)

	// Verify quality has further degraded
	finalQuality := session.GetCallQuality()
	assert.Equal(t, "low", finalQuality)
}

func TestGracefulDegradation_RealGroupCalls_ConcurrentFailures(t *testing.T) {
	// Should handle concurrent participant failures gracefully
	ctx := context.Background()

	host1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host1.Close()

	callManager := NewCallManagerWithDegradation(host1, &DegradationConfig{
		MaxParticipants:  5,
		MinParticipants:  1,
		FailureThreshold: 2,
		RecoveryTimeout:  time.Millisecond * 100,
		GracefulDegrade:  true,
		FallbackMode:     true,
	})

	// Create group call
	callID, err := callManager.InitiateGroupCall(ctx, "group1", []string{"peer1", "peer2", "peer3", "peer4"})
	require.NoError(t, err)

	// Concurrent participant failures
	numFailures := 3
	var wg sync.WaitGroup
	for i := 0; i < numFailures; i++ {
		wg.Add(1)
		go func(peerID string) {
			defer wg.Done()
			err := callManager.SimulateParticipantFailure(ctx, callID, peerID)
			assert.NoError(t, err)
		}(fmt.Sprintf("peer%d", i+2)) // peer2, peer3, peer4
	}
	wg.Wait()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify call is still active
	session := callManager.GetCallSession(callID)
	require.NotNil(t, session)

	participants := session.GetParticipants()
	assert.Len(t, participants, 2) // peer1 + localParticipant
	assert.Contains(t, participants, "peer1")

	// Verify degradation stats
	stats := callManager.GetDegradationStats()
	assert.Equal(t, 3, stats.FailedParticipants)
	assert.Equal(t, 2, stats.ActiveParticipants) // peer1 + localParticipant
	assert.True(t, stats.FallbackModeActive)
}
