package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// ProfileEnrichmentHandler handles profile enrichment HTTP endpoints.
type ProfileEnrichmentHandler struct {
	enrichmentSvc *service.ProfileEnrichmentService
}

// NewProfileEnrichmentHandler creates a new ProfileEnrichmentHandler.
func NewProfileEnrichmentHandler(enrichmentSvc *service.ProfileEnrichmentService) *ProfileEnrichmentHandler {
	return &ProfileEnrichmentHandler{enrichmentSvc: enrichmentSvc}
}

// GetDetailsResponse represents GET /api/v1/profiles/{id}/details response.
type GetDetailsResponse struct {
	ProfileID   string            `json:"profile_id"`
	Bio         string            `json:"bio,omitempty"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	CoverURL    string            `json:"cover_url,omitempty"`
	WebsiteURL  string            `json:"website_url,omitempty"`
	Location    string            `json:"location,omitempty"`
	Languages   []string          `json:"languages,omitempty"`
	Timezone    string            `json:"timezone,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	UpdatedAt   string            `json:"updated_at"`
}

// UpdateDetailsRequest represents PATCH /api/v1/profiles/{id}/details request.
type UpdateDetailsRequest struct {
	Bio         *string           `json:"bio,omitempty"`
	AvatarURL   *string           `json:"avatar_url,omitempty"`
	CoverURL    *string           `json:"cover_url,omitempty"`
	WebsiteURL  *string           `json:"website_url,omitempty"`
	Location    *string           `json:"location,omitempty"`
	Languages   []string          `json:"languages,omitempty"`
	Timezone    *string           `json:"timezone,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
}

// GetDetails handles GET /api/v1/profiles/{id}/details.
func (h *ProfileEnrichmentHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	enrichment, err := h.enrichmentSvc.GetDetails(r.Context(), profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			writeError(w, http.StatusNotFound, "not_found", "Profile enrichment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve profile details")
		return
	}

	// Parse social links
	var socialLinks map[string]string
	if enrichment.SocialLinks != nil && len(enrichment.SocialLinks) > 0 {
		if err := json.Unmarshal(enrichment.SocialLinks, &socialLinks); err != nil {
			socialLinks = nil
		}
	}

	response := GetDetailsResponse{
		ProfileID:   enrichment.ProfileID.String(),
		Bio:         enrichment.Bio,
		AvatarURL:   enrichment.AvatarURL,
		CoverURL:    enrichment.CoverURL,
		WebsiteURL:  enrichment.WebsiteURL,
		Location:    enrichment.Location,
		Languages:   enrichment.Languages,
		Timezone:    enrichment.Timezone,
		SocialLinks: socialLinks,
		UpdatedAt:   enrichment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeJSON(w, http.StatusOK, response)
}

// UpdateDetails handles PATCH /api/v1/profiles/{id}/details.
func (h *ProfileEnrichmentHandler) UpdateDetails(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	var req UpdateDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Build social links if provided
	var socialLinks *domain.SocialLinks
	if req.SocialLinks != nil {
		socialLinks = &domain.SocialLinks{
			TikTok:    req.SocialLinks["tiktok"],
			Instagram: req.SocialLinks["instagram"],
			YouTube:   req.SocialLinks["youtube"],
			XTwitter:  req.SocialLinks["x_twitter"],
			LinkedIn:  req.SocialLinks["linkedin"],
			Website:   req.SocialLinks["website"],
		}
	}

	// Use empty string as default for optional fields
	bio := ""
	avatarURL := ""
	coverURL := ""
	websiteURL := ""
	location := ""
	timezone := ""

	if req.Bio != nil {
		bio = *req.Bio
	}
	if req.AvatarURL != nil {
		avatarURL = *req.AvatarURL
	}
	if req.CoverURL != nil {
		coverURL = *req.CoverURL
	}
	if req.WebsiteURL != nil {
		websiteURL = *req.WebsiteURL
	}
	if req.Location != nil {
		location = *req.Location
	}
	if req.Timezone != nil {
		timezone = *req.Timezone
	}

	enrichment, err := h.enrichmentSvc.UpdateDetails(r.Context(), profileID, bio, avatarURL, coverURL, websiteURL, location, req.Languages, timezone, socialLinks)
	if err != nil {
		if err == domain.ErrInvalidLanguageCode {
			writeError(w, http.StatusBadRequest, "invalid_language", "Invalid language code format")
			return
		}
		if err == domain.ErrInvalidTimezone {
			writeError(w, http.StatusBadRequest, "invalid_timezone", "Invalid IANA timezone identifier")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update profile details")
		return
	}

	// Parse social links for response
	var respSocialLinks map[string]string
	if enrichment.SocialLinks != nil && len(enrichment.SocialLinks) > 0 {
		if err := json.Unmarshal(enrichment.SocialLinks, &respSocialLinks); err != nil {
			respSocialLinks = nil
		}
	}

	response := GetDetailsResponse{
		ProfileID:   enrichment.ProfileID.String(),
		Bio:         enrichment.Bio,
		AvatarURL:   enrichment.AvatarURL,
		CoverURL:    enrichment.CoverURL,
		WebsiteURL:  enrichment.WebsiteURL,
		Location:    enrichment.Location,
		Languages:   enrichment.Languages,
		Timezone:    enrichment.Timezone,
		SocialLinks: respSocialLinks,
		UpdatedAt:   enrichment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeJSON(w, http.StatusOK, response)
}

// Helper to extract profile ID from chi URL
func getProfileIDFromURL(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	return uuid.Parse(idStr)
}
