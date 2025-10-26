package calls_test

import (
	"testing"

	"ledabeer/backend/internal/calls"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignaling_SDPOffer(t *testing.T) {
	// Should create and encrypt SDP offer
	caller := calls.NewCallSession()
	offer, err := caller.CreateOffer()
	require.NoError(t, err)
	assert.NotEmpty(t, offer.SDP)
	assert.Equal(t, "offer", offer.Type)
}

func TestSignaling_SDPAnswer(t *testing.T) {
	// Should process offer and generate answer
	callee := calls.NewCallSession()
	offer := &calls.SDP{Type: "offer", SDP: "v=0..."}
	answer, err := callee.CreateAnswer(offer)
	require.NoError(t, err)
	assert.Equal(t, "answer", answer.Type)
}

func TestSignaling_ICECandidates(t *testing.T) {
	// Should gather and exchange ICE candidates
	session := calls.NewCallSession()
	candidates, err := session.GatherCandidates()
	require.NoError(t, err)
	assert.Greater(t, len(candidates), 0)
}

func TestSignaling_EncryptedExchange(t *testing.T) {
	// SDP/ICE must be encrypted with E2EE
	session := calls.NewCallSession()
	offer, _ := session.CreateOffer()
	encrypted, err := session.EncryptSignaling(offer)
	require.NoError(t, err)
	assert.NotEqual(t, offer.SDP, encrypted)
}
