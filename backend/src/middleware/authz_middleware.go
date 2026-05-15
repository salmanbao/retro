package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/go-chi/chi"
	"viralforge/backend/src/service"
)

const (
	// RBACContextKey is the context key for RBAC check results.
	RBACContextKey contextKey = "rbac_result"
)

// RBACResult represents the result of an RBAC permission check.
type RBACResult struct {
	Allowed   bool
	Reason    string
	Permission string
}

// AuthzMiddleware handles authorization checks.
type AuthzMiddleware struct {
	authzSvc *service.AuthorizationService
}

// NewAuthzMiddleware creates a new AuthzMiddleware.
func NewAuthzMiddleware(authzSvc *service.AuthorizationService) *AuthzMiddleware {
	return &AuthzMiddleware{
		authzSvc: authzSvc,
	}
}

// RequirePermission returns a middleware that checks for a specific permission.
// The profile ID is extracted from the context set by AuthMiddleware.
func (m *AuthzMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get profile ID from context (set by AuthMiddleware)
			profileID := GetActiveProfileID(r.Context())
			if profileID == nil {
				writeForbidden(w, "no active profile", permission)
				return
			}

			// Check permission
			allowed, err := m.authzSvc.HasPermission(r.Context(), *profileID, permission)
			if err != nil {
				writeInternalError(w)
				return
			}

			if !allowed {
				writeForbidden(w, "permission denied", permission)
				return
			}

			// Store RBAC result in context for handlers that need it
			ctx := context.WithValue(r.Context(), RBACContextKey, RBACResult{
				Allowed:    true,
				Permission: permission,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRBACResult retrieves the RBAC result from context.
func GetRBACResult(ctx context.Context) *RBACResult {
	if result, ok := ctx.Value(RBACContextKey).(RBACResult); ok {
		return &result
	}
	return nil
}

// writeForbidden writes a 403 Forbidden response.
func writeForbidden(w http.ResponseWriter, reason, permission string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"forbidden","message":"permission denied: ` + permission + `"}`))
}

// writeInternalError writes a 500 Internal Server Error response.
func writeInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error":"internal_error","message":"internal server error"}`))
}

// GetProfileIDFromRequest extracts profile ID from chi router URL parameter.
func GetProfileIDFromRequest(r *http.Request) *uuid.UUID {
	idStr := chi.URLParam(r, "profileID")
	if idStr == "" {
		return nil
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil
	}
	return &id
}

// GetCallerProfileID extracts the caller's profile ID from context.
// This is set by the AuthMiddleware after authentication.
func GetCallerProfileID(ctx context.Context) *uuid.UUID {
	if profileID, ok := ctx.Value(ActiveProfileIDKey).(uuid.UUID); ok {
		return &profileID
	}
	return nil
}