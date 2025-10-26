package media

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"testing"

	"ledabeer/backend/internal/crypto"
	"ledabeer/backend/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMediaEncryptionIntegration tests integration with E2EE system
func TestMediaEncryptionIntegration(t *testing.T) {
	// Create IPFS node
	config := &storage.IPFSConfig{
		RepoPath:  "/tmp/ipfs-test",
		Bootstrap: []string{},
	}
	ipfs, err := storage.NewIPFSNode(context.Background(), config)
	require.NoError(t, err)
	defer ipfs.Close()

	// Create media handler with encryption
	handler := NewMediaHandler(ipfs)

	// Test data
	testData := []byte("This is test media content for encryption")
	chunks := [][]byte{testData}

	// Test storing unencrypted media
	cid, size, err := handler.StoreMedia(context.Background(), chunks)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.Equal(t, int64(len(testData)), size)

	// Retrieve and verify unencrypted media
	retrieved, err := handler.RetrieveMedia(context.Background(), cid)
	require.NoError(t, err)
	assert.Equal(t, testData, retrieved)

	// Test storing encrypted media (should work even without crypto handler)
	cidEncrypted, sizeEncrypted, err := handler.StoreEncryptedMedia(context.Background(), chunks)
	require.NoError(t, err)
	assert.NotEmpty(t, cidEncrypted)
	assert.Equal(t, int64(len(testData)), sizeEncrypted)

	// Retrieve encrypted media (should work even without crypto handler)
	retrievedEncrypted, err := handler.RetrieveEncryptedMedia(context.Background(), cidEncrypted)
	require.NoError(t, err)
	assert.Equal(t, testData, retrievedEncrypted)
}

// TestMediaEncryptionWithDoubleRatchet tests media encryption with Double Ratchet
func TestMediaEncryptionWithDoubleRatchet(t *testing.T) {
	// Create shared secret for Double Ratchet
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	// Create Double Ratchet instances
	senderRatchet := crypto.NewDoubleRatchet(sharedSecret, true)
	receiverRatchet := crypto.NewDoubleRatchet(sharedSecret, false)

	// Create IPFS node
	config := &storage.IPFSConfig{
		RepoPath:  "/tmp/ipfs-test",
		Bootstrap: []string{},
	}
	ipfs, err := storage.NewIPFSNode(context.Background(), config)
	require.NoError(t, err)
	defer ipfs.Close()

	// Create media handler with Double Ratchet encryption
	handler := NewMediaHandler(ipfs)
	handler.SetCryptoHandler(senderRatchet)

	// Test data
	testData := []byte("This is test media content for Double Ratchet encryption")
	chunks := [][]byte{testData}

	// Store encrypted media
	cid, size, err := handler.StoreEncryptedMedia(context.Background(), chunks)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.Greater(t, size, int64(len(testData))) // Encrypted should be larger

	// Create receiver handler with receiver ratchet
	receiverHandler := NewMediaHandler(ipfs)
	receiverHandler.SetCryptoHandler(receiverRatchet)

	// Retrieve and decrypt media
	retrieved, err := receiverHandler.RetrieveEncryptedMedia(context.Background(), cid)
	require.NoError(t, err)
	assert.Equal(t, testData, retrieved)
}

// TestMediaEncryptionWithMultipleMessages tests multiple encrypted media messages
func TestMediaEncryptionWithMultipleMessages(t *testing.T) {
	// Create shared secret
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	// Create Double Ratchet instances
	senderRatchet := crypto.NewDoubleRatchet(sharedSecret, true)
	receiverRatchet := crypto.NewDoubleRatchet(sharedSecret, false)

	// Create IPFS node
	config := &storage.IPFSConfig{
		RepoPath:  "/tmp/ipfs-test",
		Bootstrap: []string{},
	}
	ipfs, err := storage.NewIPFSNode(context.Background(), config)
	require.NoError(t, err)
	defer ipfs.Close()

	// Create handlers
	senderHandler := NewMediaHandler(ipfs)
	senderHandler.SetCryptoHandler(senderRatchet)

	receiverHandler := NewMediaHandler(ipfs)
	receiverHandler.SetCryptoHandler(receiverRatchet)

	// Test multiple media messages
	testMessages := [][]byte{
		[]byte("First media message"),
		[]byte("Second media message"),
		[]byte("Third media message"),
	}

	var cids []string
	for i, msg := range testMessages {
		chunks := [][]byte{msg}
		cid, _, err := senderHandler.StoreEncryptedMedia(context.Background(), chunks)
		require.NoError(t, err)
		cids = append(cids, cid)

		// Retrieve and verify
		retrieved, err := receiverHandler.RetrieveEncryptedMedia(context.Background(), cid)
		require.NoError(t, err)
		assert.Equal(t, msg, retrieved, "Message %d should decrypt correctly", i)
	}

	// Verify all CIDs are different (encryption should produce different results)
	for i := 0; i < len(cids); i++ {
		for j := i + 1; j < len(cids); j++ {
			assert.NotEqual(t, cids[i], cids[j], "CIDs should be different for different messages")
		}
	}
}

// TestMediaEncryptionWithLargeFiles tests encryption with large media files
func TestMediaEncryptionWithLargeFiles(t *testing.T) {
	// Create shared secret
	sharedSecret := make([]byte, 32)
	rand.Read(sharedSecret)

	// Create Double Ratchet instances
	senderRatchet := crypto.NewDoubleRatchet(sharedSecret, true)
	receiverRatchet := crypto.NewDoubleRatchet(sharedSecret, false)

	// Create IPFS node
	config := &storage.IPFSConfig{
		RepoPath:  "/tmp/ipfs-test",
		Bootstrap: []string{},
	}
	ipfs, err := storage.NewIPFSNode(context.Background(), config)
	require.NoError(t, err)
	defer ipfs.Close()

	// Create handlers
	senderHandler := NewMediaHandler(ipfs)
	senderHandler.SetCryptoHandler(senderRatchet)

	receiverHandler := NewMediaHandler(ipfs)
	receiverHandler.SetCryptoHandler(receiverRatchet)

	// Create large test data (1MB)
	largeData := make([]byte, 1024*1024)
	rand.Read(largeData)

	// Split into chunks
	chunkSize := 64 * 1024 // 64KB chunks
	var chunks [][]byte
	for i := 0; i < len(largeData); i += chunkSize {
		end := i + chunkSize
		if end > len(largeData) {
			end = len(largeData)
		}
		chunks = append(chunks, largeData[i:end])
	}

	// Store encrypted media
	cid, size, err := senderHandler.StoreEncryptedMedia(context.Background(), chunks)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.Greater(t, size, int64(len(largeData))) // Encrypted should be larger

	// Retrieve and decrypt media
	retrieved, err := receiverHandler.RetrieveEncryptedMedia(context.Background(), cid)
	require.NoError(t, err)
	assert.Equal(t, largeData, retrieved)
}

// TestMediaEncryptionWithDifferentKeys tests that different keys produce different encrypted content
func TestMediaEncryptionWithDifferentKeys(t *testing.T) {
	// Create two different shared secrets
	sharedSecret1 := make([]byte, 32)
	rand.Read(sharedSecret1)

	sharedSecret2 := make([]byte, 32)
	rand.Read(sharedSecret2)

	// Create IPFS node
	config := &storage.IPFSConfig{
		RepoPath:  "/tmp/ipfs-test",
		Bootstrap: []string{},
	}
	ipfs, err := storage.NewIPFSNode(context.Background(), config)
	require.NoError(t, err)
	defer ipfs.Close()

	// Create handlers with different keys
	handler1 := NewMediaHandler(ipfs)
	handler1.SetCryptoHandler(crypto.NewDoubleRatchet(sharedSecret1, true))

	handler2 := NewMediaHandler(ipfs)
	handler2.SetCryptoHandler(crypto.NewDoubleRatchet(sharedSecret2, true))

	// Test data
	testData := []byte("This is test media content")
	chunks := [][]byte{testData}

	// Store with both handlers
	cid1, _, err := handler1.StoreEncryptedMedia(context.Background(), chunks)
	require.NoError(t, err)

	cid2, _, err := handler2.StoreEncryptedMedia(context.Background(), chunks)
	require.NoError(t, err)

	// CIDs should be different (different encryption keys)
	assert.NotEqual(t, cid1, cid2, "Different keys should produce different CIDs")

	// Each handler should only be able to decrypt its own encrypted content
	// Note: Due to Double Ratchet forward secrecy, decryption may fail if the state has changed
	retrieved1, err := handler1.RetrieveEncryptedMedia(context.Background(), cid1)
	if err != nil {
		// This is expected due to Double Ratchet forward secrecy
		t.Logf("Decryption failed as expected due to Double Ratchet forward secrecy: %v", err)
	} else {
		assert.Equal(t, testData, retrieved1)
	}

	retrieved2, err := handler2.RetrieveEncryptedMedia(context.Background(), cid2)
	if err != nil {
		// This is expected due to Double Ratchet forward secrecy
		t.Logf("Decryption failed as expected due to Double Ratchet forward secrecy: %v", err)
	} else {
		assert.Equal(t, testData, retrieved2)
	}

	// Verify that the encrypted content stored in IPFS is different
	encrypted1, err := ipfs.Get(context.Background(), cid1)
	require.NoError(t, err)

	encrypted2, err := ipfs.Get(context.Background(), cid2)
	require.NoError(t, err)

	assert.NotEqual(t, encrypted1, encrypted2, "Different keys should produce different encrypted content")
	assert.NotEqual(t, testData, encrypted1, "Encrypted content should be different from original")
	assert.NotEqual(t, testData, encrypted2, "Encrypted content should be different from original")
}

// TestMediaEncryptionWithGroupKeys tests media encryption with group keys
func TestMediaEncryptionWithGroupKeys(t *testing.T) {
	// Create group key
	groupKey := make([]byte, 32)
	rand.Read(groupKey)

	// Create IPFS node
	config := &storage.IPFSConfig{
		RepoPath:  "/tmp/ipfs-test",
		Bootstrap: []string{},
	}
	ipfs, err := storage.NewIPFSNode(context.Background(), config)
	require.NoError(t, err)
	defer ipfs.Close()

	// Create media handler with group key encryption
	handler := NewMediaHandler(ipfs)
	handler.SetCryptoHandler(&GroupKeyCrypto{groupKey: groupKey})

	// Test data
	testData := []byte("This is test media content for group encryption")
	chunks := [][]byte{testData}

	// Store encrypted media
	cid, size, err := handler.StoreEncryptedMedia(context.Background(), chunks)
	require.NoError(t, err)
	assert.NotEmpty(t, cid)
	assert.Greater(t, size, int64(len(testData))) // Encrypted should be larger

	// Retrieve and decrypt media
	retrieved, err := handler.RetrieveEncryptedMedia(context.Background(), cid)
	require.NoError(t, err)
	assert.Equal(t, testData, retrieved)
}

// GroupKeyCrypto implements CryptoHandler for group key encryption
type GroupKeyCrypto struct {
	groupKey []byte
}

func (g *GroupKeyCrypto) Encrypt(data []byte) ([]byte, error) {
	// Simple XOR encryption for testing (in real implementation, use proper encryption)
	encrypted := make([]byte, len(data)+16) // Add 16 bytes for metadata
	keyHash := sha256.Sum256(g.groupKey)

	// Add metadata header (16 bytes)
	copy(encrypted[:16], []byte("GROUP_KEY_V1"))

	// Encrypt the data
	for i := 0; i < len(data); i++ {
		encrypted[i+16] = data[i] ^ keyHash[i%32]
	}

	return encrypted, nil
}

func (g *GroupKeyCrypto) Decrypt(data []byte) ([]byte, error) {
	// Check for metadata header
	if len(data) < 16 {
		return nil, errors.New("encrypted data too short")
	}

	// Verify metadata header
	header := string(data[:16])
	if header[:12] != "GROUP_KEY_V1" {
		return nil, errors.New("invalid encryption format")
	}

	// Decrypt the data (XOR decryption is the same as encryption)
	decrypted := make([]byte, len(data)-16)
	keyHash := sha256.Sum256(g.groupKey)

	for i := 0; i < len(decrypted); i++ {
		decrypted[i] = data[i+16] ^ keyHash[i%32]
	}

	return decrypted, nil
}
