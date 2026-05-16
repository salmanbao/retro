package handler

import (
	"encoding/json"
	"net/http"

	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// AudienceDataResponse represents GET /api/v1/profiles/{id}/audience response.
type AudienceDataResponse struct {
	ProfileID            string                 `json:"profile_id"`
	PlatformHandles      map[string]string      `json:"platform_handles,omitempty"`
	ClaimedFollowers     map[string]int         `json:"claimed_followers,omitempty"`
	EngagementRate       float64                `json:"engagement_rate,omitempty"`
	AudienceDemographics map[string]interface{} `json:"audience_demographics,omitempty"`
	UpdatedAt            string                 `json:"updated_at"`
}

// UpdateAudienceRequest represents PUT /api/v1/profiles/{id}/audience request.
type UpdateAudienceRequest struct {
	PlatformHandles  map[string]string      `json:"platform_handles,omitempty"`
	ClaimedFollowers map[string]int         `json:"claimed_followers,omitempty"`
	EngagementRate   *float64               `json:"engagement_rate,omitempty"`
	Demographics     map[string]interface{} `json:"demographics,omitempty"`
}

// AudienceHandler handles audience data HTTP endpoints.
type AudienceHandler struct {
	audienceSvc *service.AudienceService
}

// NewAudienceHandler creates a new AudienceHandler.
func NewAudienceHandler(audienceSvc *service.AudienceService) *AudienceHandler {
	return &AudienceHandler{audienceSvc: audienceSvc}
}

// GetAudience handles GET /api/v1/profiles/{id}/audience.
func (h *AudienceHandler) GetAudience(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	data, err := h.audienceSvc.GetAudience(r.Context(), profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			writeError(w, http.StatusNotFound, "not_found", "Audience data not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve audience data")
		return
	}

	response := toAudienceDataResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// UpdateAudience handles PUT /api/v1/profiles/{id}/audience.
func (h *AudienceHandler) UpdateAudience(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	var req UpdateAudienceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Validate engagement rate if provided
	var engagementRate float64
	if req.EngagementRate != nil {
		engagementRate = *req.EngagementRate
		if engagementRate < 0 || engagementRate > 100 {
			writeError(w, http.StatusBadRequest, "validation_error", "Engagement rate must be between 0 and 100")
			return
		}
	}

	// Serialize demographics
	var demographics json.RawMessage
	if req.Demographics != nil {
		data, err := json.Marshal(req.Demographics)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "Invalid demographics format")
			return
		}
		if len(data) > 10*1024 {
			writeError(w, http.StatusBadRequest, "validation_error", "Audience demographics exceed maximum size of 10KB")
			return
		}
		demographics = data
	}

	data, err := h.audienceSvc.UpdateAudience(r.Context(), profileID, req.PlatformHandles, req.ClaimedFollowers, engagementRate, demographics)
	if err != nil {
		if err == service.ErrProfileNotInfluencer {
			writeError(w, http.StatusForbidden, "forbidden", "Audience data is only available for Influencer profiles")
			return
		}
		if err == service.ErrDemographicsTooLarge {
			writeError(w, http.StatusBadRequest, "validation_error", "Audience demographics exceed maximum size of 10KB")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update audience data")
		return
	}

	response := toAudienceDataResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// toAudienceDataResponse converts a domain AudienceData to a response struct.
func toAudienceDataResponse(data *domain.AudienceData) AudienceDataResponse {
	var platformHandles map[string]string
	var claimedFollowers map[string]int
	var demographics map[string]interface{}

	if data.PlatformHandles != nil && len(data.PlatformHandles) > 0 {
		json.Unmarshal(data.PlatformHandles, &platformHandles)
	}
	if data.ClaimedFollowers != nil && len(data.ClaimedFollowers) > 0 {
		json.Unmarshal(data.ClaimedFollowers, &claimedFollowers)
	}
	if data.AudienceDemographics != nil && len(data.AudienceDemographics) > 0 {
		json.Unmarshal(data.AudienceDemographics, &demographics)
	}

	return AudienceDataResponse{
		ProfileID:            data.ProfileID.String(),
		PlatformHandles:      platformHandles,
		ClaimedFollowers:     claimedFollowers,
		EngagementRate:       data.EngagementRate,
		AudienceDemographics: demographics,
		UpdatedAt:            data.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
