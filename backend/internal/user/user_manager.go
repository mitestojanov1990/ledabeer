package user

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"ledabeer/backend/internal/auth"
	"github.com/libp2p/go-libp2p/core/peer"
)

// UserManager manages the connection between authenticated users and P2P network
type UserManager struct {
	// Maps user ID to peer ID
	userToPeer map[string]peer.ID
	// Maps peer ID to user ID
	peerToUser map[peer.ID]string
	// Maps user ID to user info
	userInfo   map[string]*UserInfo
	// Maps user ID to encryption keys
	userKeys   map[string]*UserKeys
	mutex      sync.RWMutex
}

// UserInfo contains user information for P2P operations
type UserInfo struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	PeerID      peer.ID   `json:"peer_id"`
	IsOnline    bool      `json:"is_online"`
	LastSeen    time.Time `json:"last_seen"`
	PublicKey   []byte    `json:"public_key"`
}

// UserKeys contains encryption keys for a user
type UserKeys struct {
	IdentityKey    []byte `json:"identity_key"`    // Long-term identity key
	SignedPreKey   []byte `json:"signed_pre_key"`  // Signed pre-key
	OneTimeKeys    [][]byte `json:"one_time_keys"` // One-time keys
	LastKeyUpdate  time.Time `json:"last_key_update"`
}

// NewUserManager creates a new UserManager
func NewUserManager() *UserManager {
	return &UserManager{
		userToPeer: make(map[string]peer.ID),
		peerToUser: make(map[peer.ID]string),
		userInfo:   make(map[string]*UserInfo),
		userKeys:   make(map[string]*UserKeys),
	}
}

// RegisterUser registers an authenticated user with the P2P network
func (um *UserManager) RegisterUser(authUser *auth.User, peerID peer.ID, publicKey []byte) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	// Check if user is already registered
	if existingPeerID, exists := um.userToPeer[authUser.ID]; exists {
		if existingPeerID == peerID {
			// Update existing registration
			um.updateUserInfo(authUser, peerID, publicKey)
			return nil
		}
		return errors.New("user already registered with different peer")
	}

	// Check if peer is already registered
	if existingUserID, exists := um.peerToUser[peerID]; exists {
		return fmt.Errorf("peer %s already registered with user %s", peerID.String(), existingUserID)
	}

	// Register user
	um.userToPeer[authUser.ID] = peerID
	um.peerToUser[peerID] = authUser.ID

	// Create user info
	userInfo := &UserInfo{
		UserID:      authUser.ID,
		Username:    authUser.Username,
		DisplayName: authUser.DisplayName,
		Email:       authUser.Email,
		AvatarURL:   authUser.AvatarURL,
		PeerID:      peerID,
		IsOnline:    true,
		LastSeen:    time.Now(),
		PublicKey:   publicKey,
	}
	um.userInfo[authUser.ID] = userInfo

	// Generate encryption keys for user
	userKeys := &UserKeys{
		IdentityKey:   generateIdentityKey(),
		SignedPreKey:  generateSignedPreKey(),
		OneTimeKeys:   generateOneTimeKeys(10), // Generate 10 one-time keys
		LastKeyUpdate: time.Now(),
	}
	um.userKeys[authUser.ID] = userKeys

	return nil
}

// UnregisterUser removes a user from the P2P network
func (um *UserManager) UnregisterUser(userID string) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	peerID, exists := um.userToPeer[userID]
	if !exists {
		return errors.New("user not registered")
	}

	// Remove mappings
	delete(um.userToPeer, userID)
	delete(um.peerToUser, peerID)
	delete(um.userInfo, userID)
	delete(um.userKeys, userID)

	return nil
}

// GetUserByPeerID gets user info by peer ID
func (um *UserManager) GetUserByPeerID(peerID peer.ID) (*UserInfo, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	userID, exists := um.peerToUser[peerID]
	if !exists {
		return nil, errors.New("peer not registered")
	}

	userInfo, exists := um.userInfo[userID]
	if !exists {
		return nil, errors.New("user info not found")
	}

	return userInfo, nil
}

// GetPeerIDByUserID gets peer ID by user ID
func (um *UserManager) GetPeerIDByUserID(userID string) (peer.ID, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	peerID, exists := um.userToPeer[userID]
	if !exists {
		return "", errors.New("user not registered")
	}

	return peerID, nil
}

// GetUserKeys gets encryption keys for a user
func (um *UserManager) GetUserKeys(userID string) (*UserKeys, error) {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	keys, exists := um.userKeys[userID]
	if !exists {
		return nil, errors.New("user keys not found")
	}

	return keys, nil
}

// UpdateUserOnlineStatus updates the online status of a user
func (um *UserManager) UpdateUserOnlineStatus(userID string, isOnline bool) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	userInfo, exists := um.userInfo[userID]
	if !exists {
		return errors.New("user not found")
	}

	userInfo.IsOnline = isOnline
	if isOnline {
		userInfo.LastSeen = time.Now()
	}

	return nil
}

// GetAllOnlineUsers gets all currently online users
func (um *UserManager) GetAllOnlineUsers() []*UserInfo {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	var onlineUsers []*UserInfo
	for _, userInfo := range um.userInfo {
		if userInfo.IsOnline {
			onlineUsers = append(onlineUsers, userInfo)
		}
	}

	return onlineUsers
}

// SearchUsers searches for users by username, email, or display name
func (um *UserManager) SearchUsers(query string) []*UserInfo {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	var results []*UserInfo
	query = strings.ToLower(query)

	for _, userInfo := range um.userInfo {
		if strings.Contains(strings.ToLower(userInfo.Username), query) ||
			strings.Contains(strings.ToLower(userInfo.DisplayName), query) ||
			strings.Contains(strings.ToLower(userInfo.Email), query) {
			results = append(results, userInfo)
		}
	}

	return results
}

// updateUserInfo updates user information
func (um *UserManager) updateUserInfo(authUser *auth.User, peerID peer.ID, publicKey []byte) {
	if userInfo, exists := um.userInfo[authUser.ID]; exists {
		userInfo.Username = authUser.Username
		userInfo.DisplayName = authUser.DisplayName
		userInfo.Email = authUser.Email
		userInfo.AvatarURL = authUser.AvatarURL
		userInfo.PeerID = peerID
		userInfo.PublicKey = publicKey
		userInfo.IsOnline = true
		userInfo.LastSeen = time.Now()
	}
}

// Helper functions for key generation (simplified)
func generateIdentityKey() []byte {
	// In a real implementation, this would generate a proper identity key
	return []byte("identity_key_placeholder")
}

func generateSignedPreKey() []byte {
	// In a real implementation, this would generate a proper signed pre-key
	return []byte("signed_pre_key_placeholder")
}

func generateOneTimeKeys(count int) [][]byte {
	keys := make([][]byte, count)
	for i := 0; i < count; i++ {
		keys[i] = []byte(fmt.Sprintf("one_time_key_%d", i))
	}
	return keys
}
