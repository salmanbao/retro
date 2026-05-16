package handler

import (
	"encoding/json"
	"net/http"

	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// PayoutPreferencesResponse represents GET/PUT /api/v1/profiles/{id}/payout response.
type PayoutPreferencesResponse struct {
	ProfileID       string `json:"profile_id"`
	PreferredMethod string `json:"preferred_method"`
	BeneficiaryName string `json:"beneficiary_name"`
	Country         string `json:"country"`
	Currency        string `json:"currency"`
	PayoutReady     bool   `json:"payout_ready"`
	UpdatedAt       string `json:"updated_at"`
}

// UpdatePayoutPreferencesRequest represents PUT /api/v1/profiles/{id}/payout request.
type UpdatePayoutPreferencesRequest struct {
	PreferredMethod  string `json:"preferred_method"`
	BeneficiaryName  string `json:"beneficiary_name"`
	Country          string `json:"country"`
	Currency         string `json:"currency"`
	EncryptedDetails string `json:"encrypted_details,omitempty"`
	PayoutReady      *bool  `json:"payout_ready,omitempty"`
}

// PayoutHandler handles payout preferences HTTP endpoints.
type PayoutHandler struct {
	payoutSvc *service.PayoutService
}

// NewPayoutHandler creates a new PayoutHandler.
func NewPayoutHandler(payoutSvc *service.PayoutService) *PayoutHandler {
	return &PayoutHandler{payoutSvc: payoutSvc}
}

// GetPayoutPreferences handles GET /api/v1/profiles/{id}/payout.
// Returns masked payout preferences - encrypted_details is never included.
func (h *PayoutHandler) GetPayoutPreferences(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	data, err := h.payoutSvc.GetPayoutPreferences(r.Context(), profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve payout preferences")
		return
	}

	response := toPayoutPreferencesResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// UpdatePayoutPreferences handles PUT /api/v1/profiles/{id}/payout.
func (h *PayoutHandler) UpdatePayoutPreferences(w http.ResponseWriter, r *http.Request) {
	profileID, ok := middleware.GetEnrichmentProfileID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Profile ID not found in context")
		return
	}

	var req UpdatePayoutPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	// Validate required fields
	if req.BeneficiaryName == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Beneficiary name is required")
		return
	}
	if req.Country == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Country is required")
		return
	}
	if req.Currency == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Currency is required")
		return
	}
	if req.PreferredMethod == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Preferred method is required")
		return
	}

	// Validate payout method
	method := domain.PayoutMethod(req.PreferredMethod)
	if method != domain.PayoutMethodBankTransfer && method != domain.PayoutMethodPayPal && method != domain.PayoutMethodCrypto {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid payout method. Must be 'bank_transfer', 'paypal', or 'crypto'")
		return
	}

	// Set payout ready if provided
	payoutReady := false
	if req.PayoutReady != nil {
		payoutReady = *req.PayoutReady
	}

	data, err := h.payoutSvc.UpdatePayoutPreferences(r.Context(), profileID, method, req.BeneficiaryName, req.Country, req.Currency, req.EncryptedDetails, payoutReady)
	if err != nil {
		if err == domain.ErrInvalidCountryCode {
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid country code (ISO 3166-1 alpha-2)")
			return
		}
		if err == domain.ErrInvalidCurrencyCode {
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid currency code (ISO 4217)")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update payout preferences")
		return
	}

	response := toPayoutPreferencesResponse(data)
	writeJSON(w, http.StatusOK, response)
}

// toPayoutPreferencesResponse converts masked payout preferences to response.
func toPayoutPreferencesResponse(p *domain.PayoutPreferencesMasked) PayoutPreferencesResponse {
	return PayoutPreferencesResponse{
		ProfileID:       p.ProfileID.String(),
		PreferredMethod: string(p.PreferredMethod),
		BeneficiaryName: p.BeneficiaryName,
		Country:         p.Country,
		Currency:        p.Currency,
		PayoutReady:     p.PayoutReady,
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
