package integration_test

import (
	"context"
	"testing"
	"time"

	"ledabeer/backend/internal/crypto"
	"ledabeer/backend/internal/network"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_EncryptedMessageExchange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create two nodes
	alice, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer alice.Close()

	bob, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer bob.Close()

	// Perform X3DH key exchange
	aliceSession := crypto.NewX3DHSession()
	bobSession := crypto.NewX3DHSession()

	// Bob generates prekey bundle
	bobBundle := bobSession.GeneratePrekeyBundle()

	// Alice initiates key exchange with Bob's bundle
	aliceSharedSecret, aliceEphemeralKey, err := aliceSession.InitiateKeyExchange(bobBundle)
	require.NoError(t, err)

	// Bob processes Alice's key exchange
	bobSharedSecret, err := bobSession.ProcessKeyExchange(aliceEphemeralKey)
	require.NoError(t, err)

	// Verify both parties have the same shared secret
	assert.Equal(t, aliceSharedSecret, bobSharedSecret, "Both parties should derive the same shared secret")

	// Create Double Ratchet instances with shared secret
	aliceRatchet := crypto.NewDoubleRatchet(aliceSharedSecret, true) // Alice is sender
	bobRatchet := crypto.NewDoubleRatchet(bobSharedSecret, false)    // Bob is receiver

	// Test message exchange
	originalMessage := "Hello Bob, this is Alice!"

	// Alice encrypts message
	encryptedMessage, err := aliceRatchet.Encrypt([]byte(originalMessage))
	require.NoError(t, err)

	// Bob decrypts message
	decryptedMessage, err := bobRatchet.Decrypt(encryptedMessage)
	require.NoError(t, err)

	assert.Equal(t, originalMessage, string(decryptedMessage), "Decrypted message should match original")

	// Test reverse direction
	replyMessage := "Hello Alice, this is Bob!"

	// Bob encrypts reply
	encryptedReply, err := bobRatchet.Encrypt([]byte(replyMessage))
	require.NoError(t, err)

	// Alice decrypts reply
	decryptedReply, err := aliceRatchet.Decrypt(encryptedReply)
	require.NoError(t, err)

	assert.Equal(t, replyMessage, string(decryptedReply), "Decrypted reply should match original")
}

func TestE2E_MultipleMessageExchange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create two nodes
	alice, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer alice.Close()

	bob, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer bob.Close()

	// Perform X3DH key exchange
	aliceSession := crypto.NewX3DHSession()
	bobSession := crypto.NewX3DHSession()

	bobBundle := bobSession.GeneratePrekeyBundle()

	aliceSharedSecret, aliceEphemeralKey, err := aliceSession.InitiateKeyExchange(bobBundle)
	require.NoError(t, err)

	bobSharedSecret, err := bobSession.ProcessKeyExchange(aliceEphemeralKey)
	require.NoError(t, err)

	// Create ratchets
	aliceRatchet := crypto.NewDoubleRatchet(aliceSharedSecret, true) // Alice is sender
	bobRatchet := crypto.NewDoubleRatchet(bobSharedSecret, false)    // Bob is receiver

	// Exchange multiple messages
	messages := []string{
		"Message 1: Hello!",
		"Message 2: How are you?",
		"Message 3: I'm doing great!",
		"Message 4: That's wonderful!",
		"Message 5: Let's continue chatting!",
	}

	for i, message := range messages {
		// Alternate between Alice and Bob sending
		if i%2 == 0 {
			// Alice sends to Bob
			encrypted, err := aliceRatchet.Encrypt([]byte(message))
			require.NoError(t, err)

			decrypted, err := bobRatchet.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, message, string(decrypted), "Message %d should be decrypted correctly", i+1)
		} else {
			// Bob sends to Alice
			encrypted, err := bobRatchet.Encrypt([]byte(message))
			require.NoError(t, err)

			decrypted, err := aliceRatchet.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, message, string(decrypted), "Message %d should be decrypted correctly", i+1)
		}
	}
}
