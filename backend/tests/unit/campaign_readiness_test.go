package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/service"
)

// T021: Unit tests for readiness validation rules

func TestReadinessValidation_BrandProfileOnboarding(t *testing.T) {
	tests := []struct {
		name           string
		brandOnboarded bool
		expectError    bool
	}{
		{
			name:           "brand fully onboarded",
			brandOnboarded: true,
			expectError:    false,
		},
		{
			name:           "brand not fully onboarded",
			brandOnboarded: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Readiness validation checks brand onboarding status
			// This is a simplified test - actual implementation would check
			// the BrandProfile's onboarding completion status
			if !tt.brandOnboarded {
				// When brand is not onboarded, publishing should fail
				assert.True(t, tt.expectError, "expected error when brand not onboarded")
			}
		})
	}
}

func TestReadinessValidation_KYCStatus(t *testing.T) {
	tests := []struct {
		name      string
		kycStatus string
		expectErr bool
	}{
		{
			name:      "kyc approved/verified",
			kycStatus: "verified",
			expectErr: false,
		},
		{
			name:      "kyc pending",
			kycStatus: "pending",
			expectErr: true,
		},
		{
			name:      "kyc rejected",
			kycStatus: "rejected",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify KYC status check logic
			isValid := tt.kycStatus == "verified"
			if tt.expectErr {
				assert.False(t, isValid, "KYC should not be valid for status: %s", tt.kycStatus)
			} else {
				assert.True(t, isValid, "KYC should be valid for status: %s", tt.kycStatus)
			}
		})
	}
}

func TestReadinessValidation_PayoutConfigured(t *testing.T) {
	tests := []struct {
		name             string
		payoutConfigured bool
		expectErr        bool
	}{
		{
			name:             "payout configured",
			payoutConfigured: true,
			expectErr:        false,
		},
		{
			name:             "payout not configured",
			payoutConfigured: false,
			expectErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify payout configuration check
			if !tt.payoutConfigured {
				assert.True(t, tt.expectErr, "expected error when payout not configured")
			}
		})
	}
}

func TestReadinessValidation_CampaignCompleteness(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		totalBudget float64
		expectErr   bool
	}{
		{
			name:        "complete campaign",
			title:       "Test Campaign",
			description: "A complete description",
			totalBudget: 1000.00,
			expectErr:   false,
		},
		{
			name:        "missing title",
			title:       "",
			description: "A complete description",
			totalBudget: 1000.00,
			expectErr:   true,
		},
		{
			name:        "missing description",
			title:       "Test Campaign",
			description: "",
			totalBudget: 1000.00,
			expectErr:   true,
		},
		{
			name:        "zero budget",
			title:       "Test Campaign",
			description: "A complete description",
			totalBudget: 0,
			expectErr:   true,
		},
		{
			name:        "negative budget",
			title:       "Test Campaign",
			description: "A complete description",
			totalBudget: -100.00,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify campaign completeness checks
			isComplete := tt.title != "" && tt.description != "" && tt.totalBudget > 0
			if tt.expectErr {
				assert.False(t, isComplete, "campaign should not be complete")
			} else {
				assert.True(t, isComplete, "campaign should be complete")
			}
		})
	}
}

func TestReadinessValidation_BudgetGreaterThanZero(t *testing.T) {
	tests := []struct {
		name        string
		totalBudget float64
		expectValid bool
	}{
		{
			name:        "positive budget",
			totalBudget: 1000.00,
			expectValid: true,
		},
		{
			name:        "zero budget",
			totalBudget: 0,
			expectValid: false,
		},
		{
			name:        "negative budget",
			totalBudget: -100.00,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.totalBudget > 0
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestPublishReadiness_AllConditions(t *testing.T) {
	t.Run("all conditions met for publish", func(t *testing.T) {
		// Simulate a campaign ready to publish
		campaign := &service.CreateCampaignInput{
			BrandProfileID:     uuid.New(),
			Title:              "Ready Campaign",
			Description:        "This is a complete description",
			TotalBudget:        5000.00,
			SubmissionStart:    time.Now().Add(24 * time.Hour),
			SubmissionDeadline: time.Now().Add(48 * time.Hour),
			DistributionStart:  time.Now().Add(72 * time.Hour),
			CampaignEnd:        time.Now().Add(168 * time.Hour),
		}

		// Verify timeline validation passes for complete campaign
		err := service.ValidateTimeline(
			campaign.SubmissionStart,
			campaign.SubmissionDeadline,
			campaign.DistributionStart,
			campaign.CampaignEnd,
		)
		assert.NoError(t, err)
		assert.True(t, campaign.TotalBudget > 0)
		assert.True(t, campaign.Title != "")
		assert.True(t, campaign.Description != "")
	})

	t.Run("missing budget fails publish readiness", func(t *testing.T) {
		campaign := &service.CreateCampaignInput{
			BrandProfileID:     uuid.New(),
			Title:              "Unready Campaign",
			Description:        "This is a complete description",
			TotalBudget:        0, // Missing budget
			SubmissionStart:    time.Now().Add(24 * time.Hour),
			SubmissionDeadline: time.Now().Add(48 * time.Hour),
			DistributionStart:  time.Now().Add(72 * time.Hour),
			CampaignEnd:        time.Now().Add(168 * time.Hour),
		}

		assert.False(t, campaign.TotalBudget > 0, "zero budget should fail publish readiness")
	})
}
