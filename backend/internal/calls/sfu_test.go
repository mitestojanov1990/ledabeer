package calls_test

import (
	"testing"

	"ledabeer/backend/internal/calls"

	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func TestSFU_MultipleParticipants(t *testing.T) {
	sfu := calls.NewSFU()

	// Add 3 participants
	p1 := sfu.AddParticipant("peer1")
	p2 := sfu.AddParticipant("peer2")
	p3 := sfu.AddParticipant("peer3")

	assert.Equal(t, 3, sfu.ParticipantCount())
	assert.NotNil(t, p1)
	assert.NotNil(t, p2)
	assert.NotNil(t, p3)
}

func TestSFU_ForwardTrack(t *testing.T) {
	sfu := calls.NewSFU()

	p2 := sfu.AddParticipant("peer2")

	// P1 sends track
	track := createTestTrack()
	sfu.AddTrack("peer1", track)

	// P2 should receive it
	tracks := p2.GetRemoteTracks()
	assert.Equal(t, 1, len(tracks))
}

func TestSFU_ParticipantLeaves(t *testing.T) {
	sfu := calls.NewSFU()

	sfu.AddParticipant("peer1")
	sfu.AddParticipant("peer2")

	sfu.RemoveParticipant("peer1")

	// P2 should be notified
	// Track should be removed
	assert.Equal(t, 1, sfu.ParticipantCount())
}

func createTestTrack() *webrtc.TrackRemote {
	// Create a dummy track for testing
	return &webrtc.TrackRemote{}
}
