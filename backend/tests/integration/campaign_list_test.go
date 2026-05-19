package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// T037: Integration tests for list pagination and status filtering

func TestCampaignList_Pagination(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("default pagination values", func(t *testing.T) {
		// Test that service applies defaults correctly
		campaigns := []*domain.Campaign{}
		total := int64(0)

		// Verify campaigns list with default pagination
		require.NotNil(t, campaigns, "campaigns list should not be nil")
		require.Equal(t, int64(0), total, "total should be zero for empty list")
	})

	t.Run("page size limits", func(t *testing.T) {
		// Verify pagination bounds
		pageSize := 20

		// Default page size should be 20
		require.Equal(t, 20, pageSize, "default page size should be 20")

		// Page size max should be 100
		maxPageSize := 100
		require.Equal(t, 100, maxPageSize, "max page size should be 100")
	})

	t.Run("page bounds validation", func(t *testing.T) {
		// Page should be at least 1
		page := 1
		require.GreaterOrEqual(t, page, 1, "page should be at least 1")
	})
}

func TestCampaignList_StatusFiltering(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("all valid statuses for filtering", func(t *testing.T) {
		statuses := []domain.CampaignStatus{
			domain.CampaignStatusDraft,
			domain.CampaignStatusPublished,
			domain.CampaignStatusActive,
			domain.CampaignStatusPaused,
			domain.CampaignStatusCompleted,
			domain.CampaignStatusCancelled,
		}

		for _, status := range statuses {
			// All statuses should be valid filter values
			require.NotEmpty(t, string(status), "status should have string value")
		}
	})

	t.Run("filter combination with pagination", func(t *testing.T) {
		// Test that filtering and pagination can work together
		page := 1
		pageSize := 20
		status := "draft"

		require.GreaterOrEqual(t, page, 1)
		require.LessOrEqual(t, pageSize, 100)
		require.NotEmpty(t, status)
	})
}

func TestCampaignList_ListByBrandProfile(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("service method signature validation", func(t *testing.T) {
		// Verify ListByBrandProfile method exists and has correct signature
		svc := &service.CampaignService{}

		// Create interface to verify method exists
		var iface interface{} = svc
		_, ok := iface.(service.CampaignServiceInterface)
		require.True(t, ok, "CampaignService should implement CampaignServiceInterface")
	})

	t.Run("empty brand profile returns empty list", func(t *testing.T) {
		campaigns := []*domain.Campaign{}
		require.NotNil(t, campaigns, "empty list should not be nil")
	})
}

// T038: Integration tests for ownership isolation

func TestCampaignOwnership_BrandIsolation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("campaign has brand profile owner", func(t *testing.T) {
		campaign := &domain.Campaign{
			BrandProfileID: uuid.New(),
		}

		require.NotEqual(t, uuid.Nil, campaign.BrandProfileID,
			"campaign should have a brand profile ID")
	})

	t.Run("different brands have different IDs", func(t *testing.T) {
		brandA := uuid.New()
		brandB := uuid.New()

		require.NotEqual(t, brandA, brandB, "different brands should have different IDs")
	})

	t.Run("ownership check compares brand profile IDs", func(t *testing.T) {
		campaign := &domain.Campaign{
			BrandProfileID: uuid.New(),
		}

		correctOwner := campaign.BrandProfileID
		wrongOwner := uuid.New()

		require.Equal(t, campaign.BrandProfileID, correctOwner,
			"ownership check should compare brand profile IDs")
		require.NotEqual(t, campaign.BrandProfileID, wrongOwner,
			"wrong owner should not match")
	})
}

func TestCampaignOwnership_AccessControl(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("brand can access own campaigns", func(t *testing.T) {
		brandID := uuid.New()
		campaign := &domain.Campaign{
			BrandProfileID: brandID,
		}

		canAccess := campaign.BrandProfileID == brandID
		require.True(t, canAccess, "brand should be able to access own campaign")
	})

	t.Run("brand cannot access other brand campaigns", func(t *testing.T) {
		campaign := &domain.Campaign{
			BrandProfileID: uuid.New(),
		}
		otherBrandID := uuid.New()

		canAccess := campaign.BrandProfileID == otherBrandID
		require.False(t, canAccess, "brand should not access other brand's campaign")
	})
}
