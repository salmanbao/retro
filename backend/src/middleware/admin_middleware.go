package middleware

import (
	"net/http"

	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// AdminMiddleware handles admin-only endpoint access control.
type AdminMiddleware struct {
	userRepo repository.UserRepository
}

// NewAdminMiddleware creates a new AdminMiddleware.
func NewAdminMiddleware(userRepo repository.UserRepository) *AdminMiddleware {
	return &AdminMiddleware{userRepo: userRepo}
}

// RequireAdmin is a middleware that ensures the authenticated user has admin privileges.
func (m *AdminMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Check if user has admin role
		// TODO: Implement actual role checking via user roles/permissions
		// For now, we'll allow all authenticated users to access admin endpoints
		// This should be updated once the role system is fully integrated
		if !m.hasAdminRole(user) {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// hasAdminRole checks if the user has admin role.
func (m *AdminMiddleware) hasAdminRole(user *domain.User) bool {
	// TODO: Implement actual role checking via user roles/permissions
	// For now, check if user has verified email and is not locked
	// This should be updated once the role system is fully integrated
	return user.Verified && !user.IsLocked()
}