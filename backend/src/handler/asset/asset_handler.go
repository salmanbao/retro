package asset

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
	"viralforge/backend/src/service"
)

type Handler struct {
	assetSvc    *service.AssetService
	campaignSvc service.CampaignServiceInterface
	profileRepo interface {
		ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	}
}

func NewHandler(assetSvc *service.AssetService, campaignSvc service.CampaignServiceInterface, profileRepo interface {
	ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
}) *Handler {
	return &Handler{
		assetSvc:    assetSvc,
		campaignSvc: campaignSvc,
		profileRepo: profileRepo,
	}
}

type RegisterRequest struct {
	Category         string `json:"category"`
	OriginalFilename string `json:"original_filename"`
	DisplayName      string `json:"display_name"`
	MimeType         string `json:"mime_type"`
	FileSizeBytes    int64  `json:"file_size_bytes"`
	StorageKey       string `json:"storage_key"`
	Checksum         string `json:"checksum"`
}

type UpdateRequest struct {
	DisplayName      *string `json:"display_name,omitempty"`
	ProcessingStatus *string `json:"processing_status,omitempty"`
	VirusScanStatus  *string `json:"virus_scan_status,omitempty"`
}

type AssetResponse struct {
	ID                  uuid.UUID `json:"id"`
	CampaignID          uuid.UUID `json:"campaign_id"`
	Category            string    `json:"category"`
	OriginalFilename    string    `json:"original_filename"`
	DisplayName         string    `json:"display_name"`
	MimeType            string    `json:"mime_type"`
	FileSizeBytes       int64     `json:"file_size_bytes"`
	StorageKey          string    `json:"storage_key"`
	Checksum            string    `json:"checksum"`
	Version             int       `json:"version"`
	ProcessingStatus    string    `json:"processing_status"`
	VirusScanStatus     string    `json:"virus_scan_status"`
	UploadedByProfileID uuid.UUID `json:"uploaded_by_profile_id"`
	CreatedAt           string    `json:"created_at"`
	UpdatedAt           string    `json:"updated_at"`
}

type ListResponse struct {
	Data       []AssetResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := chi.URLParam(r, "campaignId")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_campaign_id", Message: "Invalid campaign ID format"})
		return
	}

	// Get brand profile ID from context (set by auth middleware)
	profileIDVal := r.Context().Value("active_profile_id")
	if profileIDVal == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}
	profileID, ok := profileIDVal.(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Invalid profile context"})
		return
	}

	// Get profile to check type
	profile, err := h.profileRepo.ByID(r.Context(), profileID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "Profile not found"})
		return
	}

	// Only brand owners can register assets
	if profile.Type != domain.ProfileTypeBrand {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Only brand profiles can register assets"})
		return
	}

	campaign, err := h.campaignSvc.GetByID(r.Context(), campaignID)
	if err != nil || campaign == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
		return
	}

	if campaign.BrandProfileID != profileID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "You do not own this campaign"})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "Invalid request body"})
		return
	}

	campaign, err = h.campaignSvc.GetByID(r.Context(), campaignID)
	if err != nil || campaign == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
		return
	}

	profileID = uuid.Nil

	input := service.RegisterAssetInput{
		CampaignID:          campaignID,
		Category:            domain.AssetCategory(req.Category),
		OriginalFilename:    req.OriginalFilename,
		DisplayName:         req.DisplayName,
		MimeType:            req.MimeType,
		FileSizeBytes:       req.FileSizeBytes,
		StorageKey:          req.StorageKey,
		Checksum:            req.Checksum,
		UploadedByProfileID: profileID,
	}

	asset, err := h.assetSvc.Register(r.Context(), input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "validation_failed", Message: err.Error()})
		return
	}

	response := AssetResponse{
		ID:                  asset.ID,
		CampaignID:          asset.CampaignID,
		Category:            string(asset.Category),
		OriginalFilename:    asset.OriginalFilename,
		DisplayName:         asset.DisplayName,
		MimeType:            asset.MimeType,
		FileSizeBytes:       asset.FileSizeBytes,
		StorageKey:          asset.StorageKey,
		Checksum:            asset.Checksum,
		Version:             asset.Version,
		ProcessingStatus:    string(asset.ProcessingStatus),
		VirusScanStatus:     string(asset.VirusScanStatus),
		UploadedByProfileID: asset.UploadedByProfileID,
		CreatedAt:           asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := chi.URLParam(r, "campaignId")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_campaign_id", Message: "Invalid campaign ID format"})
		return
	}

	// Optional auth check for List - allows both brand owners and editors
	profileIDVal := r.Context().Value("active_profile_id")
	if profileIDVal != nil {
		profileID, ok := profileIDVal.(uuid.UUID)
		if ok {
			profile, err := h.profileRepo.ByID(r.Context(), profileID)
			if err == nil {
				if profile.Type == domain.ProfileTypeBrand {
					// Brand owner - check ownership
					campaign, err := h.campaignSvc.GetByID(r.Context(), campaignID)
					if err != nil || campaign == nil {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
						return
					}
					if campaign.BrandProfileID != profileID {
						w.WriteHeader(http.StatusForbidden)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "You do not own this campaign"})
						return
					}
				} else if profile.Type == domain.ProfileTypeEditor {
					// Editor - check campaign status is published or active
					campaign, err := h.campaignSvc.GetByID(r.Context(), campaignID)
					if err != nil || campaign == nil {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
						return
					}
					if campaign.Status != domain.CampaignStatusPublished && campaign.Status != domain.CampaignStatusActive {
						w.WriteHeader(http.StatusForbidden)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Editors can only access assets for published or active campaigns"})
						return
					}
				} else {
					// Influencer or unknown type - deny access
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Access denied"})
					return
				}
			}
		}
	}

	page := 1
	pageSize := 20

	assets, total, err := h.assetSvc.ListByCampaign(r.Context(), service.ListAssetsInput{
		CampaignID: campaignID,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to list assets"})
		return
	}

	data := make([]AssetResponse, len(assets))
	for i, asset := range assets {
		data[i] = AssetResponse{
			ID:                  asset.ID,
			CampaignID:          asset.CampaignID,
			Category:            string(asset.Category),
			OriginalFilename:    asset.OriginalFilename,
			DisplayName:         asset.DisplayName,
			MimeType:            asset.MimeType,
			FileSizeBytes:       asset.FileSizeBytes,
			StorageKey:          asset.StorageKey,
			Checksum:            asset.Checksum,
			Version:             asset.Version,
			ProcessingStatus:    string(asset.ProcessingStatus),
			VirusScanStatus:     string(asset.VirusScanStatus),
			UploadedByProfileID: asset.UploadedByProfileID,
			CreatedAt:           asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:           asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListResponse{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	assetIDStr := chi.URLParam(r, "id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_asset_id", Message: "Invalid asset ID format"})
		return
	}

	// Optional auth check for Get - allows both brand owners and editors
	profileIDVal := r.Context().Value("active_profile_id")
	if profileIDVal != nil {
		profileID, ok := profileIDVal.(uuid.UUID)
		if ok {
			profile, err := h.profileRepo.ByID(r.Context(), profileID)
			if err == nil {
				// Get asset to find campaign
				asset, err := h.assetSvc.GetByID(r.Context(), assetID)
				if err != nil {
					if errors.Is(err, repository.ErrAssetNotFound) || errors.Is(err, domain.ErrAssetNotFound) {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "asset_not_found", Message: "Asset not found"})
						return
					}
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to retrieve asset"})
					return
				}

				if profile.Type == domain.ProfileTypeBrand {
					// Brand owner - check ownership
					campaign, err := h.campaignSvc.GetByID(r.Context(), asset.CampaignID)
					if err != nil || campaign == nil {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
						return
					}
					if campaign.BrandProfileID != profileID {
						w.WriteHeader(http.StatusForbidden)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "You do not own this campaign"})
						return
					}
				} else if profile.Type == domain.ProfileTypeEditor {
					// Editor - check campaign status is published or active
					campaign, err := h.campaignSvc.GetByID(r.Context(), asset.CampaignID)
					if err != nil || campaign == nil {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
						return
					}
					if campaign.Status != domain.CampaignStatusPublished && campaign.Status != domain.CampaignStatusActive {
						w.WriteHeader(http.StatusForbidden)
						json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Editors can only access assets for published or active campaigns"})
						return
					}
				} else {
					// Influencer or unknown type - deny access
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Access denied"})
					return
				}
			}
		}
	}

	asset, err := h.assetSvc.GetByID(r.Context(), assetID)
	if err != nil {
		if errors.Is(err, repository.ErrAssetNotFound) || errors.Is(err, domain.ErrAssetNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "asset_not_found", Message: "Asset not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to retrieve asset"})
		return
	}

	response := AssetResponse{
		ID:                  asset.ID,
		CampaignID:          asset.CampaignID,
		Category:            string(asset.Category),
		OriginalFilename:    asset.OriginalFilename,
		DisplayName:         asset.DisplayName,
		MimeType:            asset.MimeType,
		FileSizeBytes:       asset.FileSizeBytes,
		StorageKey:          asset.StorageKey,
		Checksum:            asset.Checksum,
		Version:             asset.Version,
		ProcessingStatus:    string(asset.ProcessingStatus),
		VirusScanStatus:     string(asset.VirusScanStatus),
		UploadedByProfileID: asset.UploadedByProfileID,
		CreatedAt:           asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	assetIDStr := chi.URLParam(r, "id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_asset_id", Message: "Invalid asset ID format"})
		return
	}

	// Get brand profile ID from context (set by auth middleware)
	profileIDVal := r.Context().Value("active_profile_id")
	if profileIDVal == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}
	profileID, ok := profileIDVal.(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Invalid profile context"})
		return
	}

	// Get profile to check type
	profile, err := h.profileRepo.ByID(r.Context(), profileID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "Profile not found"})
		return
	}

	// Only brand owners can update assets
	if profile.Type != domain.ProfileTypeBrand {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Only brand profiles can update assets"})
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "Invalid request body"})
		return
	}

	asset, err := h.assetSvc.GetByID(r.Context(), assetID)
	if err != nil {
		if errors.Is(err, repository.ErrAssetNotFound) || errors.Is(err, domain.ErrAssetNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "asset_not_found", Message: "Asset not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to retrieve asset"})
		return
	}

	updateInput := service.UpdateAssetInput{}
	if req.DisplayName != nil {
		updateInput.DisplayName = req.DisplayName
	}
	if req.ProcessingStatus != nil {
		status := domain.ProcessingStatus(*req.ProcessingStatus)
		updateInput.ProcessingStatus = &status
	}
	if req.VirusScanStatus != nil {
		status := domain.VirusScanStatus(*req.VirusScanStatus)
		updateInput.VirusScanStatus = &status
	}

	if err := h.assetSvc.Update(r.Context(), asset, updateInput); err != nil {
		if errors.Is(err, service.ErrAssetArchived) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "asset_archived", Message: "Archived assets cannot be updated"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to update asset"})
		return
	}

	response := AssetResponse{
		ID:                  asset.ID,
		CampaignID:          asset.CampaignID,
		Category:            string(asset.Category),
		OriginalFilename:    asset.OriginalFilename,
		DisplayName:         asset.DisplayName,
		MimeType:            asset.MimeType,
		FileSizeBytes:       asset.FileSizeBytes,
		StorageKey:          asset.StorageKey,
		Checksum:            asset.Checksum,
		Version:             asset.Version,
		ProcessingStatus:    string(asset.ProcessingStatus),
		VirusScanStatus:     string(asset.VirusScanStatus),
		UploadedByProfileID: asset.UploadedByProfileID,
		CreatedAt:           asset.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           asset.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	assetIDStr := chi.URLParam(r, "id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_asset_id", Message: "Invalid asset ID format"})
		return
	}

	// Get brand profile ID from context (set by auth middleware)
	profileIDVal := r.Context().Value("active_profile_id")
	if profileIDVal == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}
	profileID, ok := profileIDVal.(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Invalid profile context"})
		return
	}

	// Get profile to check type
	profile, err := h.profileRepo.ByID(r.Context(), profileID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized", Message: "Profile not found"})
		return
	}

	// Only brand owners can delete assets
	if profile.Type != domain.ProfileTypeBrand {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Only brand profiles can delete assets"})
		return
	}

	// Get asset to find campaign and verify ownership
	asset, err := h.assetSvc.GetByID(r.Context(), assetID)
	if err != nil {
		if errors.Is(err, repository.ErrAssetNotFound) || errors.Is(err, domain.ErrAssetNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "asset_not_found", Message: "Asset not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to retrieve asset"})
		return
	}

	campaign, err := h.campaignSvc.GetByID(r.Context(), asset.CampaignID)
	if err != nil || campaign == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
		return
	}

	if campaign.BrandProfileID != profileID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "You do not own this campaign"})
		return
	}

	if err := h.assetSvc.SoftDelete(r.Context(), assetID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to delete asset"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/{campaignId}/assets", h.Register)
	r.Get("/{campaignId}/assets", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}
