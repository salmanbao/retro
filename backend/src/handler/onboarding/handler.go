package onboarding

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domain "viralforge/backend/src/domain/onboarding"
	"viralforge/backend/src/middleware"
	onboardingSvc "viralforge/backend/src/service/onboarding"
)

// Handler handles onboarding HTTP endpoints.
type Handler struct {
	svc           *onboardingSvc.Service
	activationSvc *onboardingSvc.ActivationService
}

// NewHandler creates a new OnboardingHandler.
func NewHandler(svc *onboardingSvc.Service, activationSvc *onboardingSvc.ActivationService) *Handler {
	return &Handler{
		svc:           svc,
		activationSvc: activationSvc,
	}
}

// RegisterPublicRoutes registers public onboarding routes (requires authentication but not admin).
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/{id}/onboarding", h.GetOnboarding)
	r.Get("/{id}/onboarding/steps", h.GetSteps)
	r.Patch("/{id}/onboarding/steps/{stepId}", h.UpdateStep)
	r.Post("/{id}/onboarding/recalculate", h.Recalculate)
	r.Get("/{id}/onboarding/next-step", h.GetNextStep)
}

// GetOnboardingResponse represents the response for GetOnboarding.
type GetOnboardingResponse struct {
	ProfileID              string  `json:"profile_id"`
	ActivationStatus       string  `json:"activation_status"`
	PercentageComplete    int     `json:"percentage_complete"`
	RequiredStepsRemaining int     `json:"required_steps_remaining"`
	TemplateVersion       string  `json:"template_version"`
	StartedAt             *string `json:"started_at,omitempty"`
	LastActivityAt        *string `json:"last_activity_at,omitempty"`
}

// StepResponse represents a step in API responses.
type StepResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description,omitempty"`
	ActionURL    string  `json:"action_url,omitempty"`
	StepType     string  `json:"step_type"`
	Required     bool    `json:"required"`
	DisplayOrder int     `json:"display_order"`
	Status       string  `json:"status"`
	StartedAt    *string `json:"started_at,omitempty"`
	CompletedAt  *string `json:"completed_at,omitempty"`
}

// GetStepsResponse represents the response for GetSteps.
type GetStepsResponse struct {
	Steps []StepResponse `json:"steps"`
}

// UpdateStepRequest represents the request body for UpdateStep.
type UpdateStepRequest struct {
	Status string `json:"status"`
}

// GetNextStepResponse represents the response for GetNextStep.
type GetNextStepResponse struct {
	Step    *StepResponse `json:"step"`
	Message string        `json:"message,omitempty"`
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

// formatTime formats a time pointer to string.
func formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

// GetOnboarding handles GET /api/v1/profiles/{id}/domain.
func (h *Handler) GetOnboarding(w http.ResponseWriter, r *http.Request) {
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

	progress, err := h.svc.GetProgressByProfileID(profileID)
	if err != nil {
		if errors.Is(err, domain.ErrProgressNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Onboarding progress not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	percentage, _ := h.activationSvc.CalculatePercentage(progress)

	resp := GetOnboardingResponse{
		ProfileID:           profileID.String(),
		ActivationStatus:    progress.ActivationStatus,
		PercentageComplete: percentage,
		TemplateVersion:    progress.TemplateVersion,
		StartedAt:          formatTime(progress.StartedAt),
		LastActivityAt:     formatTime(progress.LastActivityAt),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetSteps handles GET /api/v1/profiles/{id}/onboarding/steps.
func (h *Handler) GetSteps(w http.ResponseWriter, r *http.Request) {
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

	progress, err := h.svc.GetProgressByProfileID(profileID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Onboarding progress not found")
		return
	}

	template, err := h.svc.GetTemplateByProfileType(progress.ProfileType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not load template")
		return
	}

	stepProgresses, err := h.svc.GetProgressByProfileID(profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not load steps")
		return
	}

	// Create step progress map
	stepProgressMap := make(map[uuid.UUID]domain.StepProgress)
	for _, sp := range stepProgresses.StepProgresses {
		stepProgressMap[sp.StepID] = sp
	}

	stepResponses := make([]StepResponse, len(template.Steps))
	for i, step := range template.Steps {
		sp := stepProgressMap[step.ID]
		stepResponses[i] = StepResponse{
			ID:           step.ID.String(),
			Title:        step.Title,
			Description:  step.Description,
			ActionURL:    step.ActionURL,
			StepType:     step.StepType,
			Required:     step.Required,
			DisplayOrder: step.DisplayOrder,
			Status:       sp.Status,
			StartedAt:    formatTime(sp.StartedAt),
			CompletedAt:  formatTime(sp.CompletedAt),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetStepsResponse{Steps: stepResponses})
}

// UpdateStep handles PATCH /api/v1/profiles/{id}/onboarding/steps/{stepId}.
func (h *Handler) UpdateStep(w http.ResponseWriter, r *http.Request) {
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

	stepIdStr := chi.URLParam(r, "stepId")
	stepID, err := uuid.Parse(stepIdStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_step_id", "Step ID must be a valid UUID")
		return
	}

	var req UpdateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if !domain.IsValidStepStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "invalid_status", "Invalid status value")
		return
	}

	progress, err := h.svc.GetProgressByProfileID(profileID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Onboarding progress not found")
		return
	}

	updated, err := h.svc.UpdateStepStatus(progress.ID, stepID, req.Status)
	if err != nil {
		if errors.Is(err, domain.ErrStepNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Step not found")
			return
		}
		if errors.Is(err, domain.ErrStepNotSkippable) {
			writeError(w, http.StatusForbidden, "cannot_skip_required", "Required steps cannot be skipped")
			return
		}
		if errors.Is(err, domain.ErrInvalidStepStatus) {
			writeError(w, http.StatusBadRequest, "invalid_transition", "Invalid step status transition")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	resp := StepResponse{
		ID:           updated.StepID.String(),
		Title:        updated.Status, // This should be filled properly
		Status:       updated.Status,
		StartedAt:    formatTime(updated.StartedAt),
		CompletedAt:  formatTime(updated.CompletedAt),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Recalculate handles POST /api/v1/profiles/{id}/onboarding/recalculate.
func (h *Handler) Recalculate(w http.ResponseWriter, r *http.Request) {
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

	progress, err := h.svc.RecalculateProgress(profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	percentage, _ := h.activationSvc.CalculatePercentage(progress)

	resp := GetOnboardingResponse{
		ProfileID:           profileID.String(),
		ActivationStatus:    progress.ActivationStatus,
		PercentageComplete: percentage,
		TemplateVersion:    progress.TemplateVersion,
		StartedAt:          formatTime(progress.StartedAt),
		LastActivityAt:     formatTime(progress.LastActivityAt),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// RegisterAdminRoutes registers admin-only onboarding routes.
func (h *Handler) RegisterAdminRoutes(r chi.Router) {
	r.Post("/{id}/onboarding/activate", h.AdminActivate)
}

// AdminActivate handles POST /api/v1/admin/profiles/{id}/onboarding/activate.
func (h *Handler) AdminActivate(w http.ResponseWriter, r *http.Request) {
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

	progress, err := h.svc.GetProgressByProfileID(profileID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Onboarding progress not found")
		return
	}

	if err := h.activationSvc.ActivateProfile(progress); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to activate profile")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile activated successfully",
	})
}

// GetNextStep handles GET /api/v1/profiles/{id}/onboarding/next-step.
func (h *Handler) GetNextStep(w http.ResponseWriter, r *http.Request) {
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

	nextStep, err := h.svc.GetNextStep(profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	if nextStep == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GetNextStepResponse{
			Step:    nil,
			Message: "All steps completed",
		})
		return
	}

	resp := GetNextStepResponse{
		Step: &StepResponse{
			ID:           nextStep.ID.String(),
			Title:        nextStep.Title,
			Description:  nextStep.Description,
			ActionURL:    nextStep.ActionURL,
			StepType:     nextStep.StepType,
			Required:     nextStep.Required,
			DisplayOrder: nextStep.DisplayOrder,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}