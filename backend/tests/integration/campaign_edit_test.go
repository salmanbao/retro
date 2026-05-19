package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// T027: Integration tests for restricted edits on published/active campaigns

func TestCampaignEdit_RestrictedEdits(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("draft can update all allowed fields", func(t *testing.T) {
		status := domain.CampaignStatusDraft
		allowed, _ := service.GetEditableFields(status)

		// Draft should allow all fields
		require.Contains(t, allowed, "all", "draft should allow all fields")
	})

	t.Run("published cannot update title", func(t *testing.T) {
		status := domain.CampaignStatusPublished
		allowed, rejected := service.GetEditableFields(status)

		require.NotContains(t, allowed, "title", "published should not allow title")
		require.Contains(t, rejected, "title", "published should reject title")
	})

	t.Run("published cannot update budget", func(t *testing.T) {
		status := domain.CampaignStatusPublished
		allowed, rejected := service.GetEditableFields(status)

		require.NotContains(t, allowed, "total_budget", "published should not allow total_budget")
		require.Contains(t, rejected, "total_budget", "published should reject total_budget")
	})

	t.Run("published can update description", func(t *testing.T) {
		status := domain.CampaignStatusPublished
		allowed, _ := service.GetEditableFields(status)

		require.Contains(t, allowed, "description", "published should allow description")
	})

	t.Run("active cannot update deadline fields", func(t *testing.T) {
		status := domain.CampaignStatusActive
		allowed, rejected := service.GetEditableFields(status)

		require.NotContains(t, allowed, "submission_start", "active should not allow submission_start")
		require.NotContains(t, allowed, "submission_deadline", "active should not allow submission_deadline")
		require.Contains(t, rejected, "submission_start", "active should reject submission_start")
	})

	t.Run("completed cannot update any field", func(t *testing.T) {
		status := domain.CampaignStatusCompleted
		allowed, rejected := service.GetEditableFields(status)

		require.Nil(t, allowed, "completed should have no allowed fields")
		require.Contains(t, rejected, "all", "completed should reject all")
	})

	t.Run("cancelled cannot update any field", func(t *testing.T) {
		status := domain.CampaignStatusCancelled
		allowed, rejected := service.GetEditableFields(status)

		require.Nil(t, allowed, "cancelled should have no allowed fields")
		require.Contains(t, rejected, "all", "cancelled should reject all")
	})
}

func TestCampaignEdit_FieldRestrictions(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("paused allows more edits than active", func(t *testing.T) {
		activeAllowed, _ := service.GetEditableFields(domain.CampaignStatusActive)
		pausedAllowed, _ := service.GetEditableFields(domain.CampaignStatusPaused)

		// Paused allows total_budget but active doesn't
		require.NotContains(t, activeAllowed, "total_budget", "active should not allow total_budget")
		require.Contains(t, pausedAllowed, "total_budget", "paused should allow total_budget")
	})

	t.Run("paused allows payout changes", func(t *testing.T) {
		pausedAllowed, _ := service.GetEditableFields(domain.CampaignStatusPaused)

		require.Contains(t, pausedAllowed, "min_payout", "paused should allow min_payout")
		require.Contains(t, pausedAllowed, "max_payout", "paused should allow max_payout")
	})

	t.Run("paused still rejects timeline fields", func(t *testing.T) {
		pausedAllowed, rejected := service.GetEditableFields(domain.CampaignStatusPaused)

		require.NotContains(t, pausedAllowed, "submission_start", "paused should not allow submission_start")
		require.NotContains(t, pausedAllowed, "submission_deadline", "paused should not allow submission_deadline")
		require.Contains(t, rejected, "submission_start", "paused should reject submission_start")
	})
}

func TestCampaignEdit_UpdateValidation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("valid update input for draft", func(t *testing.T) {
		input := service.UpdateInput{
			BrandProfileID: uuid.New(),
			CampaignID:     uuid.New(),
			Title:          strPtr("Updated Title"),
			TotalBudget:    float64Ptr(10000.00),
		}

		allowed, _ := service.GetEditableFields(domain.CampaignStatusDraft)
		require.Contains(t, allowed, "all", "draft should allow all fields")

		_ = input
	})

	t.Run("update input validation for published campaign", func(t *testing.T) {
		input := service.UpdateInput{
			CampaignID: uuid.New(),
			Title:      strPtr("Should Fail"), // Should not be allowed
		}

		allowed, _ := service.GetEditableFields(domain.CampaignStatusPublished)
		canUpdateTitle := containsField(allowed, "title")
		require.False(t, canUpdateTitle, "published campaign should not allow title update")

		_ = input
	})
}