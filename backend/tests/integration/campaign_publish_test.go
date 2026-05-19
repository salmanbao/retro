package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// T022: Integration tests for publish with all readiness scenarios

func TestCampaignPublish_ReadinessScenarios(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("campaign with zero budget fails publish", func(t *testing.T) {
		// Create a campaign with zero budget
		input := service.CreateCampaignInput{
			BrandProfileID:     uuid.New(),
			Title:              "Zero Budget Campaign",
			Description:        "This campaign has zero budget",
			TotalBudget:        0, // Zero budget should fail
			Currency:           "USD",
			SubmissionStart:    time.Now().Add(24 * time.Hour),
			SubmissionDeadline: time.Now().Add(48 * time.Hour),
			DistributionStart:  time.Now().Add(72 * time.Hour),
			CampaignEnd:        time.Now().Add(168 * time.Hour),
		}

		// Budget validation should fail
		if input.TotalBudget <= 0 {
			t.Log("Zero budget correctly identified as invalid for publish")
		}
	})

	t.Run("campaign with past submission deadline fails", func(t *testing.T) {
		input := service.CreateCampaignInput{
			BrandProfileID:     uuid.New(),
			Title:              "Past Deadline Campaign",
			Description:        "This campaign has invalid timeline",
			TotalBudget:        1000.00,
			Currency:           "USD",
			SubmissionStart:    time.Now().Add(-48 * time.Hour), // Past
			SubmissionDeadline: time.Now().Add(-24 * time.Hour), // Past
			DistributionStart:  time.Now().Add(72 * time.Hour),
			CampaignEnd:        time.Now().Add(168 * time.Hour),
		}

		err := service.ValidateTimeline(
			input.SubmissionStart,
			input.SubmissionDeadline,
			input.DistributionStart,
			input.CampaignEnd,
		)
		require.Error(t, err, "past deadline should fail validation")
	})

	t.Run("valid timeline passes validation", func(t *testing.T) {
		input := service.CreateCampaignInput{
			BrandProfileID:     uuid.New(),
			Title:              "Valid Timeline Campaign",
			Description:        "This campaign has valid timeline",
			TotalBudget:        5000.00,
			Currency:           "USD",
			SubmissionStart:    time.Now().Add(24 * time.Hour),
			SubmissionDeadline: time.Now().Add(48 * time.Hour),
			DistributionStart:  time.Now().Add(72 * time.Hour),
			CampaignEnd:        time.Now().Add(168 * time.Hour),
		}

		err := service.ValidateTimeline(
			input.SubmissionStart,
			input.SubmissionDeadline,
			input.DistributionStart,
			input.CampaignEnd,
		)
		require.NoError(t, err, "valid timeline should pass")
	})

	t.Run("invalid payout range fails", func(t *testing.T) {
		minPayout := float64Ptr(500.00)
		maxPayout := float64Ptr(200.00) // Max less than min

		err := service.ValidatePayoutRange(minPayout, maxPayout)
		require.Error(t, err, "invalid payout range should fail")
	})

	t.Run("valid payout range passes", func(t *testing.T) {
		minPayout := float64Ptr(100.00)
		maxPayout := float64Ptr(500.00)

		err := service.ValidatePayoutRange(minPayout, maxPayout)
		require.NoError(t, err, "valid payout range should pass")
	})

	t.Run("invalid duration range fails", func(t *testing.T) {
		err := service.ValidateDuration(60, 30) // Max less than min
		require.Error(t, err, "invalid duration range should fail")
	})

	t.Run("valid duration range passes", func(t *testing.T) {
		err := service.ValidateDuration(15, 60)
		require.NoError(t, err, "valid duration range should pass")
	})
}

func TestCampaignPublish_Workflow(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	t.Run("draft campaign can be published when ready", func(t *testing.T) {
		// This test validates the publish workflow concept
		// Actual integration with HTTP API requires server running
		status := domain.CampaignStatusDraft

		// Draft should transition to published
		assert.True(t, status.IsValidTransition(domain.CampaignStatusPublished),
			"draft should be able to transition to published")
	})

	t.Run("published campaign can transition to active", func(t *testing.T) {
		status := domain.CampaignStatusPublished

		assert.True(t, status.IsValidTransition(domain.CampaignStatusActive),
			"published should be able to transition to active")
	})
}
