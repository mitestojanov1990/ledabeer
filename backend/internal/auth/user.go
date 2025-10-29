package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the system
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never expose in JSON
	DisplayName  string     `json:"display_name" db:"display_name"`
	AvatarURL    string     `json:"avatar_url" db:"avatar_url"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
}

// UserRegistrationRequest represents the data needed to register a new user
type UserRegistrationRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=50"`
}

// UserLoginRequest represents the data needed to log in a user
type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the user data returned to the client (without sensitive info)
type UserResponse struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

// AuthToken represents a JWT token response
type AuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ToResponse converts a User to UserResponse (removes sensitive data)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

// NewUser creates a new User instance
func NewUser(username, email, password, displayName string) (*User, error) {
	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		DisplayName:  displayName,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
