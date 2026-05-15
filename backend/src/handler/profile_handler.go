package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// ProfileHandler handles profile HTTP endpoints.
type ProfileHandler struct {
	profileSvc *service.ProfileService
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(profileSvc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc}
}

// CreateProfileRequest represents the profile creation request body.
type CreateProfileRequest struct {
	ProfileType string                 `json:"profile_type"`
	Name        string                 `json:"name"`
	Details     map[string]interface{} `json:"details"`
}

// CreateProfileResponse represents a successful profile creation response.
type CreateProfileResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"profile_type"`
	Name      string                 `json:"name"`
	Details   map[string]interface{} `json:"details"`
	CreatedAt string                 `json:"created_at"`
}

// ProfileResponse represents a profile in API responses.
type ProfileResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"profile_type"`
	Name      string                 `json:"name"`
	Details   map[string]interface{} `json:"details"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// parseProfileType converts string to ProfileType.
func parseProfileType(s string) (domain.ProfileType, error) {
	switch s {
	case "brand":
		return domain.ProfileTypeBrand, nil
	case "editor":
		return domain.ProfileTypeEditor, nil
	case "influencer":
		return domain.ProfileTypeInfluencer, nil
	default:
		return "", domain.ErrInvalidProfileType
	}
}

// Create handles POST /api/v1/profiles.
func (h *ProfileHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Validate name
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	// Parse and validate profile type
	profileType, err := parseProfileType(req.ProfileType)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_profile_type", "Profile type must be brand, editor, or influencer")
		return
	}

	// Validate details presence
	if req.Details == nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Details are required")
		return
	}

	// Create service request
	serviceReq := &service.CreateProfileRequest{
		Type:    profileType,
		Name:    req.Name,
		Details: req.Details,
	}

	profile, err := h.profileSvc.CreateProfile(r.Context(), user.ID, serviceReq)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidProfileType) {
			writeError(w, http.StatusBadRequest, "invalid_profile_type", "Invalid profile type")
		} else if errors.Is(err, domain.ErrProfileNotOwned) {
			writeError(w, http.StatusBadRequest, "profile_error", "Profile already exists or invalid")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		}
		return
	}

	// Parse details JSON back to map
	var details map[string]interface{}
	json.Unmarshal(profile.Details, &details)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateProfileResponse{
		ID:        profile.ID.String(),
		Type:      string(profile.Type),
		Name:      profile.Name,
		Details:   details,
		CreatedAt: profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// List handles GET /api/v1/profiles.
func (h *ProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	profiles, err := h.profileSvc.ListProfiles(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	// Convert to response format
	var response []ProfileResponse
	for _, p := range profiles {
		var details map[string]interface{}
		json.Unmarshal(p.Details, &details)
		response = append(response, ProfileResponse{
			ID:        p.ID.String(),
			Type:      string(p.Type),
			Name:      p.Name,
			Details:   details,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if response == nil {
		response = []ProfileResponse{}
	}
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers profile routes on the router.
func (h *ProfileHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

// UpdateProfileRequest represents the profile update request body.
type UpdateProfileRequest struct {
	Name    string                 `json:"name"`
	Details map[string]interface{} `json:"details"`
}

// Get handles GET /api/v1/profiles/{id}.
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	profileID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_profile_id", "Profile ID must be a valid UUID")
		return
	}

	profile, err := h.profileSvc.GetProfile(r.Context(), user.ID, profileID)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotOwned) {
			writeError(w, http.StatusForbidden, "forbidden", "Profile does not belong to you")
			return
		}
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	var details map[string]interface{}
	json.Unmarshal(profile.Details, &details)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ProfileResponse{
		ID:        profile.ID.String(),
		Type:      string(profile.Type),
		Name:      profile.Name,
		Details:   details,
		CreatedAt: profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Update handles PATCH /api/v1/profiles/{id}.
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	profileID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_profile_id", "Profile ID must be a valid UUID")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	serviceReq := &service.UpdateProfileRequest{
		Name:    req.Name,
		Details: req.Details,
	}

	profile, err := h.profileSvc.UpdateProfile(r.Context(), user.ID, profileID, serviceReq)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotOwned) {
			writeError(w, http.StatusForbidden, "forbidden", "Profile does not belong to you")
			return
		}
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	var details map[string]interface{}
	json.Unmarshal(profile.Details, &details)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ProfileResponse{
		ID:        profile.ID.String(),
		Type:      string(profile.Type),
		Name:      profile.Name,
		Details:   details,
		CreatedAt: profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Delete handles DELETE /api/v1/profiles/{id}.
func (h *ProfileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	profileID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_profile_id", "Profile ID must be a valid UUID")
		return
	}

	err = h.profileSvc.DeleteProfile(r.Context(), user.ID, profileID)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotOwned) {
			writeError(w, http.StatusForbidden, "forbidden", "Profile does not belong to you")
			return
		}
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
