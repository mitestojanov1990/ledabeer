package api_test

import (
	"context"
	"testing"

	"ledabeer/backend/internal/api"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_ValidatePeerID(t *testing.T) {
	// Should validate peer ID from request metadata
	auth := api.NewAuthenticator()

	// Generate a valid peer ID
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	require.NoError(t, err)
	peerID, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	ctx := api.WithPeerID(context.Background(), peerID.String())

	validated, err := auth.Authenticate(ctx)
	require.NoError(t, err)
	assert.Equal(t, peerID.String(), validated)
}

func TestAuth_RejectInvalidPeerID(t *testing.T) {
	// Should reject invalid peer IDs
	auth := api.NewAuthenticator()

	ctx := api.WithPeerID(context.Background(), "invalid-peer-id")

	_, err := auth.Authenticate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid peer ID")
}

func TestAuth_RequirePeerID(t *testing.T) {
	// Should require peer ID in context
	auth := api.NewAuthenticator()

	ctx := context.Background() // No peer ID

	_, err := auth.Authenticate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing peer ID")
}

func TestAuth_SignatureVerification(t *testing.T) {
	// Should verify peer owns the private key
	auth := api.NewAuthenticator()

	// Generate a valid peer ID
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	require.NoError(t, err)
	peerID, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	// Client signs challenge with private key
	challenge := []byte("auth-challenge")
	signature := []byte("test-signature")

	ctx := api.WithAuth(context.Background(), peerID.String(), signature)

	err = auth.VerifySignature(ctx, challenge)
	require.NoError(t, err)
}
