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

// SessionHandler handles session-related HTTP endpoints.
type SessionHandler struct {
	sessionSvc *service.SessionService
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessionSvc *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionSvc: sessionSvc}
}

// SwitchActiveProfileRequest represents the request to switch active profile.
type SwitchActiveProfileRequest struct {
	ProfileID string `json:"profile_id"`
}

// SessionResponse represents a session in API responses.
type SessionResponse struct {
	ID              string `json:"id"`
	ActiveProfileID string `json:"active_profile_id,omitempty"`
	ExpiresAt       string `json:"expires_at"`
	CreatedAt       string `json:"created_at"`
}

// SwitchActiveProfile handles PATCH /api/v1/sessions/active.
func (h *SessionHandler) SwitchActiveProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	session := middleware.GetSessionFromContext(r.Context())
	if session == nil {
		writeError(w, http.StatusUnauthorized, "no_session", "No session found")
		return
	}

	var req SwitchActiveProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.ProfileID == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Profile ID is required")
		return
	}

	profileID, err := uuid.Parse(req.ProfileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_profile_id", "Profile ID must be a valid UUID")
		return
	}

	updatedSession, err := h.sessionSvc.SwitchActiveProfile(r.Context(), user.ID, session.ID, profileID)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotOwned) {
			writeError(w, http.StatusForbidden, "forbidden", "Profile does not belong to you")
			return
		}
		if errors.Is(err, domain.ErrSessionNotFound) || errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Session or profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var activeProfileID string
	if updatedSession.ActiveProfileID != nil {
		activeProfileID = updatedSession.ActiveProfileID.String()
	}
	json.NewEncoder(w).Encode(SessionResponse{
		ID:              updatedSession.ID.String(),
		ActiveProfileID: activeProfileID,
		ExpiresAt:       updatedSession.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:       updatedSession.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListSessions handles GET /api/v1/sessions.
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	sessions, err := h.sessionSvc.ListSessions(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	var response []SessionResponse
	for _, sess := range sessions {
		var activeProfileID string
		if sess.ActiveProfileID != nil {
			activeProfileID = sess.ActiveProfileID.String()
		}
		response = append(response, SessionResponse{
			ID:              sess.ID.String(),
			ActiveProfileID: activeProfileID,
			ExpiresAt:       sess.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:       sess.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	if response == nil {
		response = []SessionResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteSession handles DELETE /api/v1/sessions/{id}.
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Session ID must be a valid UUID")
		return
	}

	err = h.sessionSvc.RevokeSession(r.Context(), user.ID, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterRoutes registers session routes on the router.
func (h *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListSessions)
	r.Delete("/{id}", h.DeleteSession)
}