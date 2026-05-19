package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// T026: Unit tests for edit restriction logic by status

func TestGetEditableFields_Draft(t *testing.T) {
	allowed, rejected := service.GetEditableFields(domain.CampaignStatusDraft)

	assert.Contains(t, allowed, "all", "draft should allow all fields")
	assert.Empty(t, rejected, "draft should not reject any fields")
}

func TestGetEditableFields_Published(t *testing.T) {
	allowed, rejected := service.GetEditableFields(domain.CampaignStatusPublished)

	assert.NotContains(t, allowed, "all", "published should not allow all fields")
	assert.Contains(t, allowed, "summary", "published should allow summary")
	assert.Contains(t, allowed, "description", "published should allow description")
	assert.Contains(t, allowed, "talking_points", "published should allow talking_points")
	assert.Contains(t, allowed, "hashtags", "published should allow hashtags")
	assert.Contains(t, allowed, "cta_instructions", "published should allow cta_instructions")

	assert.Contains(t, rejected, "title", "published should reject title")
	assert.Contains(t, rejected, "total_budget", "published should reject total_budget")
	assert.Contains(t, rejected, "submission_start", "published should reject submission_start")
}

func TestGetEditableFields_Active(t *testing.T) {
	allowed, rejected := service.GetEditableFields(domain.CampaignStatusActive)

	// Active has same restrictions as published
	assert.NotContains(t, allowed, "all", "active should not allow all fields")
	assert.Contains(t, allowed, "summary", "active should allow summary")
	assert.Contains(t, rejected, "title", "active should reject title")
	assert.Contains(t, rejected, "total_budget", "active should reject total_budget")
}

func TestGetEditableFields_Paused(t *testing.T) {
	allowed, rejected := service.GetEditableFields(domain.CampaignStatusPaused)

	// Paused allows more than published/active but still restricts some
	assert.NotContains(t, allowed, "all", "paused should not allow all fields")
	assert.NotContains(t, rejected, "total_budget", "paused should allow total_budget")
	assert.NotContains(t, rejected, "min_payout", "paused should allow min_payout")
	assert.NotContains(t, rejected, "max_payout", "paused should allow max_payout")
	assert.NotContains(t, rejected, "cpv", "paused should allow cpv")
	assert.NotContains(t, rejected, "target_clips", "paused should allow target_clips")
	assert.NotContains(t, rejected, "target_posts", "paused should allow target_posts")

	assert.Contains(t, rejected, "title", "paused should reject title")
	assert.Contains(t, rejected, "submission_start", "paused should reject submission_start")
	assert.Contains(t, rejected, "submission_deadline", "paused should reject submission_deadline")
	assert.Contains(t, rejected, "distribution_start", "paused should reject distribution_start")
	assert.Contains(t, rejected, "campaign_end", "paused should reject campaign_end")
}

func TestGetEditableFields_Completed(t *testing.T) {
	allowed, rejected := service.GetEditableFields(domain.CampaignStatusCompleted)

	assert.Nil(t, allowed, "completed should have nil allowed")
	assert.Contains(t, rejected, "all", "completed should reject all fields")
}

func TestGetEditableFields_Cancelled(t *testing.T) {
	allowed, rejected := service.GetEditableFields(domain.CampaignStatusCancelled)

	assert.Nil(t, allowed, "cancelled should have nil allowed")
	assert.Contains(t, rejected, "all", "cancelled should reject all fields")
}

func TestEditRestrictions_FieldLevel(t *testing.T) {
	tests := []struct {
		name       string
		status     domain.CampaignStatus
		field      string
		shouldAllow bool
	}{
		// Draft allows all
		{"draft title", domain.CampaignStatusDraft, "title", true},
		{"draft budget", domain.CampaignStatusDraft, "total_budget", true},
		{"draft deadline", domain.CampaignStatusDraft, "submission_deadline", true},

		// Published restricts budget-related
		{"published title", domain.CampaignStatusPublished, "title", false},
		{"published budget", domain.CampaignStatusPublished, "total_budget", false},
		{"published summary", domain.CampaignStatusPublished, "summary", true},
		{"published description", domain.CampaignStatusPublished, "description", true},

		// Active same as published
		{"active title", domain.CampaignStatusActive, "title", false},
		{"active budget", domain.CampaignStatusActive, "total_budget", false},
		{"active hashtags", domain.CampaignStatusActive, "hashtags", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _ := service.GetEditableFields(tt.status)
			canEdit := containsField(allowed, tt.field)
			assert.Equal(t, tt.shouldAllow, canEdit,
				"status=%s, field=%s: expected allowed=%v, got %v",
				tt.status, tt.field, tt.shouldAllow, canEdit)
		})
	}
}

func TestEditRestrictions_UpdateInputValidation(t *testing.T) {
	t.Run("draft can update all critical fields", func(t *testing.T) {
		input := service.UpdateInput{
			Title:       strPtr("New Title"),
			TotalBudget: float64Ptr(5000.00),
		}
		allowed, _ := service.GetEditableFields(domain.CampaignStatusDraft)

		canUpdateTitle := containsField(allowed, "title")
		canUpdateBudget := containsField(allowed, "total_budget")

		assert.True(t, canUpdateTitle, "draft should allow title update")
		assert.True(t, canUpdateBudget, "draft should allow budget update")
		_ = input // Use the input
	})

	t.Run("published cannot update title", func(t *testing.T) {
		input := service.UpdateInput{
			Title: strPtr("New Title"),
		}
		allowed, _ := service.GetEditableFields(domain.CampaignStatusPublished)

		canUpdateTitle := containsField(allowed, "title")
		assert.False(t, canUpdateTitle, "published should not allow title update")
		_ = input
	})

	t.Run("completed cannot update anything", func(t *testing.T) {
		input := service.UpdateInput{
			Summary: strPtr("New Summary"),
		}
		allowed, _ := service.GetEditableFields(domain.CampaignStatusCompleted)

		assert.Nil(t, allowed, "completed should have no allowed fields")
		_ = input
	})
}

// Helper functions

func containsField(fields []string, name string) bool {
	for _, f := range fields {
		if f == "all" || f == name {
			return true
		}
	}
	return false
}

func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}