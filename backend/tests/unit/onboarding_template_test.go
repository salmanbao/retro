package unit

import (
	"testing"

	domain "viralforge/backend/src/domain/onboarding"
)

func TestIsValidStepType(t *testing.T) {
	tests := []struct {
		name     string
		stepType string
		want     bool
	}{
		{"valid tutorial", "tutorial", true},
		{"valid checklist", "checklist", true},
		{"valid verification", "verification", true},
		{"valid profile_completion", "profile_completion", true},
		{"invalid empty", "", false},
		{"invalid unknown", "unknown_type", false},
		{"invalid case sensitive", "Tutorial", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.IsValidStepType(tt.stepType); got != tt.want {
				t.Errorf("IsValidStepType(%q) = %v, want %v", tt.stepType, got, tt.want)
			}
		})
	}
}

func TestOnboardingStep_Struct(t *testing.T) {
	step := domain.OnboardingStep{
		Title:           "Test Step",
		Description:     "Test Description",
		ActionURL:       "/test",
		StepType:        "checklist",
		Required:        true,
		DisplayOrder:    1,
		AutoCompleteKey: "test_key",
	}

	if step.Title != "Test Step" {
		t.Errorf("Title = %v, want Test Step", step.Title)
	}
	if step.StepType != "checklist" {
		t.Errorf("StepType = %v, want checklist", step.StepType)
	}
	if step.Required != true {
		t.Errorf("Required = %v, want true", step.Required)
	}
}

func TestProfileType_Constants(t *testing.T) {
	if domain.ProfileTypeBrand != "brand" {
		t.Errorf("ProfileTypeBrand = %v, want brand", domain.ProfileTypeBrand)
	}
	if domain.ProfileTypeEditor != "editor" {
		t.Errorf("ProfileTypeEditor = %v, want editor", domain.ProfileTypeEditor)
	}
	if domain.ProfileTypeInfluencer != "influencer" {
		t.Errorf("ProfileTypeInfluencer = %v, want influencer", domain.ProfileTypeInfluencer)
	}
}

func TestIsValidProfileType(t *testing.T) {
	tests := []struct {
		profileType string
		want        bool
	}{
		{"brand", true},
		{"editor", true},
		{"influencer", true},
		{"admin", false},
		{"", false},
		{"BRAND", false},
	}

	for _, tt := range tests {
		t.Run(tt.profileType, func(t *testing.T) {
			if got := domain.IsValidProfileType(tt.profileType); got != tt.want {
				t.Errorf("IsValidProfileType(%q) = %v, want %v", tt.profileType, got, tt.want)
			}
		})
	}
}

func TestActivationStatus_Constants(t *testing.T) {
	if domain.ActivationStatusNotStarted != "not_started" {
		t.Errorf("ActivationStatusNotStarted = %v, want not_started", domain.ActivationStatusNotStarted)
	}
	if domain.ActivationStatusOnboarding != "onboarding" {
		t.Errorf("ActivationStatusOnboarding = %v, want onboarding", domain.ActivationStatusOnboarding)
	}
	if domain.ActivationStatusPendingReview != "pending_review" {
		t.Errorf("ActivationStatusPendingReview = %v, want pending_review", domain.ActivationStatusPendingReview)
	}
	if domain.ActivationStatusActivated != "activated" {
		t.Errorf("ActivationStatusActivated = %v, want activated", domain.ActivationStatusActivated)
	}
}

func TestStepStatus_Constants(t *testing.T) {
	if domain.StepStatusNotStarted != "not_started" {
		t.Errorf("StepStatusNotStarted = %v, want not_started", domain.StepStatusNotStarted)
	}
	if domain.StepStatusInProgress != "in_progress" {
		t.Errorf("StepStatusInProgress = %v, want in_progress", domain.StepStatusInProgress)
	}
	if domain.StepStatusCompleted != "completed" {
		t.Errorf("StepStatusCompleted = %v, want completed", domain.StepStatusCompleted)
	}
	if domain.StepStatusSkipped != "skipped" {
		t.Errorf("StepStatusSkipped = %v, want skipped", domain.StepStatusSkipped)
	}
}
