package auth

import (
	"database/sql"
	"fmt"
	"time"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, is_active, created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`

	user := &User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, is_active, created_at, updated_at, last_login_at
		FROM users WHERE username = $1
	`

	user := &User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, is_active, created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`

	user := &User{}
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// UpdateUser updates a user in the database
func (r *UserRepository) UpdateUser(user *User) error {
	query := `
		UPDATE users 
		SET username = $2, email = $3, password_hash = $4, display_name = $5, avatar_url = $6, is_active = $7, updated_at = $8, last_login_at = $9
		WHERE id = $1
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.IsActive,
		user.UpdatedAt,
		user.LastLoginAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last login time for a user
func (r *UserRepository) UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`

	_, err := r.db.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// CheckUsernameExists checks if a username already exists
func (r *UserRepository) CheckUsernameExists(username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = $1`

	var count int
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username: %w", err)
	}

	return count > 0, nil
}

// CheckEmailExists checks if an email already exists
func (r *UserRepository) CheckEmailExists(email string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE email = $1`

	var count int
	err := r.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}

	return count > 0, nil
}

// SearchUsers searches for users by username, email, or display name
func (r *UserRepository) SearchUsers(query string, limit int) ([]*User, error) {
	searchQuery := `
		SELECT id, username, email, password_hash, display_name, avatar_url, is_active, created_at, updated_at, last_login_at
		FROM users 
		WHERE username ILIKE $1 OR email ILIKE $1 OR display_name ILIKE $1
		ORDER BY username
		LIMIT $2
	`
	
	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(searchQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.DisplayName,
			&user.AvatarURL,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
