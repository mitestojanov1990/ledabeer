package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"ledabeer/backend/internal/auth"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	tokenManager *auth.TokenManager
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(tokenManager *auth.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenManager: tokenManager,
	}
}

// Authenticate is middleware that validates JWT tokens
func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := am.extractToken(r)
		if err != nil {
			am.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or missing token")
			return
		}

		// Validate token
		claims, err := am.tokenManager.ValidateAccessToken(token)
		if err != nil {
			am.writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "claims", claims)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts JWT token from Authorization header
func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// writeErrorResponse writes a JSON error response
func (am *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

// GetUsernameFromContext extracts username from request context
func GetUsernameFromContext(ctx context.Context) (string, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return "", errors.New("username not found in context")
	}
	return username, nil
}

// GetEmailFromContext extracts email from request context
func GetEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value("email").(string)
	if !ok {
		return "", errors.New("email not found in context")
	}
	return email, nil
}

// GetClaimsFromContext extracts JWT claims from request context
func GetClaimsFromContext(ctx context.Context) (*auth.JWTClaims, error) {
	claims, ok := ctx.Value("claims").(*auth.JWTClaims)
	if !ok {
		return nil, errors.New("claims not found in context")
	}
	return claims, nil
}
