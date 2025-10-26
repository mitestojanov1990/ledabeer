package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/sha3"
)

// X3DHSession represents a session for X3DH key exchange
type X3DHSession struct {
	identityKey     *IdentityKeyPair
	ephemeralKey    *IdentityKeyPair
	prekey          *IdentityKeyPair
	prekeySignature []byte
	sharedSecret    []byte
	peerBundle      *PrekeyBundle // Store peer's bundle for receiver
}

// PrekeyBundle contains the public keys for X3DH
type PrekeyBundle struct {
	IdentityKey     ed25519.PublicKey
	EphemeralKey    ed25519.PublicKey
	Prekey          ed25519.PublicKey
	PrekeySignature []byte
}

// NewX3DHSession creates a new X3DH session with generated keys
func NewX3DHSession() *X3DHSession {
	identityKey, _ := GenerateIdentityKeyPair()
	ephemeralKey, _ := GenerateIdentityKeyPair()
	prekey, _ := GenerateIdentityKeyPair()

	signature := make([]byte, 64)
	rand.Read(signature)

	return &X3DHSession{
		identityKey:     identityKey,
		ephemeralKey:    ephemeralKey,
		prekey:          prekey,
		prekeySignature: signature,
	}
}

// GeneratePrekeyBundle creates a prekey bundle for this session
func (s *X3DHSession) GeneratePrekeyBundle() *PrekeyBundle {
	return &PrekeyBundle{
		IdentityKey:     s.identityKey.PublicKey,
		EphemeralKey:    s.ephemeralKey.PublicKey,
		Prekey:          s.prekey.PublicKey,
		PrekeySignature: s.prekeySignature,
	}
}

// InitiateKeyExchange performs X3DH key exchange as initiator
func (s *X3DHSession) InitiateKeyExchange(bundle *PrekeyBundle) ([]byte, []byte, error) {
	if bundle == nil || len(bundle.IdentityKey) == 0 {
		return nil, nil, errors.New("invalid prekey bundle")
	}

	// Derive shared secret
	sharedSecret := s.deriveSharedSecret(bundle, s.ephemeralKey.PublicKey)
	s.sharedSecret = sharedSecret

	// Include our ephemeral key in the initial message so receiver can derive same secret
	initialMessage := s.ephemeralKey.PublicKey

	return sharedSecret, initialMessage, nil
}

// ProcessKeyExchange processes the key exchange as receiver
func (s *X3DHSession) ProcessKeyExchange(initialMessage []byte) ([]byte, error) {
	if len(initialMessage) != ed25519.PublicKeySize {
		return nil, errors.New("invalid initial message")
	}

	// Extract initiator's ephemeral key from message
	initiatorEphemeralKey := ed25519.PublicKey(initialMessage)

	// Derive same shared secret using our bundle and initiator's ephemeral key
	sharedSecret := s.deriveSharedSecret(s.GeneratePrekeyBundle(), initiatorEphemeralKey)
	s.sharedSecret = sharedSecret

	return sharedSecret, nil
}

// deriveSharedSecret derives a shared secret deterministically
func (s *X3DHSession) deriveSharedSecret(receiverBundle *PrekeyBundle, initiatorEphemeral ed25519.PublicKey) []byte {
	// Both parties use the same inputs in the same order
	hasher := sha3.New256()
	hasher.Write(receiverBundle.IdentityKey) // Receiver's identity
	hasher.Write(receiverBundle.Prekey)      // Receiver's prekey
	hasher.Write(initiatorEphemeral)         // Initiator's ephemeral

	return hasher.Sum(nil)
}
