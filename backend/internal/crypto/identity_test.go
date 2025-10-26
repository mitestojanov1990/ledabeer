package crypto_test

import (
	"testing"

	"ledabeer/backend/internal/crypto"

	"github.com/stretchr/testify/assert"
)

func TestGenerateIdentityKeyPair(t *testing.T) {
	// Should generate unique Ed25519 identity keys
	key1, err := crypto.GenerateIdentityKeyPair()
	assert.NoError(t, err)
	assert.NotNil(t, key1.PublicKey)
	assert.NotNil(t, key1.PrivateKey)

	key2, err := crypto.GenerateIdentityKeyPair()
	assert.NoError(t, err)
	assert.NotEqual(t, key1.PublicKey, key2.PublicKey, "Keys must be unique")
}

func TestSerializeIdentityKeys(t *testing.T) {
	// Should serialize/deserialize without loss
	original, _ := crypto.GenerateIdentityKeyPair()
	bytes := original.Serialize()
	restored, err := crypto.DeserializeIdentityKeyPair(bytes)

	assert.NoError(t, err)
	assert.Equal(t, original.PublicKey, restored.PublicKey)
}
