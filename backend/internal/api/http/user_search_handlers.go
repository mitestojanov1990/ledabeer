package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ledabeer/backend/internal/user"
)

// UserSearchHandlers handles user search HTTP endpoints
type UserSearchHandlers struct {
	userManager *user.UserManager
}

// NewUserSearchHandlers creates a new UserSearchHandlers instance
func NewUserSearchHandlers(userManager *user.UserManager) *UserSearchHandlers {
	return &UserSearchHandlers{
		userManager: userManager,
	}
}

// SearchUsersRequest represents a user search request
type SearchUsersRequest struct {
	Query string `json:"query" validate:"required,min=1,max=100"`
	Limit int    `json:"limit" validate:"min=1,max=50"`
}

// SearchUsersResponse represents the response for user search
type SearchUsersResponse struct {
	Users []UserSearchResult `json:"users"`
	Total int                `json:"total"`
}

// UserSearchResult represents a user in search results
type UserSearchResult struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	IsOnline    bool   `json:"is_online"`
}

// SearchUsers handles user search requests
func (h *UserSearchHandlers) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchUsersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Basic validation
	if req.Query == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Query is required")
		return
	}

	// Set default limit
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	// Search users
	searchResults := h.userManager.SearchUsers(req.Query)

	// Convert to response format
	users := make([]UserSearchResult, 0, len(searchResults))
	for _, userInfo := range searchResults {
		// Limit results
		if len(users) >= req.Limit {
			break
		}

		users = append(users, UserSearchResult{
			UserID:      userInfo.UserID,
			Username:    userInfo.Username,
			DisplayName: userInfo.DisplayName,
			Email:       userInfo.Email,
			AvatarURL:   userInfo.AvatarURL,
			IsOnline:    userInfo.IsOnline,
		})
	}

	response := SearchUsersResponse{
		Users: users,
		Total: len(users),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// FindUserByEmail handles finding a user by email
func (h *UserSearchHandlers) FindUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Search for user by email
	searchResults := h.userManager.SearchUsers(email)
	
	// Find exact email match
	var foundUser *user.UserInfo
	for _, userInfo := range searchResults {
		if strings.EqualFold(userInfo.Email, email) {
			foundUser = userInfo
			break
		}
	}

	if foundUser == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	userResult := UserSearchResult{
		UserID:      foundUser.UserID,
		Username:    foundUser.Username,
		DisplayName: foundUser.DisplayName,
		Email:       foundUser.Email,
		AvatarURL:   foundUser.AvatarURL,
		IsOnline:    foundUser.IsOnline,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResult)
}

// FindUserByUsername handles finding a user by username
func (h *UserSearchHandlers) FindUserByUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Username is required")
		return
	}

	// Search for user by username
	searchResults := h.userManager.SearchUsers(username)
	
	// Find exact username match
	var foundUser *user.UserInfo
	for _, userInfo := range searchResults {
		if strings.EqualFold(userInfo.Username, username) {
			foundUser = userInfo
			break
		}
	}

	if foundUser == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	userResult := UserSearchResult{
		UserID:      foundUser.UserID,
		Username:    foundUser.Username,
		DisplayName: foundUser.DisplayName,
		Email:       foundUser.Email,
		AvatarURL:   foundUser.AvatarURL,
		IsOnline:    foundUser.IsOnline,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResult)
}

// writeErrorResponse writes a JSON error response
func (h *UserSearchHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}
