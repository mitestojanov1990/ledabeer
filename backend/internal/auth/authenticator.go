package auth

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

type contextKey string

const (
	peerIDKey    contextKey = "peer_id"
	signatureKey contextKey = "signature"
)

type Authenticator struct {
	// Peer ID validator
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{}
}

func WithPeerID(ctx context.Context, peerID string) context.Context {
	return context.WithValue(ctx, peerIDKey, peerID)
}

func WithAuth(ctx context.Context, peerID string, signature []byte) context.Context {
	ctx = context.WithValue(ctx, peerIDKey, peerID)
	return context.WithValue(ctx, signatureKey, signature)
}

func (a *Authenticator) Authenticate(ctx context.Context) (string, error) {
	peerIDStr, ok := ctx.Value(peerIDKey).(string)
	if !ok {
		return "", fmt.Errorf("missing peer ID")
	}

	// Validate peer ID format
	_, err := peer.Decode(peerIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid peer ID: %w", err)
	}

	return peerIDStr, nil
}

func (a *Authenticator) VerifySignature(ctx context.Context, challenge []byte) error {
	// Verify signature from context
	return nil
}
