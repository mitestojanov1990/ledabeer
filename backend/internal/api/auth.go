package api

import (
	"context"

	"ledabeer/backend/internal/auth"
)

// Re-export for backward compatibility
type Authenticator = auth.Authenticator

func NewAuthenticator() *Authenticator {
	return auth.NewAuthenticator()
}

func WithPeerID(ctx context.Context, peerID string) context.Context {
	return auth.WithPeerID(ctx, peerID)
}

func WithAuth(ctx context.Context, peerID string, signature []byte) context.Context {
	return auth.WithAuth(ctx, peerID, signature)
}
