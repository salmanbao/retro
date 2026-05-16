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

// VerificationResponse represents GET/POST /api/v1/profiles/{id}/verification response.
type VerificationResponse struct {
	ProfileID         string   `json:"profile_id"`
	Status            string   `json:"status"`
	EvidenceURLs      []string `json:"evidence_urls,omitempty"`
	VerificationNotes string   `json:"verification_notes,omitempty"`
	ReviewedAt        string   `json:"reviewed_at,omitempty"`
	ReviewedBy        string   `json:"reviewed_by,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// SubmitVerificationRequest represents POST /api/v1/profiles/{id}/verification request.
type SubmitVerificationRequest struct {
	EvidenceURLs []string `json:"evidence_urls"`
	Notes        string   `json:"notes,omitempty"`
}

// AdminVerificationRequest represents PUT /api/v1/admin/profiles/{id}/verification/review request.
type AdminVerificationRequest struct {
	Status     string `json:"status"`
	ReviewedBy string `json:"reviewed_by"`
	Notes      string `json:"notes,omitempty"`
}

// KYCResponse represents GET /api/v1/profiles/{id}/kyc response.
type KYCResponse struct {
	ProfileID   string `json:"profile_id"`
	Status      string `json:"status"`
	ReviewNotes string `json:"review_notes,omitempty"`
	ReviewedAt  string `json:"reviewed_at,omitempty"`
	ReviewedBy  string `json:"reviewed_by,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

// AdminKYCRequest represents PUT /api/v1/admin/profiles/{id}/kyc request.
type AdminKYCRequest struct {
	Status     string `json:"status"`
	ReviewedBy string `json:"reviewed_by"`
	Notes      string `json:"notes,omitempty"`
}

// VerificationHandler handles follower verification HTTP endpoints.
type VerificationHandler struct {
	verificationSvc *service.VerificationService
}

// NewVerificationHandler creates a new VerificationHandler.
func NewVerificationHandler(verificationSvc *service.VerificationService) *VerificationHandler {
	return &VerificationHandler{verificationSvc: verificationSvc}
}

// GetVerification handles GET /api/v1/profiles/{id}/verification.
func (h *VerificationHandler) GetVerification(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	data, err := h.verificationSvc.GetVerification(r.Context(), profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve verification status")
		return
	}

	response := toVerificationResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// SubmitVerification handles POST /api/v1/profiles/{id}/verification.
func (h *VerificationHandler) SubmitVerification(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	var req SubmitVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if len(req.EvidenceURLs) == 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "At least one evidence URL is required")
		return
	}

	data, err := h.verificationSvc.SubmitVerification(r.Context(), profileID, req.EvidenceURLs, req.Notes)
	if err != nil {
		if err == service.ErrProfileNotInfluencerVerification {
			writeError(w, http.StatusForbidden, "forbidden", "Verification is only available for Influencer profiles")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to submit verification")
		return
	}

	response := toVerificationResponse(data)
	writeJSON(w, http.StatusCreated, response)
}

// AdminReviewVerification handles PUT /api/v1/admin/profiles/{id}/verification/review.
func (h *VerificationHandler) AdminReviewVerification(w http.ResponseWriter, r *http.Request) {
	profileIDStr := chi.URLParam(r, "profileId")
	profileID, err := parseUUID(profileIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid profile ID format")
		return
	}

	var req AdminVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Validate status
	status := domain.VerificationStatus(req.Status)
	if !isValidVerificationStatus(status) {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid status. Must be 'verified' or 'rejected'")
		return
	}

	data, err := h.verificationSvc.AdminReviewVerification(r.Context(), profileID, status, req.ReviewedBy, req.Notes)
	if err != nil {
		if err == service.ErrInvalidVerificationStatus {
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid verification status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update verification")
		return
	}

	response := toVerificationResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// KYCHandler handles KYC status HTTP endpoints.
type KYCHandler struct {
	kycSvc *service.KYCService
}

// NewKYCHandler creates a new KYCHandler.
func NewKYCHandler(kycSvc *service.KYCService) *KYCHandler {
	return &KYCHandler{kycSvc: kycSvc}
}

// GetKYCStatus handles GET /api/v1/profiles/{id}/kyc.
func (h *KYCHandler) GetKYCStatus(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	data, err := h.kycSvc.GetKYCStatus(r.Context(), profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve KYC status")
		return
	}

	response := toKYCResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// AdminUpdateKYC handles PUT /api/v1/admin/profiles/{id}/kyc.
func (h *KYCHandler) AdminUpdateKYC(w http.ResponseWriter, r *http.Request) {
	profileIDStr := chi.URLParam(r, "profileId")
	profileID, err := parseUUID(profileIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid profile ID format")
		return
	}

	var req AdminKYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Validate status
	status := domain.KYCStatusValue(req.Status)
	if !isValidKYCStatusValue(status) {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid KYC status")
		return
	}

	data, err := h.kycSvc.AdminUpdateKYC(r.Context(), profileID, status, req.ReviewedBy, req.Notes)
	if err != nil {
		if err == service.ErrInvalidKYCStatus {
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid KYC status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update KYC status")
		return
	}

	response := toKYCResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// Helper functions

func toVerificationResponse(v *domain.FollowerVerification) VerificationResponse {
	var evidenceURLs []string
	if v.EvidenceURLs != nil && len(v.EvidenceURLs) > 0 {
		json.Unmarshal(v.EvidenceURLs, &evidenceURLs)
	}

	var reviewedAt string
	if v.ReviewedAt != nil {
		reviewedAt = v.ReviewedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return VerificationResponse{
		ProfileID:         v.ProfileID.String(),
		Status:            string(v.Status),
		EvidenceURLs:      evidenceURLs,
		VerificationNotes: v.VerificationNotes,
		ReviewedAt:        reviewedAt,
		ReviewedBy:        v.ReviewedBy,
		CreatedAt:         v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toKYCResponse(k *domain.KYCStatus) KYCResponse {
	var reviewedAt string
	if k.ReviewedAt != nil {
		reviewedAt = k.ReviewedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return KYCResponse{
		ProfileID:   k.ProfileID.String(),
		Status:      string(k.Status),
		ReviewNotes: k.ReviewNotes,
		ReviewedAt:  reviewedAt,
		ReviewedBy:  k.ReviewedBy,
		UpdatedAt:   k.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func isValidVerificationStatus(status domain.VerificationStatus) bool {
	return status == domain.VerificationStatusVerified || status == domain.VerificationStatusRejected
}

func isValidKYCStatusValue(status domain.KYCStatusValue) bool {
	switch status {
	case domain.KYCStatusNotStarted, domain.KYCStatusPending, domain.KYCStatusApproved, domain.KYCStatusRejected, domain.KYCStatusSuspended:
		return true
	default:
		return false
	}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
