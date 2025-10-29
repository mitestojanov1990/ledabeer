package integration

import (
	"fmt"
	"sync"

	"ledabeer/backend/internal/auth"
	"ledabeer/backend/internal/calls"
	"ledabeer/backend/internal/e2ee"
	"ledabeer/backend/internal/media"
	"ledabeer/backend/internal/user"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ServiceManager manages all services and their integration
type ServiceManager struct {
	// Core services
	authService    *auth.Authenticator
	userService    *auth.UserService
	userManager    *user.UserManager
	
	// Feature services
	e2eeService    *e2ee.E2EEService
	mediaService   *media.MediaService
	callService    *calls.CallService
	
	// P2P network
	host           host.Host
	
	// State
	initialized    bool
	mutex          sync.RWMutex
}

// NewServiceManager creates a new ServiceManager
func NewServiceManager(host host.Host) *ServiceManager {
	return &ServiceManager{
		host: host,
	}
}

// Initialize initializes all services and their dependencies
func (sm *ServiceManager) Initialize() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.initialized {
		return nil
	}

	// Initialize user repository (in-memory for now)
	userRepo := auth.NewMemoryUserRepository()
	
	// Initialize core services
	sm.authService = auth.NewAuthenticator(userRepo)
	sm.userService = auth.NewUserService(userRepo)
	sm.userManager = user.NewUserManager()
	
	// Initialize feature services
	sm.e2eeService = e2ee.NewE2EEService(sm.userManager)
	sm.mediaService = media.NewMediaService(sm.userManager)
	sm.callService = calls.NewCallService(sm.userManager)
	
	sm.initialized = true
	return nil
}

// GetAuthService returns the authentication service
func (sm *ServiceManager) GetAuthService() *auth.Authenticator {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.authService
}

// GetUserService returns the user service
func (sm *ServiceManager) GetUserService() *auth.UserService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.userService
}

// GetUserManager returns the user manager
func (sm *ServiceManager) GetUserManager() *user.UserManager {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.userManager
}

// GetE2EEService returns the E2EE service
func (sm *ServiceManager) GetE2EEService() *e2ee.E2EEService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.e2eeService
}

// GetMediaService returns the media service
func (sm *ServiceManager) GetMediaService() *media.MediaService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.mediaService
}

// GetCallService returns the call service
func (sm *ServiceManager) GetCallService() *calls.CallService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.callService
}

// GetHost returns the libp2p host
func (sm *ServiceManager) GetHost() host.Host {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.host
}

// RegisterUserWithP2P registers an authenticated user with the P2P network
func (sm *ServiceManager) RegisterUserWithP2P(userID string, peerID string, publicKey []byte) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.initialized {
		return fmt.Errorf("service manager not initialized")
	}

	// Get user from auth service
	authUser, err := sm.authService.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Convert string peerID to libp2p peer.ID
	libp2pPeerID, err := peer.IDFromString(peerID)
	if err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}

	// Register user with P2P network
	if err := sm.userManager.RegisterUser(authUser, libp2pPeerID, publicKey); err != nil {
		return fmt.Errorf("failed to register user with P2P: %w", err)
	}

	return nil
}

// UnregisterUserFromP2P removes a user from the P2P network
func (sm *ServiceManager) UnregisterUserFromP2P(userID string) error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.initialized {
		return fmt.Errorf("service manager not initialized")
	}

	// Unregister user from P2P network
	if err := sm.userManager.UnregisterUser(userID); err != nil {
		return fmt.Errorf("failed to unregister user from P2P: %w", err)
	}

	return nil
}

// GetUserByPeerID gets user information by peer ID
func (sm *ServiceManager) GetUserByPeerID(peerID string) (*user.UserInfo, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.initialized {
		return nil, fmt.Errorf("service manager not initialized")
	}

	// Convert string peerID to libp2p peer.ID
	libp2pPeerID, err := peer.IDFromString(peerID)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID: %w", err)
	}

	// Get user info
	userInfo, err := sm.userManager.GetUserByPeerID(libp2pPeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by peer ID: %w", err)
	}

	return userInfo, nil
}

// GetPeerIDByUserID gets peer ID by user ID
func (sm *ServiceManager) GetPeerIDByUserID(userID string) (string, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.initialized {
		return "", fmt.Errorf("service manager not initialized")
	}

	// Get peer ID
	peerID, err := sm.userManager.GetPeerIDByUserID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get peer ID by user ID: %w", err)
	}

	return peerID.String(), nil
}

// IsInitialized returns whether the service manager is initialized
func (sm *ServiceManager) IsInitialized() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.initialized
}

// GetServiceStatus returns the status of all services
func (sm *ServiceManager) GetServiceStatus() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	status := map[string]interface{}{
		"initialized": sm.initialized,
		"services": map[string]bool{
			"auth":    sm.authService != nil,
			"user":    sm.userService != nil,
			"user_manager": sm.userManager != nil,
			"e2ee":    sm.e2eeService != nil,
			"media":   sm.mediaService != nil,
			"calls":   sm.callService != nil,
		},
		"p2p_host": sm.host != nil,
	}

	if sm.userManager != nil {
		status["registered_users"] = len(sm.userManager.GetAllOnlineUsers())
	}

	return status
}
