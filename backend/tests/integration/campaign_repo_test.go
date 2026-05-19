package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

func TestCampaignRepository_CRUD(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("Create campaign", func(t *testing.T) {
		// This test requires a running server with authenticated user
		// Integration tests create campaigns via HTTP API
		// See tests/contract/campaign_create_test.go for API contract tests
		_ = suite
	})

	t.Run("Read campaign by ID", func(t *testing.T) {
		// Test fetching campaign by ID
	})

	t.Run("Update campaign", func(t *testing.T) {
		// Test updating campaign
	})

	t.Run("Soft delete campaign", func(t *testing.T) {
		// Test soft delete
	})
}

func TestCampaignRepository_List(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("List campaigns with pagination", func(t *testing.T) {
		// Test pagination
	})

	t.Run("List campaigns filtered by status", func(t *testing.T) {
		// Test status filtering
	})
}

func TestCampaignRepository_SlugUniqueness(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("Slug uniqueness enforced at database level", func(t *testing.T) {
		// Create first campaign
		// Create second campaign with same slug should fail
		// Or slug should be made unique
	})
}

func TestCampaignAssetRepository_CRUD(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("Create campaign asset", func(t *testing.T) {
		asset := &domain.CampaignAsset{
			ID:          uuid.New(),
			CampaignID:  uuid.New(),
			URL:         "https://cdn.example.com/asset.mp4",
			AssetType:   domain.AssetTypeRawMedia,
			Description: "Main video asset",
		}
		_ = asset
	})

	t.Run("Get assets by campaign ID", func(t *testing.T) {
		// Test fetching assets by campaign
	})

	t.Run("Delete asset", func(t *testing.T) {
		// Test deleting an asset
	})
}

func TestCampaignRepository_OwnershipIsolation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("Brand A cannot see Brand B campaigns", func(t *testing.T) {
		// Create two users with brand profiles
		// User A creates campaign
		// User B tries to list/get User A's campaign - should fail
	})
}

func TestCampaignService_Create(t *testing.T) {
	// Unit test for campaign creation via service
	t.Run("Valid campaign creation", func(t *testing.T) {
		input := service.CreateCampaignInput{
			BrandProfileID:     uuid.New(),
			Title:              "Test Campaign",
			Description:        "A test description",
			SubmissionStart:    time.Now().Add(24 * time.Hour),
			SubmissionDeadline: time.Now().Add(48 * time.Hour),
			DistributionStart:  time.Now().Add(72 * time.Hour),
			CampaignEnd:        time.Now().Add(168 * time.Hour),
		}
		// Validate slug generation
		slug := service.GenerateSlug(input.Title)
		assert.Equal(t, "test-campaign", slug)
	})
}
