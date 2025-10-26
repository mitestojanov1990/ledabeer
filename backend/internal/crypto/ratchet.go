package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

// DoubleRatchet implements a simplified Double Ratchet algorithm
type DoubleRatchet struct {
	rootKey    []byte
	chainKey   []byte
	messageNum uint32
	isSender   bool
}

// NewDoubleRatchet creates a new Double Ratchet with a shared secret
func NewDoubleRatchet(sharedSecret []byte, isSender bool) *DoubleRatchet {
	return &DoubleRatchet{
		rootKey:    sharedSecret,
		chainKey:   deriveKey(sharedSecret, []byte("chain")),
		messageNum: 0,
		isSender:   isSender,
	}
}

// Encrypt encrypts a message using the Double Ratchet
func (r *DoubleRatchet) Encrypt(plaintext []byte) ([]byte, error) {
	// Derive message key from chain key
	messageKey := r.deriveMessageKey()

	// Encrypt the plaintext
	ciphertext, err := encryptAESGCM(messageKey, plaintext)
	if err != nil {
		return nil, err
	}

	// Prepend message number
	result := make([]byte, 4+len(ciphertext))
	binary.BigEndian.PutUint32(result[:4], r.messageNum)
	copy(result[4:], ciphertext)

	// Advance the ratchet
	r.advanceChainKey()
	r.messageNum++

	return result, nil
}

// Decrypt decrypts a message using the Double Ratchet
func (r *DoubleRatchet) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 4 {
		return nil, errors.New("ciphertext too short")
	}

	// Extract message number
	messageNum := binary.BigEndian.Uint32(ciphertext[:4])

	// Check if this is the expected message number (simplified - no out-of-order handling)
	if messageNum != r.messageNum {
		return nil, errors.New("unexpected message number - forward secrecy violated")
	}

	// Derive message key from chain key
	messageKey := r.deriveMessageKey()

	// Decrypt the ciphertext
	plaintext, err := decryptAESGCM(messageKey, ciphertext[4:])
	if err != nil {
		return nil, err
	}

	// Advance the ratchet
	r.advanceChainKey()
	r.messageNum++

	return plaintext, nil
}

// deriveMessageKey derives a message key from the current chain key
func (r *DoubleRatchet) deriveMessageKey() []byte {
	return deriveKey(r.chainKey, []byte("message"))
}

// advanceChainKey advances the chain key (KDF ratchet step)
func (r *DoubleRatchet) advanceChainKey() {
	r.chainKey = deriveKey(r.chainKey, []byte("advance"))
}

// deriveKey derives a key using HMAC-SHA256
func deriveKey(key []byte, info []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(info)
	return h.Sum(nil)
}

// encryptAESGCM encrypts data using AES-256-GCM
func encryptAESGCM(key []byte, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAESGCM decrypts data using AES-256-GCM
func decryptAESGCM(key []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
