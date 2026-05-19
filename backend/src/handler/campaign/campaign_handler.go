package campaign

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// Handler handles campaign HTTP endpoints.
type Handler struct {
	campaignSvc service.CampaignServiceInterface
}

// NewHandler creates a new campaign handler.
func NewHandler(campaignSvc service.CampaignServiceInterface) *Handler {
	return &Handler{campaignSvc: campaignSvc}
}

// CreateRequest represents a campaign creation request.
type CreateRequest struct {
	Title              string   `json:"title"`
	Summary            string   `json:"summary"`
	Description        string   `json:"description"`
	Objective          string   `json:"objective"`
	ProductName        string   `json:"product_name"`
	LandingURL         string   `json:"landing_url"`
	TotalBudget        float64  `json:"total_budget"`
	Currency           string   `json:"currency"`
	TargetClips        int      `json:"target_clips"`
	TargetPosts        int      `json:"target_posts"`
	CPV                float64  `json:"cpv"`
	MinPayout          *float64 `json:"min_payout,omitempty"`
	MaxPayout          *float64 `json:"max_payout,omitempty"`
	SubmissionStart    string   `json:"submission_start"`
	SubmissionDeadline string   `json:"submission_deadline"`
	DistributionStart  string   `json:"distribution_start"`
	CampaignEnd        string   `json:"campaign_end"`
	AllowedCountries   []string `json:"allowed_countries"`
	AllowedLanguages   []string `json:"allowed_languages"`
	MinFollowers       int      `json:"min_followers"`
	Platforms          []string `json:"platforms"`
	CreatorCategories  []string `json:"creator_categories"`
	MinDurationSecs    int      `json:"min_duration_secs"`
	MaxDurationSecs    int      `json:"max_duration_secs"`
	AspectRatio        string   `json:"aspect_ratio"`
	TalkingPoints      []string `json:"talking_points"`
	ProhibitedClaims   []string `json:"prohibited_claims"`
	Hashtags           []string `json:"hashtags"`
	CTAInstructions    string   `json:"cta_instructions"`
}

// CreateResponse represents a campaign creation response.
type CreateResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ListResponse represents a campaign list response.
type ListResponse struct {
	Campaigns  []*CampaignListItem `json:"campaigns"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// CampaignListItem represents a campaign in a list.
type CampaignListItem struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Slug               string  `json:"slug"`
	Status             string  `json:"status"`
	TotalBudget        float64 `json:"total_budget"`
	Currency           string  `json:"currency"`
	SubmissionDeadline string  `json:"submission_deadline"`
	CreatedAt          string  `json:"created_at"`
}

// CampaignDetailResponse represents full campaign details.
type CampaignDetailResponse struct {
	ID                 string           `json:"id"`
	BrandProfileID     string           `json:"brand_profile_id"`
	Title              string           `json:"title"`
	Slug               string           `json:"slug"`
	Summary            string           `json:"summary"`
	Description        string           `json:"description"`
	Objective          string           `json:"objective"`
	ProductName        string           `json:"product_name"`
	LandingURL         string           `json:"landing_url"`
	TotalBudget        float64          `json:"total_budget"`
	Currency           string           `json:"currency"`
	TargetClips        int              `json:"target_clips"`
	TargetPosts        int              `json:"target_posts"`
	CPV                float64          `json:"cpv"`
	MinPayout          *float64         `json:"min_payout,omitempty"`
	MaxPayout          *float64         `json:"max_payout,omitempty"`
	SubmissionStart    string           `json:"submission_start"`
	SubmissionDeadline string           `json:"submission_deadline"`
	DistributionStart  string           `json:"distribution_start"`
	CampaignEnd        string           `json:"campaign_end"`
	AllowedCountries   []string         `json:"allowed_countries"`
	AllowedLanguages   []string         `json:"allowed_languages"`
	MinFollowers       int              `json:"min_followers"`
	Platforms          []string         `json:"platforms"`
	CreatorCategories  []string         `json:"creator_categories"`
	MinDurationSecs    int              `json:"min_duration_secs"`
	MaxDurationSecs    int              `json:"max_duration_secs"`
	AspectRatio        string           `json:"aspect_ratio"`
	TalkingPoints      []string         `json:"talking_points"`
	ProhibitedClaims   []string         `json:"prohibited_claims"`
	Hashtags           []string         `json:"hashtags"`
	CTAInstructions    string           `json:"cta_instructions"`
	Assets             []*AssetResponse `json:"assets,omitempty"`
	Status             string           `json:"status"`
	Version            int              `json:"version"`
	CreatedAt          string           `json:"created_at"`
	UpdatedAt          string           `json:"updated_at"`
}

// AssetResponse represents a campaign asset.
type AssetResponse struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	AssetType   string `json:"asset_type"`
	Description string `json:"description,omitempty"`
}

// UpdateRequest represents a campaign update request.
type UpdateRequest map[string]interface{}

// StatusResponse represents a status change response.
type StatusResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at,omitempty"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// parseTime parses an ISO8601 time string.
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func writeError(w http.ResponseWriter, status int, err, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err, Message: msg})
}

// Create handles POST /api/v1/campaigns
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Title is required")
		return
	}
	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Description is required")
		return
	}

	// Get brand profile ID from context (set by middleware)
	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	// Parse times
	submissionStart, err := parseTime(req.SubmissionStart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid submission_start format")
		return
	}
	submissionDeadline, err := parseTime(req.SubmissionDeadline)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid submission_deadline format")
		return
	}
	distributionStart, err := parseTime(req.DistributionStart)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid distribution_start format")
		return
	}
	campaignEnd, err := parseTime(req.CampaignEnd)
	if err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid campaign_end format")
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.MinDurationSecs == 0 {
		req.MinDurationSecs = 15
	}
	if req.MaxDurationSecs == 0 {
		req.MaxDurationSecs = 60
	}
	if req.AspectRatio == "" {
		req.AspectRatio = "9:16"
	}

	input := service.CreateCampaignInput{
		BrandProfileID:     brandProfileID,
		Title:              req.Title,
		Summary:            req.Summary,
		Description:        req.Description,
		Objective:          req.Objective,
		ProductName:        req.ProductName,
		LandingURL:         req.LandingURL,
		TotalBudget:        req.TotalBudget,
		Currency:           req.Currency,
		TargetClips:        req.TargetClips,
		TargetPosts:        req.TargetPosts,
		CPV:                req.CPV,
		MinPayout:          req.MinPayout,
		MaxPayout:          req.MaxPayout,
		SubmissionStart:    submissionStart,
		SubmissionDeadline: submissionDeadline,
		DistributionStart:  distributionStart,
		CampaignEnd:        campaignEnd,
		AllowedCountries:   req.AllowedCountries,
		AllowedLanguages:   req.AllowedLanguages,
		MinFollowers:       req.MinFollowers,
		Platforms:          req.Platforms,
		CreatorCategories:  req.CreatorCategories,
		MinDurationSecs:    req.MinDurationSecs,
		MaxDurationSecs:    req.MaxDurationSecs,
		AspectRatio:        req.AspectRatio,
		TalkingPoints:      req.TalkingPoints,
		ProhibitedClaims:   req.ProhibitedClaims,
		Hashtags:           req.Hashtags,
		CTAInstructions:    req.CTAInstructions,
	}

	campaign, err := h.campaignSvc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrSlugAlreadyExists) {
			writeError(w, http.StatusConflict, "slug_exists", "Campaign with this slug already exists")
			return
		}
		if errors.Is(err, domain.ErrInvalidTimeline) {
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid timeline configuration")
			return
		}
		if errors.Is(err, domain.ErrInvalidPayoutRange) {
			writeError(w, http.StatusBadRequest, "validation_error", "Minimum payout cannot exceed maximum payout")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateResponse{
		ID:        campaign.ID.String(),
		Title:     campaign.Title,
		Slug:      campaign.Slug,
		Status:    string(campaign.Status),
		CreatedAt: campaign.CreatedAt.Format(time.RFC3339),
	})
}

// List handles GET /api/v1/campaigns
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
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
	status := r.URL.Query().Get("status")

	campaigns, total, err := h.campaignSvc.ListByBrandProfile(r.Context(), brandProfileID, status, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve campaigns")
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	items := make([]*CampaignListItem, len(campaigns))
	for i, c := range campaigns {
		items[i] = &CampaignListItem{
			ID:                 c.ID.String(),
			Title:              c.Title,
			Slug:               c.Slug,
			Status:             string(c.Status),
			TotalBudget:        c.TotalBudget,
			Currency:           c.Currency,
			SubmissionDeadline: c.SubmissionDeadline.Format(time.RFC3339),
			CreatedAt:          c.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListResponse{
		Campaigns:  items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// Get handles GET /api/v1/campaigns/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	campaign, err := h.campaignSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve campaign")
		return
	}

	if campaign.BrandProfileID != brandProfileID {
		writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
		return
	}

	assets := make([]*AssetResponse, len(campaign.Assets))
	for i, a := range campaign.Assets {
		assets[i] = &AssetResponse{
			ID:          a.ID.String(),
			URL:         a.URL,
			AssetType:   string(a.AssetType),
			Description: a.Description,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CampaignDetailResponse{
		ID:                 campaign.ID.String(),
		BrandProfileID:     campaign.BrandProfileID.String(),
		Title:              campaign.Title,
		Slug:               campaign.Slug,
		Summary:            campaign.Summary,
		Description:        campaign.Description,
		Objective:          campaign.Objective,
		ProductName:        campaign.ProductName,
		LandingURL:         campaign.LandingURL,
		TotalBudget:        campaign.TotalBudget,
		Currency:           campaign.Currency,
		TargetClips:        campaign.TargetClips,
		TargetPosts:        campaign.TargetPosts,
		CPV:                campaign.CPV,
		MinPayout:          campaign.MinPayout,
		MaxPayout:          campaign.MaxPayout,
		SubmissionStart:    campaign.SubmissionStart.Format(time.RFC3339),
		SubmissionDeadline: campaign.SubmissionDeadline.Format(time.RFC3339),
		DistributionStart:  campaign.DistributionStart.Format(time.RFC3339),
		CampaignEnd:        campaign.CampaignEnd.Format(time.RFC3339),
		AllowedCountries:   campaign.AllowedCountries,
		AllowedLanguages:   campaign.AllowedLanguages,
		MinFollowers:       campaign.MinFollowers,
		Platforms:          campaign.Platforms,
		CreatorCategories:  campaign.CreatorCategories,
		MinDurationSecs:    campaign.MinDurationSecs,
		MaxDurationSecs:    campaign.MaxDurationSecs,
		AspectRatio:        campaign.AspectRatio,
		TalkingPoints:      campaign.TalkingPoints,
		ProhibitedClaims:   campaign.ProhibitedClaims,
		Hashtags:           campaign.Hashtags,
		CTAInstructions:    campaign.CTAInstructions,
		Assets:             assets,
		Status:             string(campaign.Status),
		Version:            campaign.Version,
		CreatedAt:          campaign.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          campaign.UpdatedAt.Format(time.RFC3339),
	})
}

// Update handles PATCH /api/v1/campaigns/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	input := service.UpdateInput{
		CampaignID:     id,
		BrandProfileID: brandProfileID,
	}

	// Parse field updates
	if v, ok := req["title"].(string); ok {
		input.Title = &v
	}
	if v, ok := req["summary"].(string); ok {
		input.Summary = &v
	}
	if v, ok := req["description"].(string); ok {
		input.Description = &v
	}
	if v, ok := req["total_budget"].(float64); ok {
		input.TotalBudget = &v
	}
	if v, ok := req["min_payout"].(float64); ok {
		input.MinPayout = &v
	}
	if v, ok := req["max_payout"].(float64); ok {
		input.MaxPayout = &v
	}

	campaign, err := h.campaignSvc.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrCampaignNotOwned) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrRestrictedEdit) {
			writeError(w, http.StatusBadRequest, "restricted_edit", "This field cannot be edited in the current campaign status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update campaign")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		ID:        campaign.ID.String(),
		Status:    string(campaign.Status),
		UpdatedAt: campaign.UpdatedAt.Format(time.RFC3339),
	})
}

// Cancel handles POST /api/v1/campaigns/{id}/cancel
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	campaign, err := h.campaignSvc.Cancel(r.Context(), id, brandProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrCampaignNotOwned) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusConflict, "invalid_transition", "Cannot cancel campaign in current status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to cancel campaign")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var deletedAt string
	if campaign.DeletedAt != nil {
		deletedAt = campaign.DeletedAt.Format(time.RFC3339)
	}
	json.NewEncoder(w).Encode(StatusResponse{
		ID:        campaign.ID.String(),
		Status:    string(campaign.Status),
		DeletedAt: deletedAt,
	})
}

// Publish handles POST /api/v1/campaigns/{id}/publish
func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	campaign, err := h.campaignSvc.Publish(r.Context(), id, brandProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrCampaignNotOwned) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusConflict, "invalid_transition", "Cannot publish campaign in current status")
			return
		}
		if errors.Is(err, domain.ErrBudgetRequired) {
			writeError(w, http.StatusBadRequest, "budget_required", "Total budget must be greater than zero")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to publish campaign")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		ID:        campaign.ID.String(),
		Status:    string(campaign.Status),
		UpdatedAt: campaign.UpdatedAt.Format(time.RFC3339),
	})
}

// Pause handles POST /api/v1/campaigns/{id}/pause
func (h *Handler) Pause(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	campaign, err := h.campaignSvc.Pause(r.Context(), id, brandProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrCampaignNotOwned) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusConflict, "invalid_transition", "Cannot pause campaign in current status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to pause campaign")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		ID:        campaign.ID.String(),
		Status:    string(campaign.Status),
		UpdatedAt: campaign.UpdatedAt.Format(time.RFC3339),
	})
}

// Resume handles POST /api/v1/campaigns/{id}/resume
func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	campaign, err := h.campaignSvc.Resume(r.Context(), id, brandProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrCampaignNotOwned) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusConflict, "invalid_transition", "Cannot resume campaign in current status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to resume campaign")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		ID:        campaign.ID.String(),
		Status:    string(campaign.Status),
		UpdatedAt: campaign.UpdatedAt.Format(time.RFC3339),
	})
}

// Complete handles POST /api/v1/campaigns/{id}/complete
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid campaign ID")
		return
	}

	brandProfileIDStr := r.Context().Value("brand_profile_id")
	if brandProfileIDStr == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Brand profile required")
		return
	}
	brandProfileID, ok := brandProfileIDStr.(uuid.UUID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Invalid profile context")
		return
	}

	campaign, err := h.campaignSvc.Complete(r.Context(), id, brandProfileID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrCampaignNotOwned) {
			writeError(w, http.StatusNotFound, "not_found", "Campaign not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusConflict, "invalid_transition", "Cannot complete campaign in current status")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to complete campaign")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		ID:        campaign.ID.String(),
		Status:    string(campaign.Status),
		UpdatedAt: campaign.UpdatedAt.Format(time.RFC3339),
	})
}

// RegisterRoutes registers campaign routes on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Post("/{id}/publish", h.Publish)
	r.Post("/{id}/pause", h.Pause)
	r.Post("/{id}/resume", h.Resume)
	r.Post("/{id}/complete", h.Complete)
	r.Post("/{id}/cancel", h.Cancel)
}
