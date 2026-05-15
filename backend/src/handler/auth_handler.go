package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authSvc         *service.AuthService
	loginHistorySvc *service.LoginHistoryService
	twoFactorSvc    *service.TwoFactorService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *service.AuthService, loginHistorySvc *service.LoginHistoryService, twoFactorSvc *service.TwoFactorService) *AuthHandler {
	return &AuthHandler{
		authSvc:         authSvc,
		loginHistorySvc: loginHistorySvc,
		twoFactorSvc:    twoFactorSvc,
	}
}

// RegisterRequest represents the registration request body.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents a successful registration response.
type RegisterResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}

	err := h.authSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidEmailFormat) {
			writeError(w, http.StatusBadRequest, "invalid_email", "Invalid email format")
		} else if errors.Is(err, domain.ErrInvalidPasswordFormat) {
			writeError(w, http.StatusBadRequest, "invalid_password", "Password must be at least 8 characters with uppercase, lowercase, number, and special character")
		} else if errors.Is(err, domain.ErrEmailAlreadyRegistered) {
			writeError(w, http.StatusConflict, "email_exists", "Email is already registered")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		Message: "Account created. Please check your email to verify your address.",
	})
}

// VerifyEmailRequest represents the email verification request body.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail handles POST /api/v1/auth/verify-email.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Token is required")
		return
	}

	err := h.authSvc.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) || errors.Is(err, domain.ErrTokenExpired) || errors.Is(err, domain.ErrTokenAlreadyUsed) {
			writeError(w, http.StatusBadRequest, "invalid_token", "Token is invalid, expired, or already used")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email verified successfully"})
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a successful login response.
type LoginResponse struct {
	Token      string `json:"token,omitempty"`
	ExpiresAt  string `json:"expires_at"`
	UserID     string `json:"user_id,omitempty"`
	Requires2FA bool  `json:"requires_2fa,omitempty"`
}

// Login handles POST /api/v1/auth/login.
// If 2FA is enabled, returns 200 with requires_2fa=true and user_id for 2FA verification.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Password is required")
		return
	}

	userAgent := r.UserAgent()
	ipAddress := getClientIP(r)
	session, err := h.authSvc.Login(r.Context(), req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect")
		} else if errors.Is(err, domain.ErrEmailNotVerified) {
			writeError(w, http.StatusForbidden, "email_not_verified", "Please verify your email before logging in")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		}
		return
	}

	// Check if 2FA is enabled for this user
	if h.twoFactorSvc != nil {
		enabled, _ := h.twoFactorSvc.Is2FAEnabled(r.Context(), session.UserID)
		if enabled {
			// 2FA required - return partial response for 2FA verification
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(LoginResponse{
				Token:     "", // Empty token - requires 2FA
				ExpiresAt: session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
				UserID:    session.UserID.String(),
				Requires2FA: true,
			})
			return
		}
	}

	// Record login history
	if h.loginHistorySvc != nil {
		fingerprint := service.GenerateDeviceFingerprint(userAgent, "", "", "")
		h.loginHistorySvc.RecordLogin(r.Context(), session.UserID, ipAddress, userAgent, fingerprint)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token:     session.TokenHash,
		ExpiresAt: session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// LogoutRequest represents the logout request body (empty for bearer token logout).
type LogoutRequest struct{}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authorization header required")
		return
	}

	err := h.authSvc.Logout(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid_session", "Session not found or already expired")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MeResponse represents the current user response.
type MeResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Verified  bool   `json:"verified"`
	CreatedAt string `json:"created_at"`
}

// Me handles GET /api/v1/auth/me.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MeResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Verified:  user.Verified,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ForgotPasswordRequest represents the password reset request body.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword handles POST /api/v1/auth/forgot-password.
// Always returns 200 OK to prevent email enumeration attacks.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}

	// Always return success to prevent email enumeration
	h.authSvc.RequestPasswordReset(r.Context(), req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If an account exists with this email, a password reset link has been sent."})
}

// ResetPasswordRequest represents the password reset confirmation body.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles POST /api/v1/auth/reset-password.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Token is required")
		return
	}
	if req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "New password is required")
		return
	}

	err := h.authSvc.ConfirmPasswordReset(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPasswordFormat) {
			writeError(w, http.StatusBadRequest, "invalid_password", "Password must be at least 8 characters with uppercase, lowercase, number, and special character")
		} else if errors.Is(err, domain.ErrTokenNotFound) || errors.Is(err, domain.ErrTokenExpired) || errors.Is(err, domain.ErrTokenAlreadyUsed) {
			writeError(w, http.StatusBadRequest, "invalid_token", "Token is invalid, expired, or already used")
		} else {
			writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully. Please log in with your new password."})
}

// LoginHistoryResponse represents the login history response.
type LoginHistoryResponse struct {
	History    []*domain.LoginHistory `json:"history"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// GetLoginHistory handles GET /api/v1/auth/login-history.
func (h *AuthHandler) GetLoginHistory(w http.ResponseWriter, r *http.Request) {
	if h.loginHistorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "Login history service not available")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	histories, total, err := h.loginHistorySvc.GetLoginHistory(r.Context(), user.ID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve login history")
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginHistoryResponse{
		History:    histories,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// RegisterRoutes registers auth routes on the router.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/verify-email", h.VerifyEmail)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Post("/reset-password", h.ResetPassword)
	r.Get("/me", h.Me)
}

// RegisterLoginHistoryRoutes registers the login history route.
func (h *AuthHandler) RegisterLoginHistoryRoutes(r chi.Router) {
	r.Get("/login-history", h.GetLoginHistory)
}

func writeError(w http.ResponseWriter, status int, err, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err, Message: msg})
}

// extractBearerToken extracts the bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// getClientIP extracts the real client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}