package creativebrief

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
	briefSvc    *service.CreativeBriefService
	campaignSvc service.CampaignServiceInterface
	profileRepo interface {
		ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	}
}

func NewHandler(briefSvc *service.CreativeBriefService, campaignSvc service.CampaignServiceInterface, profileRepo interface {
	ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
}) *Handler {
	return &Handler{
		briefSvc:    briefSvc,
		campaignSvc: campaignSvc,
		profileRepo: profileRepo,
	}
}

type CreateRequest struct {
	KeyMessages            []string `json:"key_messages"`
	ProductBenefits        []string `json:"product_benefits"`
	MandatoryTalkingPoints []string `json:"mandatory_talking_points"`
	ProhibitedClaims       []string `json:"prohibited_claims"`
	RequiredHashtags       []string `json:"required_hashtags"`
	CallToActionText       *string  `json:"call_to_action_text,omitempty"`
	ToneAndStyleGuidelines *string  `json:"tone_and_style_guidelines,omitempty"`
	TargetAudienceDesc     *string  `json:"target_audience_description,omitempty"`
	CompetitorReferences   []string `json:"competitor_references"`
	ExampleVideoLinks      []string `json:"example_video_links"`
}

type UpdateRequest struct {
	KeyMessages            []string `json:"key_messages"`
	ProductBenefits        []string `json:"product_benefits"`
	MandatoryTalkingPoints []string `json:"mandatory_talking_points"`
	ProhibitedClaims       []string `json:"prohibited_claims"`
	RequiredHashtags       []string `json:"required_hashtags"`
	CallToActionText       *string  `json:"call_to_action_text,omitempty"`
	ToneAndStyleGuidelines *string  `json:"tone_and_style_guidelines,omitempty"`
	TargetAudienceDesc     *string  `json:"target_audience_description,omitempty"`
	CompetitorReferences   []string `json:"competitor_references"`
	ExampleVideoLinks      []string `json:"example_video_links"`
}

type BriefResponse struct {
	ID                     uuid.UUID `json:"id"`
	CampaignID             uuid.UUID `json:"campaign_id"`
	KeyMessages            []string  `json:"key_messages"`
	ProductBenefits        []string  `json:"product_benefits"`
	MandatoryTalkingPoints []string  `json:"mandatory_talking_points"`
	ProhibitedClaims       []string  `json:"prohibited_claims"`
	RequiredHashtags       []string  `json:"required_hashtags"`
	CallToActionText       *string   `json:"call_to_action_text,omitempty"`
	ToneAndStyleGuidelines *string   `json:"tone_and_style_guidelines,omitempty"`
	TargetAudienceDesc     *string   `json:"target_audience_description,omitempty"`
	CompetitorReferences   []string  `json:"competitor_references"`
	ExampleVideoLinks      []string  `json:"example_video_links"`
	CreatedAt              string    `json:"created_at"`
	UpdatedAt              string    `json:"updated_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := chi.URLParam(r, "campaignId")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_campaign_id", Message: "Invalid campaign ID format"})
		return
	}

	// Get brand profile ID from context (set by auth middleware)
	profileIDVal := r.Context().Value("active_profile_id")
	if profileIDVal != nil {
		profileID, ok := profileIDVal.(uuid.UUID)
		if ok {
			profile, err := h.profileRepo.ByID(r.Context(), profileID)
			if err == nil && profile.Type == domain.ProfileTypeBrand {
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
			} else if err == nil && profile.Type == domain.ProfileTypeEditor {
				// Editor - check campaign status is published or active
				campaign, err := h.campaignSvc.GetByID(r.Context(), campaignID)
				if err != nil || campaign == nil {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
					return
				}
				if campaign.Status != domain.CampaignStatusPublished && campaign.Status != domain.CampaignStatusActive {
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Editors can only access briefs for published or active campaigns"})
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

	brief, err := h.briefSvc.GetByCampaignID(r.Context(), campaignID)
	if err != nil {
		if errors.Is(err, repository.ErrBriefNotFound) || errors.Is(err, domain.ErrBriefNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "creative_brief_not_found", Message: "No creative brief found for this campaign"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to retrieve creative brief"})
		return
	}

	response := BriefResponse{
		ID:                     brief.ID,
		CampaignID:             brief.CampaignID,
		KeyMessages:            brief.KeyMessages,
		ProductBenefits:        brief.ProductBenefits,
		MandatoryTalkingPoints: brief.MandatoryTalkingPoints,
		ProhibitedClaims:       brief.ProhibitedClaims,
		RequiredHashtags:       brief.RequiredHashtags,
		CallToActionText:       brief.CallToActionText,
		ToneAndStyleGuidelines: brief.ToneAndStyleGuidelines,
		TargetAudienceDesc:     brief.TargetAudienceDesc,
		CompetitorReferences:   brief.CompetitorReferences,
		ExampleVideoLinks:      brief.ExampleVideoLinks,
		CreatedAt:              brief.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:              brief.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
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

	// Get campaign to verify ownership
	campaign, err := h.campaignSvc.GetByID(r.Context(), campaignID)
	if err != nil || campaign == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
		return
	}

	// Only brand owners can create/update briefs
	if profile.Type != domain.ProfileTypeBrand {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "Only brand profiles can manage creative briefs"})
		return
	}

	if campaign.BrandProfileID != profileID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "forbidden", Message: "You do not own this campaign"})
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid_request", Message: "Invalid request body"})
		return
	}

	input := service.CreateBriefInput{
		CampaignID:             campaignID,
		KeyMessages:            req.KeyMessages,
		ProductBenefits:        req.ProductBenefits,
		MandatoryTalkingPoints: req.MandatoryTalkingPoints,
		ProhibitedClaims:       req.ProhibitedClaims,
		RequiredHashtags:       req.RequiredHashtags,
		CallToActionText:       req.CallToActionText,
		ToneAndStyleGuidelines: req.ToneAndStyleGuidelines,
		TargetAudienceDesc:     req.TargetAudienceDesc,
		CompetitorReferences:   req.CompetitorReferences,
		ExampleVideoLinks:      req.ExampleVideoLinks,
	}

	brief, err := h.briefSvc.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrBriefAlreadyExists) {
			existing, getErr := h.briefSvc.GetByCampaignID(r.Context(), campaignID)
			if getErr != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "conflict", Message: "Creative brief already exists for this campaign"})
				return
			}

			updateInput := service.UpdateBriefInput{
				KeyMessages:            req.KeyMessages,
				ProductBenefits:        req.ProductBenefits,
				MandatoryTalkingPoints: req.MandatoryTalkingPoints,
				ProhibitedClaims:       req.ProhibitedClaims,
				RequiredHashtags:       req.RequiredHashtags,
				CallToActionText:       req.CallToActionText,
				ToneAndStyleGuidelines: req.ToneAndStyleGuidelines,
				TargetAudienceDesc:     req.TargetAudienceDesc,
				CompetitorReferences:   req.CompetitorReferences,
				ExampleVideoLinks:      req.ExampleVideoLinks,
			}

			if updateErr := h.briefSvc.Update(r.Context(), existing, updateInput); updateErr != nil {
				if errors.Is(updateErr, service.ErrCampaignNotEditable) {
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_editable", Message: "Cannot modify brief for published or active campaign"})
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "internal_error", Message: "Failed to update creative brief"})
				return
			}

			response := BriefResponse{
				ID:                     existing.ID,
				CampaignID:             existing.CampaignID,
				KeyMessages:            existing.KeyMessages,
				ProductBenefits:        existing.ProductBenefits,
				MandatoryTalkingPoints: existing.MandatoryTalkingPoints,
				ProhibitedClaims:       existing.ProhibitedClaims,
				RequiredHashtags:       existing.RequiredHashtags,
				CallToActionText:       existing.CallToActionText,
				ToneAndStyleGuidelines: existing.ToneAndStyleGuidelines,
				TargetAudienceDesc:     existing.TargetAudienceDesc,
				CompetitorReferences:   existing.CompetitorReferences,
				ExampleVideoLinks:      existing.ExampleVideoLinks,
				CreatedAt:              existing.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:              existing.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		if errors.Is(err, domain.ErrCampaignNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "campaign_not_found", Message: "Campaign not found"})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "validation_failed", Message: err.Error()})
		return
	}

	response := BriefResponse{
		ID:                     brief.ID,
		CampaignID:             brief.CampaignID,
		KeyMessages:            brief.KeyMessages,
		ProductBenefits:        brief.ProductBenefits,
		MandatoryTalkingPoints: brief.MandatoryTalkingPoints,
		ProhibitedClaims:       brief.ProhibitedClaims,
		RequiredHashtags:       brief.RequiredHashtags,
		CallToActionText:       brief.CallToActionText,
		ToneAndStyleGuidelines: brief.ToneAndStyleGuidelines,
		TargetAudienceDesc:     brief.TargetAudienceDesc,
		CompetitorReferences:   brief.CompetitorReferences,
		ExampleVideoLinks:      brief.ExampleVideoLinks,
		CreatedAt:              brief.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:              brief.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/{campaignId}/brief", h.Get)
	r.Put("/{campaignId}/brief", h.Put)
}
