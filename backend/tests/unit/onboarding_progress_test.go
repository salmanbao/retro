package unit

import (
	"testing"

	"github.com/google/uuid"
	domain "viralforge/backend/src/domain/onboarding"
)

func TestIsValidStepStatusTransition(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		to     string
		expect bool
	}{
		{"not_started to in_progress", domain.StepStatusNotStarted, domain.StepStatusInProgress, true},
		{"not_started to completed", domain.StepStatusNotStarted, domain.StepStatusCompleted, false},
		{"not_started to skipped", domain.StepStatusNotStarted, domain.StepStatusSkipped, false},
		{"in_progress to completed", domain.StepStatusInProgress, domain.StepStatusCompleted, true},
		{"in_progress to skipped", domain.StepStatusInProgress, domain.StepStatusSkipped, true},
		{"in_progress to not_started", domain.StepStatusInProgress, domain.StepStatusNotStarted, false},
		{"completed to anything", domain.StepStatusCompleted, domain.StepStatusNotStarted, false},
		{"skipped to anything", domain.StepStatusSkipped, domain.StepStatusNotStarted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.IsValidStepStatusTransition(tt.from, tt.to)
			if got != tt.expect {
				t.Errorf("IsValidStepStatusTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.expect)
			}
		})
	}
}

func TestOnboardingProgress_ActivationStatus(t *testing.T) {
	progress := domain.OnboardingProgress{
		ID:               uuid.New(),
		ProfileID:        uuid.New(),
		ActivationStatus: domain.ActivationStatusNotStarted,
	}

	if progress.ActivationStatus != domain.ActivationStatusNotStarted {
		t.Errorf("ActivationStatus = %v, want not_started", progress.ActivationStatus)
	}
}

func TestStepProgress_NotStarted(t *testing.T) {
	step := domain.StepProgress{
		ID:     uuid.New(),
		StepID: uuid.New(),
		Status: domain.StepStatusNotStarted,
	}
	if step.Status != domain.StepStatusNotStarted {
		t.Errorf("Status = %v, want not_started", step.Status)
	}
}