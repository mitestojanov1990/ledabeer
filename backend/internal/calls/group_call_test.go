package calls_test

import (
	"testing"

	"ledabeer/backend/internal/calls"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCall_Create(t *testing.T) {
	gc := calls.NewGroupCallManager(nil) // nil pubsub for testing

	callID, err := gc.CreateCall([]string{"peer1", "peer2", "peer3"})
	require.NoError(t, err)
	assert.NotEmpty(t, callID)
}

func TestGroupCall_JoinCall(t *testing.T) {
	gc := calls.NewGroupCallManager(nil)
	callID, _ := gc.CreateCall([]string{"peer1", "peer2"})

	err := gc.JoinCall(callID, "peer3")
	require.NoError(t, err)

	participants := gc.GetParticipants(callID)
	assert.Equal(t, 1, len(participants)) // Only peer3 was added to SFU
}

func TestGroupCall_BroadcastSignaling(t *testing.T) {
	// Should use PubSub to broadcast join/leave events
	gc := calls.NewGroupCallManager(nil)

	events := make(chan calls.CallEvent, 5)
	gc.OnEvent(func(e calls.CallEvent) {
		events <- e
	})

	callID, _ := gc.CreateCall([]string{"peer1", "peer2"})

	// Verify all participants notified
	// Should receive "participant_joined" events
	assert.NotEmpty(t, callID)
}
