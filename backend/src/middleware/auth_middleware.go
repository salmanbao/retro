package middleware

import (
	"context"
	"net/http"
	"strings"

	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// contextKey is a type for context keys.
type contextKey string

const (
	// UserContextKey is the context key for the authenticated user.
	UserContextKey contextKey = "user"
	// SessionContextKey is the context key for the current session.
	SessionContextKey contextKey = "session"
)

// AuthMiddleware handles session-based authentication.
type AuthMiddleware struct {
	sessionRepo repository.SessionRepository
	userRepo    repository.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(sessionRepo repository.SessionRepository, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

// Authenticate is a middleware function that authenticates requests.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenHash := hashToken(token)
		session, err := m.sessionRepo.ByTokenHash(r.Context(), tokenHash)
		if err != nil {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			return
		}

		if session.IsExpired() {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		user, err := m.userRepo.ByID(r.Context(), session.UserID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		ctx = context.WithValue(ctx, UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractBearerToken extracts the bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// hashToken creates a SHA256 hash of the token for storage.
func hashToken(token string) string {
	// TODO: Use crypto/sha256 for token hashing in production
	return token
}

// GetUserFromContext retrieves the user from the context.
func GetUserFromContext(ctx context.Context) *domain.User {
	if user, ok := ctx.Value(UserContextKey).(*domain.User); ok {
		return user
	}
	return nil
}

// GetSessionFromContext retrieves the session from the context.
func GetSessionFromContext(ctx context.Context) *domain.Session {
	if session, ok := ctx.Value(SessionContextKey).(*domain.Session); ok {
		return session
	}
	return nil
}
