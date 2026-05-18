package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/repository"
)

const (
	// EnrichmentProfileIDKey is the context key for the profile being accessed.
	EnrichmentProfileIDKey contextKey = "enrichment_profile_id"
)

// OwnershipMiddleware verifies profile ownership.
type OwnershipMiddleware struct {
	profileRepo repository.ProfileRepository
}

// NewOwnershipMiddleware creates a new OwnershipMiddleware.
func NewOwnershipMiddleware(profileRepo repository.ProfileRepository) *OwnershipMiddleware {
	return &OwnershipMiddleware{profileRepo: profileRepo}
}

// RequireOwnership middleware ensures the authenticated user owns the profile.
func (m *OwnershipMiddleware) RequireOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get profile ID from URL - chi uses URLParam
		profileIDStr := chi.URLParam(r, "id")
		if profileIDStr == "" {
			http.Error(w, "Profile ID required", http.StatusBadRequest)
			return
		}

		profileID, err := uuid.Parse(profileIDStr)
		if err != nil {
			http.Error(w, "Invalid profile ID format", http.StatusBadRequest)
			return
		}

		// Get user from context
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Check if user owns any profile that matches the requested profile ID
		profiles, err := m.profileRepo.ByUserID(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "Failed to verify ownership", http.StatusInternalServerError)
			return
		}

		owned := false
		for _, profile := range profiles {
			if profile.ID == profileID {
				owned = true
				break
			}
		}

		if !owned {
			http.Error(w, "You do not have access to this profile", http.StatusForbidden)
			return
		}

		// Store profile ID in context for later use
		ctx := context.WithValue(r.Context(), EnrichmentProfileIDKey, profileID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetEnrichmentProfileID retrieves the profile ID from context.
func GetEnrichmentProfileID(ctx context.Context) (uuid.UUID, bool) {
	if id, ok := ctx.Value(EnrichmentProfileIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}
