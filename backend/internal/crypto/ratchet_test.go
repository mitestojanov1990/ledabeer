package crypto_test

import (
	"crypto/rand"
	"testing"

	"ledabeer/backend/internal/crypto"

	"github.com/stretchr/testify/assert"
)

func TestDoubleRatchet_EncryptDecrypt(t *testing.T) {
	// Setup two ratchets with shared secret from X3DH
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	aliceRatchet := crypto.NewDoubleRatchet(sharedSecret, true) // sender
	bobRatchet := crypto.NewDoubleRatchet(sharedSecret, false)  // receiver

	plaintext := []byte("Hello Bob!")

	// Alice encrypts
	ciphertext, err := aliceRatchet.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	// Bob decrypts
	decrypted, err := bobRatchet.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDoubleRatchet_ForwardSecrecy(t *testing.T) {
	// Old keys should not decrypt new messages
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	alice := crypto.NewDoubleRatchet(sharedSecret, true)
	bob := crypto.NewDoubleRatchet(sharedSecret, false)

	// First message
	msg1 := []byte("Message 1")
	cipher1, _ := alice.Encrypt(msg1)
	bob.Decrypt(cipher1)

	// Second message (ratchet advances)
	msg2 := []byte("Message 2")
	cipher2, _ := alice.Encrypt(msg2)

	// Try to decrypt cipher2 with old bob state - should fail
	bobOld := crypto.NewDoubleRatchet(sharedSecret, false)
	_, err := bobOld.Decrypt(cipher2)
	assert.Error(t, err, "Old keys should not decrypt new messages")
}

func TestDoubleRatchet_MultipleMessages(t *testing.T) {
	// Should handle multiple messages in sequence
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	alice := crypto.NewDoubleRatchet(sharedSecret, true)
	bob := crypto.NewDoubleRatchet(sharedSecret, false)

	messages := []string{"Message 1", "Message 2", "Message 3"}

	for _, msg := range messages {
		ciphertext, err := alice.Encrypt([]byte(msg))
		assert.NoError(t, err)

		plaintext, err := bob.Decrypt(ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, msg, string(plaintext))
	}
}
