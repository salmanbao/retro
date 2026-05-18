package onboarding

import (
	"time"

	"github.com/google/uuid"
	domain "viralforge/backend/src/domain/onboarding"
	"viralforge/backend/src/repository/onboarding"
)

// SeedTemplates creates the initial onboarding templates for all profile types
func SeedTemplates(repo *onboarding.TemplateRepo) error {
	templates := []struct {
		profileType string
		version     string
		steps       []domain.OnboardingStep
	}{
		// Brand Template (version 1.0)
		{
			profileType: domain.ProfileTypeBrand,
			version:     "1.0",
			steps: []domain.OnboardingStep{
				{
					ID:              uuid.New(),
					Title:           "Complete company profile",
					Description:     "Add your company name, bio, and logo",
					ActionURL:       "/profiles/me/edit",
					StepType:        "profile_completion",
					Required:        true,
					DisplayOrder:    1,
					AutoCompleteKey: "profile_enrichment",
				},
				{
					ID:              uuid.New(),
					Title:           "Add payout preferences",
					Description:     "Configure how you want to receive payments",
					ActionURL:       "/profiles/me/payout",
					StepType:        "checklist",
					Required:        true,
					DisplayOrder:    2,
					AutoCompleteKey: "payout_preferences",
				},
				{
					ID:              uuid.New(),
					Title:           "Complete KYC",
					Description:     "Submit identity verification documents",
					ActionURL:       "/profiles/me/kyc",
					StepType:        "verification",
					Required:        true,
					DisplayOrder:    3,
					AutoCompleteKey: "kyc_status",
				},
				{
					ID:           uuid.New(),
					Title:        "Create first campaign",
					Description:  "Set up your first campaign to start engaging creators",
					ActionURL:    "/campaigns/new",
					StepType:     "tutorial",
					Required:     false,
					DisplayOrder: 4,
				},
			},
		},
		// Editor Template (version 1.0)
		{
			profileType: domain.ProfileTypeEditor,
			version:     "1.0",
			steps: []domain.OnboardingStep{
				{
					ID:              uuid.New(),
					Title:           "Complete public profile",
					Description:     "Add your bio, avatar, and portfolio links",
					ActionURL:       "/profiles/me/edit",
					StepType:        "profile_completion",
					Required:        true,
					DisplayOrder:    1,
					AutoCompleteKey: "profile_enrichment",
				},
				{
					ID:           uuid.New(),
					Title:        "Upload portfolio items",
					Description:  "Add at least 3 portfolio items showcasing your work",
					ActionURL:    "/profiles/me/portfolio",
					StepType:     "checklist",
					Required:     true,
					DisplayOrder: 2,
				},
				{
					ID:              uuid.New(),
					Title:           "Add payout preferences",
					Description:     "Configure how you want to receive payments",
					ActionURL:       "/profiles/me/payout",
					StepType:        "checklist",
					Required:        true,
					DisplayOrder:    3,
					AutoCompleteKey: "payout_preferences",
				},
				{
					ID:              uuid.New(),
					Title:           "Complete KYC",
					Description:     "Submit identity verification documents",
					ActionURL:       "/profiles/me/kyc",
					StepType:        "verification",
					Required:        true,
					DisplayOrder:    4,
					AutoCompleteKey: "kyc_status",
				},
			},
		},
		// Influencer Template (version 1.0)
		{
			profileType: domain.ProfileTypeInfluencer,
			version:     "1.0",
			steps: []domain.OnboardingStep{
				{
					ID:              uuid.New(),
					Title:           "Complete public profile",
					Description:     "Add your bio, avatar, and social media links",
					ActionURL:       "/profiles/me/edit",
					StepType:        "profile_completion",
					Required:        true,
					DisplayOrder:    1,
					AutoCompleteKey: "profile_enrichment",
				},
				{
					ID:              uuid.New(),
					Title:           "Add social accounts",
					Description:     "Connect your social media accounts for verification",
					ActionURL:       "/profiles/me/social",
					StepType:        "checklist",
					Required:        true,
					DisplayOrder:    2,
					AutoCompleteKey: "social_links",
				},
				{
					ID:           uuid.New(),
					Title:        "Submit follower verification",
					Description:  "Verify your follower count for campaign eligibility",
					ActionURL:    "/profiles/me/followers",
					StepType:     "verification",
					Required:     true,
					DisplayOrder: 3,
				},
				{
					ID:              uuid.New(),
					Title:           "Add payout preferences",
					Description:     "Configure how you want to receive payments",
					ActionURL:       "/profiles/me/payout",
					StepType:        "checklist",
					Required:        true,
					DisplayOrder:    4,
					AutoCompleteKey: "payout_preferences",
				},
				{
					ID:              uuid.New(),
					Title:           "Complete KYC",
					Description:     "Submit identity verification documents",
					ActionURL:       "/profiles/me/kyc",
					StepType:        "verification",
					Required:        true,
					DisplayOrder:    5,
					AutoCompleteKey: "kyc_status",
				},
			},
		},
	}

	for _, t := range templates {
		now := time.Now()
		template := &domain.OnboardingTemplate{
			ID:          uuid.New(),
			ProfileType: t.profileType,
			Version:     t.version,
			CreatedAt:   now,
			UpdatedAt:   now,
			Steps:       t.steps,
		}

		if err := repo.Create(template); err != nil {
			return err
		}
	}

	return nil
}
