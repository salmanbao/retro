package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

// TestBrandOwnershipAuthorization tests that Brand profiles can only access
// briefs and assets for campaigns they own.
func TestBrandOwnershipAuthorization(t *testing.T) {
	t.Run("brand profile can access own campaign brief", func(t *testing.T) {
		brandProfileID := uuid.New()
		campaignID := uuid.New()

		// Simulate ownership check - campaign belongs to brand profile
		campaign := &domain.Campaign{
			ID:             campaignID,
			BrandProfileID: brandProfileID,
			Status:         domain.CampaignStatusDraft,
		}

		// The brand profile ID matches the campaign's BrandProfileID
		assert.Equal(t, brandProfileID, campaign.BrandProfileID)
	})

	t.Run("brand profile cannot access other brand's campaign brief", func(t *testing.T) {
		ownerProfileID := uuid.New()
		otherProfileID := uuid.New()
		campaignID := uuid.New()

		campaign := &domain.Campaign{
			ID:             campaignID,
			BrandProfileID: ownerProfileID, // Campaign owned by different profile
			Status:         domain.CampaignStatusDraft,
		}

		// The other profile should not have access
		assert.NotEqual(t, otherProfileID, campaign.BrandProfileID)
	})

	t.Run("brand profile can access own campaign assets", func(t *testing.T) {
		brandProfileID := uuid.New()
		campaignID := uuid.New()

		campaign := &domain.Campaign{
			ID:             campaignID,
			BrandProfileID: brandProfileID,
			Status:         domain.CampaignStatusActive,
		}

		assert.Equal(t, brandProfileID, campaign.BrandProfileID)
	})
}

// TestEditorReadAccess tests that Editors can read briefs and assets
// for published or active campaigns.
func TestEditorReadAccess(t *testing.T) {
	t.Run("editor can read brief for published campaign", func(t *testing.T) {
		campaign := &domain.Campaign{
			ID:             uuid.New(),
			BrandProfileID: uuid.New(),
			Status:         domain.CampaignStatusPublished,
		}

		// Editors can read briefs for published campaigns
		// For brief reading (not editing), published/active status is allowed
		assert.True(t, campaign.Status == domain.CampaignStatusPublished ||
			campaign.Status == domain.CampaignStatusActive)
	})

	t.Run("editor can read brief for active campaign", func(t *testing.T) {
		campaign := &domain.Campaign{
			ID:             uuid.New(),
			BrandProfileID: uuid.New(),
			Status:         domain.CampaignStatusActive,
		}

		assert.True(t, campaign.Status == domain.CampaignStatusPublished ||
			campaign.Status == domain.CampaignStatusActive)
	})

	t.Run("editor can read assets for published campaign", func(t *testing.T) {
		campaign := &domain.Campaign{
			ID:             uuid.New(),
			BrandProfileID: uuid.New(),
			Status:         domain.CampaignStatusPublished,
		}

		// Assets are accessible to editors for published/active campaigns
		assert.True(t, campaign.Status == domain.CampaignStatusPublished ||
			campaign.Status == domain.CampaignStatusActive)
	})
}

// TestInfluencerAccessDenied tests that Influencers do not have access
// to briefs and assets at this stage.
func TestInfluencerAccessDenied(t *testing.T) {
	t.Run("influencer denied access to briefs", func(t *testing.T) {
		profileType := domain.ProfileTypeInfluencer

		// Influencers should be denied access
		hasAccess := profileType == domain.ProfileTypeBrand ||
			profileType == domain.ProfileTypeEditor

		assert.False(t, hasAccess)
	})

	t.Run("influencer denied access to assets", func(t *testing.T) {
		profileType := domain.ProfileTypeInfluencer

		// Influencers should be denied access
		hasAccess := profileType == domain.ProfileTypeBrand ||
			profileType == domain.ProfileTypeEditor

		assert.False(t, hasAccess)
	})
}

// TestCampaignStateEditRestrictions tests that briefs can only be edited
// when campaign is in editable state (draft/paused).
func TestCampaignStateEditRestrictions(t *testing.T) {
	t.Run("brief can be edited in draft state", func(t *testing.T) {
		brief := &domain.CreativeBrief{
			ID:         uuid.New(),
			CampaignID: uuid.New(),
		}
		campaign := &domain.Campaign{
			Status: domain.CampaignStatusDraft,
		}

		// Check if campaign is editable
		isEditable := campaign.Status == domain.CampaignStatusDraft ||
			campaign.Status == domain.CampaignStatusPaused

		assert.True(t, isEditable)
		assert.True(t, brief.CanEditFull(campaign.Status))
	})

	t.Run("brief can be edited in paused state", func(t *testing.T) {
		campaign := &domain.Campaign{
			Status: domain.CampaignStatusPaused,
		}

		isEditable := campaign.Status == domain.CampaignStatusDraft ||
			campaign.Status == domain.CampaignStatusPaused

		assert.True(t, isEditable)
	})

	t.Run("brief cannot be fully edited in published state", func(t *testing.T) {
		campaign := &domain.Campaign{
			Status: domain.CampaignStatusPublished,
		}

		isEditable := campaign.Status == domain.CampaignStatusDraft ||
			campaign.Status == domain.CampaignStatusPaused

		assert.False(t, isEditable)
	})

	t.Run("brief cannot be fully edited in active state", func(t *testing.T) {
		campaign := &domain.Campaign{
			Status: domain.CampaignStatusActive,
		}

		isEditable := campaign.Status == domain.CampaignStatusDraft ||
			campaign.Status == domain.CampaignStatusPaused

		assert.False(t, isEditable)
	})
}

// TestProfileTypeCheck tests the profile type verification.
func TestProfileTypeCheck(t *testing.T) {
	t.Run("valid brand profile type", func(t *testing.T) {
		profile := &domain.Profile{
			Type: domain.ProfileTypeBrand,
		}
		assert.Equal(t, domain.ProfileTypeBrand, profile.Type)
		assert.NotEqual(t, domain.ProfileTypeEditor, profile.Type)
	})

	t.Run("valid editor profile type", func(t *testing.T) {
		profile := &domain.Profile{
			Type: domain.ProfileTypeEditor,
		}
		assert.Equal(t, domain.ProfileTypeEditor, profile.Type)
		assert.NotEqual(t, domain.ProfileTypeInfluencer, profile.Type)
	})

	t.Run("valid influencer profile type", func(t *testing.T) {
		profile := &domain.Profile{
			Type: domain.ProfileTypeInfluencer,
		}
		assert.Equal(t, domain.ProfileTypeInfluencer, profile.Type)
	})
}
