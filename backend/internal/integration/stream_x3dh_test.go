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

func TestStreamX3DH_CompleteKeyExchange(t *testing.T) {
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

	// Connect Alice to Bob
	bobAddr := bob.Addrs()[0].String() + "/p2p/" + bob.ID().String()
	peerInfo, err := network.ParseAddrInfo(bobAddr)
	require.NoError(t, err)

	err = alice.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	// Create X3DH sessions
	aliceSession := crypto.NewX3DHSession()
	bobSession := crypto.NewX3DHSession()

	// Bob generates prekey bundle
	bobBundle := bobSession.GeneratePrekeyBundle()

	// Create channels for communication
	ephemeralKeyReceived := make(chan []byte, 1)
	encryptedMessageReceived := make(chan []byte, 1)

	// Set up stream handler for ephemeral key
	ephemeralHandler := network.NewStreamHandler(func(data []byte) error {
		ephemeralKeyReceived <- data
		return nil
	})
	bob.SetStreamHandler("/x3dh/ephemeral/1.0.0", ephemeralHandler.Handle)

	// Set up stream handler for encrypted message
	encryptedHandler := network.NewStreamHandler(func(data []byte) error {
		encryptedMessageReceived <- data
		return nil
	})
	bob.SetStreamHandler("/x3dh/encrypted/1.0.0", encryptedHandler.Handle)

	// Alice performs X3DH key exchange
	aliceSharedSecret, aliceEphemeralKey, err := aliceSession.InitiateKeyExchange(bobBundle)
	require.NoError(t, err)

	// Send ephemeral key to Bob
	ephemeralStream, err := alice.NewStream(ctx, bob.ID(), "/x3dh/ephemeral/1.0.0")
	require.NoError(t, err)
	err = network.WriteMessage(ephemeralStream, aliceEphemeralKey)
	require.NoError(t, err)
	ephemeralStream.Close()

	// Wait for Bob to receive the ephemeral key
	select {
	case receivedEphemeralKey := <-ephemeralKeyReceived:
		// Bob processes the key exchange
		bobSharedSecret, err := bobSession.ProcessKeyExchange(receivedEphemeralKey)
		require.NoError(t, err)

		// Verify both parties have the same shared secret
		assert.Equal(t, aliceSharedSecret, bobSharedSecret, "Both parties should derive the same shared secret")

		// Test encrypted communication using the shared secret
		aliceRatchet := crypto.NewDoubleRatchet(aliceSharedSecret, true)
		bobRatchet := crypto.NewDoubleRatchet(bobSharedSecret, false)

		// Alice sends encrypted message to Bob
		message := "Hello Bob, this is an encrypted message!"
		encryptedMessage, err := aliceRatchet.Encrypt([]byte(message))
		require.NoError(t, err)

		// Send encrypted message over stream
		encryptedStream, err := alice.NewStream(ctx, bob.ID(), "/x3dh/encrypted/1.0.0")
		require.NoError(t, err)
		err = network.WriteMessage(encryptedStream, encryptedMessage)
		require.NoError(t, err)
		encryptedStream.Close()

		// Bob receives and decrypts the message
		select {
		case receivedData := <-encryptedMessageReceived:
			decryptedMessage, err := bobRatchet.Decrypt(receivedData)
			require.NoError(t, err)
			assert.Equal(t, message, string(decryptedMessage), "Decrypted message should match original")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for Bob to receive encrypted message")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for Bob to receive ephemeral key")
	}
}

func TestStreamX3DH_MultipleKeyExchanges(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
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

	// Connect Alice to Bob
	bobAddr := bob.Addrs()[0].String() + "/p2p/" + bob.ID().String()
	peerInfo, err := network.ParseAddrInfo(bobAddr)
	require.NoError(t, err)

	err = alice.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	// Test multiple key exchanges
	for i := 0; i < 3; i++ {
		t.Logf("Performing key exchange %d", i+1)

		// Create new sessions for each exchange
		aliceSession := crypto.NewX3DHSession()
		bobSession := crypto.NewX3DHSession()

		// Bob generates prekey bundle
		bobBundle := bobSession.GeneratePrekeyBundle()

		// Create channels for this exchange
		ephemeralKeyReceived := make(chan []byte, 1)
		encryptedMessageReceived := make(chan []byte, 1)

		// Set up stream handlers
		ephemeralHandler := network.NewStreamHandler(func(data []byte) error {
			ephemeralKeyReceived <- data
			return nil
		})
		bob.SetStreamHandler("/x3dh/ephemeral/1.0.0", ephemeralHandler.Handle)

		encryptedHandler := network.NewStreamHandler(func(data []byte) error {
			encryptedMessageReceived <- data
			return nil
		})
		bob.SetStreamHandler("/x3dh/encrypted/1.0.0", encryptedHandler.Handle)

		// Alice performs X3DH key exchange
		aliceSharedSecret, aliceEphemeralKey, err := aliceSession.InitiateKeyExchange(bobBundle)
		require.NoError(t, err)

		// Send ephemeral key to Bob
		ephemeralStream, err := alice.NewStream(ctx, bob.ID(), "/x3dh/ephemeral/1.0.0")
		require.NoError(t, err)
		err = network.WriteMessage(ephemeralStream, aliceEphemeralKey)
		require.NoError(t, err)
		ephemeralStream.Close()

		// Wait for Bob to receive ephemeral key
		select {
		case receivedEphemeralKey := <-ephemeralKeyReceived:
			// Bob processes the key exchange
			bobSharedSecret, err := bobSession.ProcessKeyExchange(receivedEphemeralKey)
			require.NoError(t, err)

			// Verify shared secrets match
			assert.Equal(t, aliceSharedSecret, bobSharedSecret, "Shared secrets should match in exchange %d", i+1)

			// Test encrypted communication
			aliceRatchet := crypto.NewDoubleRatchet(aliceSharedSecret, true)
			bobRatchet := crypto.NewDoubleRatchet(bobSharedSecret, false)

			// Send encrypted message
			message := "Test message " + string(rune(i+1))
			encryptedMessage, err := aliceRatchet.Encrypt([]byte(message))
			require.NoError(t, err)

			encryptedStream, err := alice.NewStream(ctx, bob.ID(), "/x3dh/encrypted/1.0.0")
			require.NoError(t, err)
			err = network.WriteMessage(encryptedStream, encryptedMessage)
			require.NoError(t, err)
			encryptedStream.Close()

			// Bob receives and decrypts
			select {
			case receivedData := <-encryptedMessageReceived:
				decryptedMessage, err := bobRatchet.Decrypt(receivedData)
				require.NoError(t, err)
				assert.Equal(t, message, string(decryptedMessage), "Message should be decrypted correctly in exchange %d", i+1)
			case <-time.After(5 * time.Second):
				t.Fatalf("Timeout waiting for Bob to receive encrypted message in exchange %d", i+1)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for Bob to receive ephemeral key in exchange %d", i+1)
		}
	}
}
