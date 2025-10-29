package auth

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	CreateUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	UpdateLastLogin(userID string) error
	CheckUsernameExists(username string) (bool, error)
	CheckEmailExists(email string) (bool, error)
	SearchUsers(query string, limit int) ([]*User, error)
}
