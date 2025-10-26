package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
)

// IdentityKeyPair represents an Ed25519 key pair for peer identity
type IdentityKeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// KeyGenerator interface for generating identity keys
type KeyGenerator interface {
	GenerateIdentityKeyPair() (*IdentityKeyPair, error)
}

// Ed25519KeyGenerator implements KeyGenerator using Ed25519
type Ed25519KeyGenerator struct{}

// NewEd25519KeyGenerator creates a new Ed25519 key generator
func NewEd25519KeyGenerator() *Ed25519KeyGenerator {
	return &Ed25519KeyGenerator{}
}

// GenerateIdentityKeyPair creates a new Ed25519 identity key pair
func (g *Ed25519KeyGenerator) GenerateIdentityKeyPair() (*IdentityKeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &IdentityKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// GenerateIdentityKeyPair creates a new Ed25519 identity key pair (convenience function)
func GenerateIdentityKeyPair() (*IdentityKeyPair, error) {
	generator := NewEd25519KeyGenerator()
	return generator.GenerateIdentityKeyPair()
}

// Serialize returns the private key bytes for storage
func (k *IdentityKeyPair) Serialize() []byte {
	return []byte(k.PrivateKey)
}

// DeserializeIdentityKeyPair reconstructs a key pair from serialized private key
func DeserializeIdentityKeyPair(data []byte) (*IdentityKeyPair, error) {
	if len(data) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}

	privateKey := ed25519.PrivateKey(data)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &IdentityKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}
