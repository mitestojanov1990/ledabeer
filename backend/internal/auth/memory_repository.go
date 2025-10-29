package auth

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryUserRepository is an in-memory implementation of UserRepository for testing
type MemoryUserRepository struct {
	users     map[string]*User
	emails    map[string]string // email -> userID
	usernames map[string]string // username -> userID
	mutex     sync.RWMutex
}

// NewMemoryUserRepository creates a new in-memory user repository
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:     make(map[string]*User),
		emails:    make(map[string]string),
		usernames: make(map[string]string),
	}
}

// CreateUser creates a new user in memory
func (r *MemoryUserRepository) CreateUser(user *User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if username already exists
	if _, exists := r.usernames[user.Username]; exists {
		return fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if _, exists := r.emails[user.Email]; exists {
		return fmt.Errorf("email already exists")
	}

	// Store user
	r.users[user.ID] = user
	r.emails[user.Email] = user.ID
	r.usernames[user.Username] = user.ID

	return nil
}

// GetUserByID retrieves a user by ID
func (r *MemoryUserRepository) GetUserByID(id string) (*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Return a copy to avoid race conditions
	userCopy := *user
	return &userCopy, nil
}

// GetUserByUsername retrieves a user by username
func (r *MemoryUserRepository) GetUserByUsername(username string) (*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userID, exists := r.usernames[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Return a copy to avoid race conditions
	userCopy := *user
	return &userCopy, nil
}

// GetUserByEmail retrieves a user by email
func (r *MemoryUserRepository) GetUserByEmail(email string) (*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userID, exists := r.emails[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Return a copy to avoid race conditions
	userCopy := *user
	return &userCopy, nil
}

// UpdateUser updates a user in memory
func (r *MemoryUserRepository) UpdateUser(user *User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user exists
	_, exists := r.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Update user
	r.users[user.ID] = user
	return nil
}

// UpdateLastLogin updates the last login time for a user
func (r *MemoryUserRepository) UpdateLastLogin(userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	return nil
}

// CheckUsernameExists checks if a username already exists
func (r *MemoryUserRepository) CheckUsernameExists(username string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.usernames[username]
	return exists, nil
}

// CheckEmailExists checks if an email already exists
func (r *MemoryUserRepository) CheckEmailExists(email string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.emails[email]
	return exists, nil
}

// SearchUsers searches for users by username, email, or display name
func (r *MemoryUserRepository) SearchUsers(query string, limit int) ([]*User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []*User{}, nil
	}

	var results []*User
	count := 0

	// Search through all users
	for _, user := range r.users {
		if count >= limit {
			break
		}

		// Check if query matches username, email, or display name
		usernameMatch := strings.Contains(strings.ToLower(user.Username), query)
		emailMatch := strings.Contains(strings.ToLower(user.Email), query)
		displayNameMatch := strings.Contains(strings.ToLower(user.DisplayName), query)

		if usernameMatch || emailMatch || displayNameMatch {
			// Create a copy to avoid race conditions
			userCopy := *user
			results = append(results, &userCopy)
			count++
		}
	}

	return results, nil
}
