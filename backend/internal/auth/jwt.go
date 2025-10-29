package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// JWT constants
	AccessTokenExpiry  = 15 * time.Minute   // 15 minutes
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
	TokenTypeBearer    = "Bearer"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT token operations
type TokenManager struct {
	accessTokenSecret  []byte
	refreshTokenSecret []byte
}

// NewTokenManager creates a new TokenManager
func NewTokenManager(accessTokenSecret, refreshTokenSecret string) *TokenManager {
	return &TokenManager{
		accessTokenSecret:  []byte(accessTokenSecret),
		refreshTokenSecret: []byte(refreshTokenSecret),
	}
}

// GenerateAccessToken generates a new access token for a user
func (tm *TokenManager) GenerateAccessToken(user *User) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ledabeer",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.accessTokenSecret)
}

// GenerateRefreshToken generates a new refresh token for a user
func (tm *TokenManager) GenerateRefreshToken(user *User) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ledabeer",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.refreshTokenSecret)
}

// ValidateAccessToken validates an access token and returns the claims
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return tm.validateToken(tokenString, tm.accessTokenSecret)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (tm *TokenManager) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return tm.validateToken(tokenString, tm.refreshTokenSecret)
}

// validateToken validates a JWT token with the given secret
func (tm *TokenManager) validateToken(tokenString string, secret []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (tm *TokenManager) GenerateTokenPair(user *User) (*AuthToken, error) {
	accessToken, err := tm.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := tm.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    TokenTypeBearer,
		ExpiresIn:    int64(AccessTokenExpiry.Seconds()),
	}, nil
}
