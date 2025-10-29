package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return errors.New("hashed password cannot be empty")
	}
	if password == "" {
		return errors.New("password cannot be empty")
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePasswordStrength validates password strength
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Add more validation rules as needed
	// - Must contain at least one uppercase letter
	// - Must contain at least one lowercase letter
	// - Must contain at least one number
	// - Must contain at least one special character

	return nil
}
