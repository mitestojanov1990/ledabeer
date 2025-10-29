package auth

import (
	"errors"
	"fmt"
	"strings"
)

// UserService handles user-related operations for chat system
type UserService struct {
	repo UserRepositoryInterface
}

// NewUserService creates a new UserService
func NewUserService(repo UserRepositoryInterface) *UserService {
	return &UserService{
		repo: repo,
	}
}

// SearchUsersRequest represents a user search request
type SearchUsersRequest struct {
	Query string `json:"query" validate:"required,min=1,max=100"`
	Limit int    `json:"limit" validate:"min=1,max=50"`
}

// SearchUsersResponse represents the response for user search
type SearchUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

// UserSearchResult represents a user in search results
type UserSearchResult struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	IsOnline    bool   `json:"is_online"` // This would come from peer status
}

// SearchUsers searches for users by username, email, or display name
func (us *UserService) SearchUsers(req *SearchUsersRequest) (*SearchUsersResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20 // Default limit
	}
	if req.Limit > 50 {
		req.Limit = 50 // Max limit
	}

	// Clean and validate query
	query := strings.TrimSpace(req.Query)
	if len(query) < 1 {
		return &SearchUsersResponse{Users: []UserResponse{}, Total: 0}, nil
	}

	// Search by username, email, or display name
	users, err := us.repo.SearchUsers(query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *user.ToResponse()
	}

	return &SearchUsersResponse{
		Users: userResponses,
		Total: len(userResponses),
	}, nil
}

// GetUserByID gets a user by ID
func (us *UserService) GetUserByID(userID string) (*UserResponse, error) {
	user, err := us.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	response := user.ToResponse()
	return response, nil
}

// GetUserByUsername gets a user by username
func (us *UserService) GetUserByUsername(username string) (*UserResponse, error) {
	user, err := us.repo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	response := user.ToResponse()
	return response, nil
}

// GetUserByEmail gets a user by email
func (us *UserService) GetUserByEmail(email string) (*UserResponse, error) {
	user, err := us.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	response := user.ToResponse()
	return response, nil
}

// ValidateUserExists validates if a user exists for chat creation
func (us *UserService) ValidateUserExists(userID string) error {
	user, err := us.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}
	if !user.IsActive {
		return errors.New("user account is inactive")
	}
	return nil
}

// GetUserDisplayInfo gets minimal user info for chat display
func (us *UserService) GetUserDisplayInfo(userID string) (*UserDisplayInfo, error) {
	user, err := us.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &UserDisplayInfo{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
	}, nil
}

// UserDisplayInfo represents minimal user info for chat display
type UserDisplayInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}
