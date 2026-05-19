package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

var (
	ErrBriefAlreadyExists  = errors.New("creative brief already exists for this campaign")
	ErrCampaignNotEditable = errors.New("campaign is not in an editable state")
	ErrAssetArchived       = errors.New("asset is archived and cannot be updated")
)

type CreativeBriefService struct {
	briefRepo    repository.CreativeBriefRepository
	campaignRepo repository.CampaignRepository
}

func NewCreativeBriefService(
	briefRepo repository.CreativeBriefRepository,
	campaignRepo repository.CampaignRepository,
) *CreativeBriefService {
	return &CreativeBriefService{
		briefRepo:    briefRepo,
		campaignRepo: campaignRepo,
	}
}

type CreateBriefInput struct {
	CampaignID             uuid.UUID
	KeyMessages            []string
	ProductBenefits        []string
	MandatoryTalkingPoints []string
	ProhibitedClaims       []string
	RequiredHashtags       []string
	CallToActionText       *string
	ToneAndStyleGuidelines *string
	TargetAudienceDesc     *string
	CompetitorReferences   []string
	ExampleVideoLinks      []string
}

func (s *CreativeBriefService) Create(ctx context.Context, input CreateBriefInput) (*domain.CreativeBrief, error) {
	existing, err := s.briefRepo.ByCampaignID(ctx, input.CampaignID)
	if err != nil && !errors.Is(err, repository.ErrBriefNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrBriefAlreadyExists
	}

	campaign, err := s.campaignRepo.ByID(ctx, input.CampaignID)
	if err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, domain.ErrCampaignNotFound
	}

	brief := &domain.CreativeBrief{
		CampaignID:             input.CampaignID,
		KeyMessages:            input.KeyMessages,
		ProductBenefits:        input.ProductBenefits,
		MandatoryTalkingPoints: input.MandatoryTalkingPoints,
		ProhibitedClaims:       input.ProhibitedClaims,
		RequiredHashtags:       input.RequiredHashtags,
		CallToActionText:       input.CallToActionText,
		ToneAndStyleGuidelines: input.ToneAndStyleGuidelines,
		TargetAudienceDesc:     input.TargetAudienceDesc,
		CompetitorReferences:   input.CompetitorReferences,
		ExampleVideoLinks:      input.ExampleVideoLinks,
	}

	if err := brief.Validate(); err != nil {
		return nil, err
	}

	if err := s.briefRepo.Create(ctx, brief); err != nil {
		return nil, err
	}

	return brief, nil
}

func (s *CreativeBriefService) GetByCampaignID(ctx context.Context, campaignID uuid.UUID) (*domain.CreativeBrief, error) {
	brief, err := s.briefRepo.ByCampaignID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, repository.ErrBriefNotFound) {
			return nil, domain.ErrBriefNotFound
		}
		return nil, err
	}
	return brief, nil
}

type UpdateBriefInput struct {
	KeyMessages            []string
	ProductBenefits        []string
	MandatoryTalkingPoints []string
	ProhibitedClaims       []string
	RequiredHashtags       []string
	CallToActionText       *string
	ToneAndStyleGuidelines *string
	TargetAudienceDesc     *string
	CompetitorReferences   []string
	ExampleVideoLinks      []string
}

func (s *CreativeBriefService) Update(ctx context.Context, brief *domain.CreativeBrief, input UpdateBriefInput) error {
	campaign, err := s.campaignRepo.ByID(ctx, brief.CampaignID)
	if err != nil {
		return err
	}
	if campaign == nil {
		return domain.ErrCampaignNotFound
	}

	if !brief.CanEditFull(campaign.Status) {
		return ErrCampaignNotEditable
	}

	brief.KeyMessages = input.KeyMessages
	brief.ProductBenefits = input.ProductBenefits
	brief.MandatoryTalkingPoints = input.MandatoryTalkingPoints
	brief.ProhibitedClaims = input.ProhibitedClaims
	brief.RequiredHashtags = input.RequiredHashtags
	brief.CallToActionText = input.CallToActionText

	if brief.CanEditFull(campaign.Status) {
		brief.ToneAndStyleGuidelines = input.ToneAndStyleGuidelines
		brief.TargetAudienceDesc = input.TargetAudienceDesc
		brief.ExampleVideoLinks = input.ExampleVideoLinks
	}

	if err := brief.Validate(); err != nil {
		return err
	}

	return s.briefRepo.Update(ctx, brief)
}
