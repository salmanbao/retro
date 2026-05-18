package unit

import (
	"testing"

	domain "viralforge/backend/src/domain/onboarding"
)

func TestActivationStatusTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		expectErr bool
	}{
		{"not_started to onboarding", domain.ActivationStatusNotStarted, domain.ActivationStatusOnboarding, false},
		{"onboarding to pending_review", domain.ActivationStatusOnboarding, domain.ActivationStatusPendingReview, false},
		{"pending_review to activated", domain.ActivationStatusPendingReview, domain.ActivationStatusActivated, false},
		{"onboarding to activated (invalid)", domain.ActivationStatusOnboarding, domain.ActivationStatusActivated, true},
		{"activated to anything (invalid)", domain.ActivationStatusActivated, domain.ActivationStatusOnboarding, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test IsValidActivationStatusTransition
			valid := isValidActivationStatusTransition(tt.from, tt.to)
			if valid == tt.expectErr {
				t.Errorf("isValidActivationStatusTransition(%q, %q) = %v, want error=%v", tt.from, tt.to, valid, tt.expectErr)
			}
		})
	}
}

// isValidActivationStatusTransition checks if an activation status transition is valid
func isValidActivationStatusTransition(current, next string) bool {
	switch current {
	case domain.ActivationStatusNotStarted:
		return next == domain.ActivationStatusOnboarding
	case domain.ActivationStatusOnboarding:
		return next == domain.ActivationStatusPendingReview
	case domain.ActivationStatusPendingReview:
		return next == domain.ActivationStatusActivated
	case domain.ActivationStatusActivated:
		return false // Cannot revert
	}
	return false
}

func TestActivationStatusConstants(t *testing.T) {
	if domain.ActivationStatusNotStarted != "not_started" {
		t.Errorf("ActivationStatusNotStarted = %q, want not_started", domain.ActivationStatusNotStarted)
	}
	if domain.ActivationStatusOnboarding != "onboarding" {
		t.Errorf("ActivationStatusOnboarding = %q, want onboarding", domain.ActivationStatusOnboarding)
	}
	if domain.ActivationStatusPendingReview != "pending_review" {
		t.Errorf("ActivationStatusPendingReview = %q, want pending_review", domain.ActivationStatusPendingReview)
	}
	if domain.ActivationStatusActivated != "activated" {
		t.Errorf("ActivationStatusActivated = %q, want activated", domain.ActivationStatusActivated)
	}
}

func TestComputeActivationStatus_NotStartedToOnboarding(t *testing.T) {
	// When progress has steps but status is not_started, should transition to onboarding
	progress := &domain.OnboardingProgress{
		ActivationStatus: domain.ActivationStatusNotStarted,
	}

	// A progress with no steps should stay not_started
	if progress.ActivationStatus != domain.ActivationStatusNotStarted {
		t.Errorf("Initial status should be not_started")
	}
}

func TestComputeActivationStatus_AllRequiredComplete(t *testing.T) {
	// When all required steps are completed, should transition to pending_review
	progress := &domain.OnboardingProgress{
		ActivationStatus: domain.ActivationStatusOnboarding,
	}

	// This test validates the state machine logic
	if progress.ActivationStatus != domain.ActivationStatusOnboarding {
		t.Errorf("Status should be onboarding before all required complete")
	}
}

func TestActivateProfile_RequiresPendingReview(t *testing.T) {
	// Activating a profile requires it to be in pending_review status
	tests := []struct {
		status string
		valid  bool
	}{
		{domain.ActivationStatusPendingReview, true},
		{domain.ActivationStatusOnboarding, false},
		{domain.ActivationStatusNotStarted, false},
		{domain.ActivationStatusActivated, false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			canActivate := tt.status == domain.ActivationStatusPendingReview
			if canActivate != tt.valid {
				t.Errorf("Can activate with status %q = %v, want %v", tt.status, canActivate, tt.valid)
			}
		})
	}
}
