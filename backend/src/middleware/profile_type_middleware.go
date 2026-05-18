package middleware

import (
	"net/http"

	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// ProfileTypeMiddleware enforces profile type restrictions for certain operations.
type ProfileTypeMiddleware struct {
	profileRepo repository.ProfileRepository
}

// NewProfileTypeMiddleware creates a new ProfileTypeMiddleware.
func NewProfileTypeMiddleware(profileRepo repository.ProfileRepository) *ProfileTypeMiddleware {
	return &ProfileTypeMiddleware{profileRepo: profileRepo}
}

// RequireEditor is a middleware that ensures the profile has Editor type for portfolio operations.
func (m *ProfileTypeMiddleware) RequireEditor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profileID, ok := GetEnrichmentProfileID(r.Context())
		if !ok {
			http.Error(w, "Profile ID not found", http.StatusBadRequest)
			return
		}

		profile, err := m.profileRepo.ByID(r.Context(), profileID)
		if err != nil {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}

		if !m.hasEditorType(profile) {
			http.Error(w, "Editor profile required for this operation", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireInfluencer is a middleware that ensures the profile has Influencer type for audience/verification operations.
func (m *ProfileTypeMiddleware) RequireInfluencer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profileID, ok := GetEnrichmentProfileID(r.Context())
		if !ok {
			http.Error(w, "Profile ID not found", http.StatusBadRequest)
			return
		}

		profile, err := m.profileRepo.ByID(r.Context(), profileID)
		if err != nil {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}

		if !m.hasInfluencerType(profile) {
			http.Error(w, "Influencer profile required for this operation", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// hasEditorType checks if the profile has Editor type.
func (m *ProfileTypeMiddleware) hasEditorType(profile *domain.Profile) bool {
	// Check if profile has Editor type using the Type field
	return profile.Type == domain.ProfileTypeEditor
}

// hasInfluencerType checks if the profile has Influencer type.
func (m *ProfileTypeMiddleware) hasInfluencerType(profile *domain.Profile) bool {
	// Check if profile has Influencer type using the Type field
	return profile.Type == domain.ProfileTypeInfluencer
}
