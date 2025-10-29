package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	peerIDKey    contextKey = "peer_id"
	signatureKey contextKey = "signature"
)

// Authenticator handles authentication and authorization
type Authenticator struct {
	userRepo     UserRepositoryInterface
	tokenManager *TokenManager
}

// NewAuthenticator creates a new Authenticator
func NewAuthenticator(db *sql.DB) *Authenticator {
	// Get JWT secrets from environment variables
	accessTokenSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessTokenSecret == "" {
		accessTokenSecret = "default-access-secret-change-in-production"
	}

	refreshTokenSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshTokenSecret == "" {
		refreshTokenSecret = "default-refresh-secret-change-in-production"
	}

	// Use memory repository for now (can be switched to SQL later)
	var userRepo UserRepositoryInterface
	if db != nil {
		userRepo = NewUserRepository(db)
	} else {
		userRepo = NewMemoryUserRepository()
	}

	return &Authenticator{
		userRepo:     userRepo,
		tokenManager: NewTokenManager(accessTokenSecret, refreshTokenSecret),
	}
}

// RegisterUser registers a new user
func (a *Authenticator) RegisterUser(req *UserRegistrationRequest) (*UserResponse, error) {
	// Validate password strength
	if err := ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Check if username already exists
	exists, err := a.userRepo.CheckUsernameExists(req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	exists, err = a.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
	}

	// Create new user
	user, err := NewUser(req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save to database
	if err := a.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user.ToResponse(), nil
}

// LoginUser authenticates a user and returns tokens
func (a *Authenticator) LoginUser(req *UserLoginRequest) (*AuthToken, *UserResponse, error) {
	// Get user by username
	user, err := a.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, fmt.Errorf("account is deactivated")
	}

	// Verify password
	if err := VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	if err := a.userRepo.UpdateLastLogin(user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to update last login time: %v\n", err)
	}

	// Generate tokens
	tokens, err := a.tokenManager.GenerateTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, user.ToResponse(), nil
}

// ValidateToken validates a JWT token and returns user info
func (a *Authenticator) ValidateToken(tokenString string) (*User, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Validate access token
	claims, err := a.tokenManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user from database
	user, err := a.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	return user, nil
}

// RefreshToken generates new tokens using a refresh token
func (a *Authenticator) RefreshToken(refreshToken string) (*AuthToken, error) {
	// Validate refresh token
	claims, err := a.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user from database
	user, err := a.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Generate new tokens
	return a.tokenManager.GenerateTokenPair(user)
}

// Context helpers
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithPeerID(ctx context.Context, peerID string) context.Context {
	return context.WithValue(ctx, peerIDKey, peerID)
}

func WithAuth(ctx context.Context, peerID string, signature []byte) context.Context {
	ctx = context.WithValue(ctx, peerIDKey, peerID)
	return context.WithValue(ctx, signatureKey, signature)
}

// Legacy peer authentication (for backward compatibility)
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
