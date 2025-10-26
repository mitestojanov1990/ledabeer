package crypto_test

import (
	"testing"

	"ledabeer/backend/internal/crypto"

	"github.com/stretchr/testify/assert"
)

func TestX3DH_InitiatorReceiverKeyAgreement(t *testing.T) {
	// Alice (initiator) and Bob (receiver) should derive same shared secret
	alice := crypto.NewX3DHSession()
	bob := crypto.NewX3DHSession()

	// Bob publishes prekey bundle
	bobBundle := bob.GeneratePrekeyBundle()

	// Alice initiates key exchange
	aliceSharedSecret, initialMsg, err := alice.InitiateKeyExchange(bobBundle)
	assert.NoError(t, err)

	// Bob processes initial message
	bobSharedSecret, err := bob.ProcessKeyExchange(initialMsg)
	assert.NoError(t, err)

	// Shared secrets must match
	assert.Equal(t, aliceSharedSecret, bobSharedSecret)
}

func TestX3DH_InvalidBundleRejected(t *testing.T) {
	alice := crypto.NewX3DHSession()
	invalidBundle := &crypto.PrekeyBundle{} // Empty bundle

	_, _, err := alice.InitiateKeyExchange(invalidBundle)
	assert.Error(t, err, "Should reject invalid bundle")
}
