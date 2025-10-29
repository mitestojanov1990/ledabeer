package e2ee

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"ledabeer/backend/internal/user"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

// E2EEService handles end-to-end encryption for authenticated users
type E2EEService struct {
	userManager *user.UserManager
}

// NewE2EEService creates a new E2EEService
func NewE2EEService(userManager *user.UserManager) *E2EEService {
	return &E2EEService{
		userManager: userManager,
	}
}

// EncryptedMessage represents an encrypted message
type EncryptedMessage struct {
	MessageID     string    `json:"message_id"`
	FromUserID    string    `json:"from_user_id"`
	ToUserID      string    `json:"to_user_id"`
	EncryptedData []byte    `json:"encrypted_data"`
	Nonce         []byte    `json:"nonce"`
	Timestamp     time.Time `json:"timestamp"`
	KeyID         string    `json:"key_id"` // Which key was used for encryption
}

// KeyExchangeRequest represents a key exchange request
type KeyExchangeRequest struct {
	FromUserID    string `json:"from_user_id"`
	ToUserID      string `json:"to_user_id"`
	IdentityKey   []byte `json:"identity_key"`
	SignedPreKey  []byte `json:"signed_pre_key"`
	OneTimeKey    []byte `json:"one_time_key"`
	Signature     []byte `json:"signature"`
}

// KeyExchangeResponse represents a key exchange response
type KeyExchangeResponse struct {
	FromUserID    string `json:"from_user_id"`
	ToUserID      string `json:"to_user_id"`
	IdentityKey   []byte `json:"identity_key"`
	SignedPreKey  []byte `json:"signed_pre_key"`
	OneTimeKey    []byte `json:"one_time_key"`
	Signature     []byte `json:"signature"`
	Timestamp     time.Time `json:"timestamp"`
}

// EncryptMessage encrypts a message for a specific user
func (e *E2EEService) EncryptMessage(fromUserID, toUserID, messageID string, plaintext []byte) (*EncryptedMessage, error) {
	// Get sender's keys
	senderKeys, err := e.userManager.GetUserKeys(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender keys: %w", err)
	}

	// Get recipient's keys
	recipientKeys, err := e.userManager.GetUserKeys(toUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient keys: %w", err)
	}

	// Perform key agreement (simplified X3DH)
	sharedSecret, err := e.performKeyAgreement(senderKeys, recipientKeys)
	if err != nil {
		return nil, fmt.Errorf("key agreement failed: %w", err)
	}

	// Derive encryption key from shared secret
	encryptionKey := e.deriveEncryptionKey(sharedSecret, fromUserID, toUserID)

	// Encrypt the message
	encryptedData, nonce, err := e.encryptData(plaintext, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return &EncryptedMessage{
		MessageID:     messageID,
		FromUserID:    fromUserID,
		ToUserID:      toUserID,
		EncryptedData: encryptedData,
		Nonce:         nonce,
		Timestamp:     time.Now(),
		KeyID:         fmt.Sprintf("%s_%s", fromUserID, toUserID),
	}, nil
}

// DecryptMessage decrypts a message for a specific user
func (e *E2EEService) DecryptMessage(userID string, encryptedMsg *EncryptedMessage) ([]byte, error) {
	// Get user's keys
	userKeys, err := e.userManager.GetUserKeys(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user keys: %w", err)
	}

	// Get sender's keys
	senderKeys, err := e.userManager.GetUserKeys(encryptedMsg.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender keys: %w", err)
	}

	// Perform key agreement (same as encryption)
	sharedSecret, err := e.performKeyAgreement(senderKeys, userKeys)
	if err != nil {
		return nil, fmt.Errorf("key agreement failed: %w", err)
	}

	// Derive encryption key from shared secret
	encryptionKey := e.deriveEncryptionKey(sharedSecret, encryptedMsg.FromUserID, userID)

	// Decrypt the message
	plaintext, err := e.decryptData(encryptedMsg.EncryptedData, encryptedMsg.Nonce, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// InitiateKeyExchange initiates a key exchange between two users
func (e *E2EEService) InitiateKeyExchange(fromUserID, toUserID string) (*KeyExchangeRequest, error) {
	// Get sender's keys
	senderKeys, err := e.userManager.GetUserKeys(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender keys: %w", err)
	}

	// Create key exchange request
	request := &KeyExchangeRequest{
		FromUserID:   fromUserID,
		ToUserID:     toUserID,
		IdentityKey:  senderKeys.IdentityKey,
		SignedPreKey: senderKeys.SignedPreKey,
		OneTimeKey:   senderKeys.OneTimeKeys[0], // Use first one-time key
		Signature:    []byte("signature_placeholder"), // In real implementation, sign the keys
	}

	return request, nil
}

// RespondToKeyExchange responds to a key exchange request
func (e *E2EEService) RespondToKeyExchange(request *KeyExchangeRequest) (*KeyExchangeResponse, error) {
	// Get recipient's keys
	recipientKeys, err := e.userManager.GetUserKeys(request.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient keys: %w", err)
	}

	// Create key exchange response
	response := &KeyExchangeResponse{
		FromUserID:   request.ToUserID,
		ToUserID:     request.FromUserID,
		IdentityKey:  recipientKeys.IdentityKey,
		SignedPreKey: recipientKeys.SignedPreKey,
		OneTimeKey:   recipientKeys.OneTimeKeys[0], // Use first one-time key
		Signature:    []byte("signature_placeholder"), // In real implementation, sign the keys
		Timestamp:    time.Now(),
	}

	return response, nil
}

// performKeyAgreement performs key agreement using X3DH (simplified)
func (e *E2EEService) performKeyAgreement(senderKeys, recipientKeys *user.UserKeys) ([]byte, error) {
	// In a real implementation, this would perform the X3DH key agreement protocol
	// For now, we'll use a simplified approach with Curve25519
	
	// Generate ephemeral key pair for sender
	ephemeralPrivate, err := curve25519.X25519(recipientKeys.IdentityKey, senderKeys.IdentityKey)
	if err != nil {
		return nil, err
	}

	// Use the ephemeral key as shared secret (simplified)
	sharedSecret := sha256.Sum256(ephemeralPrivate)
	return sharedSecret[:], nil
}

// deriveEncryptionKey derives an encryption key from shared secret
func (e *E2EEService) deriveEncryptionKey(sharedSecret []byte, fromUserID, toUserID string) []byte {
	// Use HKDF or similar key derivation function
	// For now, we'll use a simple hash
	keyMaterial := append(sharedSecret, []byte(fromUserID+toUserID)...)
	keyHash := sha256.Sum256(keyMaterial)
	return keyHash[:32] // Use first 32 bytes as encryption key
}

// encryptData encrypts data using ChaCha20-Poly1305
func (e *E2EEService) encryptData(plaintext []byte, key []byte) ([]byte, []byte, error) {
	if len(key) != 32 {
		return nil, nil, errors.New("encryption key must be 32 bytes")
	}

	// Generate random nonce
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	// Create cipher
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt
	encrypted := cipher.Seal(nil, nonce, plaintext, nil)
	return encrypted, nonce, nil
}

// decryptData decrypts data using ChaCha20-Poly1305
func (e *E2EEService) decryptData(ciphertext []byte, nonce []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes")
	}
	if len(nonce) != 12 {
		return nil, errors.New("nonce must be 12 bytes")
	}

	// Create cipher
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
