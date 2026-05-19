package submission

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/repository"
	"viralforge/backend/src/service"
)

// Handler handles submission HTTP endpoints.
type Handler struct {
	submissionSvc *service.SubmissionService
	profileRepo   interface {
		ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	}
}

// NewHandler creates a new submission handler.
func NewHandler(
	submissionSvc *service.SubmissionService,
	profileRepo interface {
		ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	},
) *Handler {
	return &Handler{
		submissionSvc: submissionSvc,
		profileRepo:   profileRepo,
	}
}

// CreateRequest represents a submission creation request.
type CreateRequest struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	VideoURL        string   `json:"video_url"`
	ThumbnailURL    string   `json:"thumbnail_url"`
	DurationSeconds int      `json:"duration_seconds"`
	Notes           string   `json:"notes"`
	Tags            []string `json:"tags"`
}

// UpdateRequest represents a submission update request.
type UpdateRequest struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	VideoURL        string   `json:"video_url"`
	ThumbnailURL    string   `json:"thumbnail_url"`
	DurationSeconds int      `json:"duration_seconds"`
	Notes           string   `json:"notes"`
	Tags            []string `json:"tags"`
}

// Response represents a generic submission response.
type Response struct {
	ID              string   `json:"id"`
	CampaignID      string   `json:"campaign_id"`
	EditorProfileID string   `json:"editor_profile_id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	VideoURL        string   `json:"video_url"`
	ThumbnailURL    string   `json:"thumbnail_url"`
	DurationSeconds int      `json:"duration_seconds"`
	Notes           string   `json:"notes"`
	Tags            []string `json:"tags"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// RegisterRoutes registers submission routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.ListByCampaign)
}

// listByCampaignResponse represents campaign submission list response.
type listByCampaignResponse struct {
	Submissions []*Response `json:"submissions"`
	Total       int         `json:"total"`
}

// Create handles POST /api/v1/campaigns/{campaignId}/submissions
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	campaignIDStr := chi.URLParam(r, "campaignId")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "invalid campaign ID", http.StatusBadRequest)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get editor profile ID from context (authenticated user)
	editorProfileIDPtr := middleware.GetActiveProfileID(ctx)
	if editorProfileIDPtr == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	editorProfileID := *editorProfileIDPtr

	input := service.CreateSubmissionInput{
		CampaignID:      campaignID,
		Title:           req.Title,
		Description:     req.Description,
		VideoURL:        req.VideoURL,
		ThumbnailURL:    req.ThumbnailURL,
		DurationSeconds: req.DurationSeconds,
		Notes:           req.Notes,
		Tags:            req.Tags,
	}

	submission, err := h.submissionSvc.CreateDraft(ctx, editorProfileID, input)
	if err != nil {
		if errors.Is(err, service.ErrNotEligible) ||
			errors.Is(err, service.ErrCampaignNotAccepting) ||
			errors.Is(err, service.ErrDeadlinePassed) ||
			errors.Is(err, service.ErrNoCreativeBrief) ||
			errors.Is(err, service.ErrNoAssets) ||
			errors.Is(err, service.ErrDuplicateSubmission) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if errors.Is(err, domain.ErrCampaignNotFound) {
			http.Error(w, "campaign not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to create submission", http.StatusInternalServerError)
		return
	}

	resp := toResponse(submission)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ListByCampaign handles GET /api/v1/campaigns/{campaignId}/submissions
func (h *Handler) ListByCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	campaignIDStr := chi.URLParam(r, "campaignId")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "invalid campaign ID", http.StatusBadRequest)
		return
	}

	submissions, err := h.submissionSvc.GetByCampaignID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			http.Error(w, "campaign not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to list submissions", http.StatusInternalServerError)
		return
	}

	resp := &listByCampaignResponse{
		Submissions: make([]*Response, len(submissions)),
		Total:       len(submissions),
	}
	for i, s := range submissions {
		resp.Submissions[i] = toResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetByID handles GET /api/v1/submissions/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid submission ID", http.StatusBadRequest)
		return
	}

	submission, err := h.submissionSvc.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	resp := toResponse(submission)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Update handles PATCH /api/v1/submissions/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid submission ID", http.StatusBadRequest)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	editorProfileIDPtr := middleware.GetActiveProfileID(ctx)
	if editorProfileIDPtr == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	editorProfileID := *editorProfileIDPtr

	submission, err := h.submissionSvc.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	input := service.UpdateSubmissionInput{
		Title:           req.Title,
		Description:     req.Description,
		VideoURL:        req.VideoURL,
		ThumbnailURL:    req.ThumbnailURL,
		DurationSeconds: req.DurationSeconds,
		Notes:           req.Notes,
		Tags:            req.Tags,
	}

	err = h.submissionSvc.UpdateDraft(ctx, editorProfileID, submission, input)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			http.Error(w, "not authorized to update this submission", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrCannotEditNonDraft) {
			http.Error(w, "cannot edit non-draft submission", http.StatusConflict)
			return
		}
		http.Error(w, "failed to update submission", http.StatusInternalServerError)
		return
	}

	resp := toResponse(submission)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Submit handles POST /api/v1/submissions/{id}/submit
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid submission ID", http.StatusBadRequest)
		return
	}

	editorProfileIDPtr := middleware.GetActiveProfileID(ctx)
	if editorProfileIDPtr == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	editorProfileID := *editorProfileIDPtr

	submission, err := h.submissionSvc.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	err = h.submissionSvc.Submit(ctx, editorProfileID, submission)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			http.Error(w, "not authorized to submit this submission", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrCannotSubmitNonDraft) {
			http.Error(w, "cannot submit non-draft submission", http.StatusConflict)
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			http.Error(w, "invalid state transition", http.StatusConflict)
			return
		}
		http.Error(w, "failed to submit submission", http.StatusInternalServerError)
		return
	}

	resp := toResponse(submission)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Withdraw handles POST /api/v1/submissions/{id}/withdraw
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid submission ID", http.StatusBadRequest)
		return
	}

	editorProfileIDPtr := middleware.GetActiveProfileID(ctx)
	if editorProfileIDPtr == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	editorProfileID := *editorProfileIDPtr

	submission, err := h.submissionSvc.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSubmissionNotFound) {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get submission", http.StatusInternalServerError)
		return
	}

	err = h.submissionSvc.Withdraw(ctx, editorProfileID, submission)
	if err != nil {
		if errors.Is(err, service.ErrNotOwner) {
			http.Error(w, "not authorized to withdraw this submission", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrCannotWithdraw) {
			http.Error(w, "cannot withdraw submission in current state", http.StatusConflict)
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			http.Error(w, "invalid state transition", http.StatusConflict)
			return
		}
		http.Error(w, "failed to withdraw submission", http.StatusInternalServerError)
		return
	}

	resp := toResponse(submission)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func toResponse(s *domain.Submission) *Response {
	resp := &Response{
		ID:              s.ID.String(),
		CampaignID:      s.CampaignID.String(),
		EditorProfileID: s.EditorProfileID.String(),
		Title:           s.Title,
		Description:     s.Description,
		VideoURL:        s.VideoURL,
		ThumbnailURL:    s.ThumbnailURL,
		DurationSeconds: s.DurationSeconds,
		Notes:           s.Notes,
		Tags:            s.Tags,
		Status:          string(s.Status),
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
	}
	return resp
}
