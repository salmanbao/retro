package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// TwoFactorHandler handles 2FA HTTP endpoints.
type TwoFactorHandler struct {
	twoFactorSvc *service.TwoFactorService
	authSvc      *service.AuthService
}

// NewTwoFactorHandler creates a new TwoFactorHandler.
func NewTwoFactorHandler(twoFactorSvc *service.TwoFactorService, authSvc *service.AuthService) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFactorSvc: twoFactorSvc,
		authSvc:      authSvc,
	}
}

// Setup2FAResponse represents the 2FA setup response.
type Setup2FAResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// Setup2FA handles POST /api/v1/auth/2fa/setup.
func (h *TwoFactorHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	// Check if user has verified email
	if !user.Verified {
		writeError(w, http.StatusForbidden, "email_not_verified", "Please verify your email before enabling 2FA")
		return
	}

	// Check if 2FA is already enabled
	enabled, err := h.twoFactorSvc.Is2FAEnabled(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to check 2FA status")
		return
	}
	if enabled {
		writeError(w, http.StatusConflict, "2fa_already_enabled", "2FA is already enabled on this account")
		return
	}

	// Generate TOTP secret and backup codes
	secret, qrCodeURL, backupCodes, err := h.twoFactorSvc.Setup2FA(r.Context(), user.ID, "ViralForge")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to setup 2FA")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Setup2FAResponse{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	})
}

// Verify2FARequest represents the 2FA verification request body.
type Verify2FARequest struct {
	Code string `json:"code"`
}

// Verify2FA handles POST /api/v1/auth/2fa/verify.
func (h *TwoFactorHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Verification code is required")
		return
	}

	err := h.twoFactorSvc.Verify2FASetup(r.Context(), user.ID, req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_code", "Invalid verification code")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "2FA enabled successfully"})
}

// Disable2FARequest represents the 2FA disable request body.
type Disable2FARequest struct {
	BackupCode string `json:"backup_code"`
}

// Disable2FA handles POST /api/v1/auth/2fa/disable.
func (h *TwoFactorHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req Disable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.BackupCode == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Backup code is required")
		return
	}

	err := h.twoFactorSvc.Disable2FAWithBackupCode(r.Context(), user.ID, req.BackupCode)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_backup_code", "Invalid backup code")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "2FA disabled successfully"})
}

// Validate2FARequest represents the 2FA validation request body (for login flow).
type Validate2FARequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// Validate2FA handles POST /api/v1/auth/2fa/validate.
// This is used during login when 2FA is enabled.
// It validates the TOTP code and returns the session token if successful.
func (h *TwoFactorHandler) Validate2FA(w http.ResponseWriter, r *http.Request) {
	var req Validate2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Verification code is required")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "User ID is required")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid user ID format")
		return
	}

	valid, err := h.twoFactorSvc.ValidateTOTP(r.Context(), userID, req.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to validate 2FA")
		return
	}

	if !valid {
		writeError(w, http.StatusUnauthorized, "invalid_code", "Invalid verification code")
		return
	}

	// Get the pending session for this user (most recent session created during login)
	sessions, err := h.authSvc.GetPendingSession(r.Context(), userID)
	if err != nil || len(sessions) == 0 {
		writeError(w, http.StatusInternalServerError, "internal_error", "No pending session found. Please log in again.")
		return
	}

	// The most recent session is the pending one from login
	session := sessions[0]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":      true,
		"token":      session.TokenHash,
		"expires_at": session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// RegisterRoutes registers 2FA routes on the router.
func (h *TwoFactorHandler) RegisterRoutes(r chi.Router) {
	r.Post("/setup", h.Setup2FA)
	r.Post("/verify", h.Verify2FA)
	r.Post("/disable", h.Disable2FA)
	r.Post("/validate", h.Validate2FA)
}

// getUserFromContext is a helper to get user from context.
func getUserFromContext(ctx context.Context) *domain.User {
	// This would normally use middleware.GetUserFromContext but for simplicity
	// we'll use a type assertion since the middleware stores it there
	if user, ok := ctx.Value("user").(*domain.User); ok {
		return user
	}
	return nil
}
