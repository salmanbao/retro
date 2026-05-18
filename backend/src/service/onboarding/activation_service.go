package onboarding

import (
	"time"

	"github.com/google/uuid"
	domain "viralforge/backend/src/domain/onboarding"
)

// ActivationService handles activation state machine transitions
type ActivationService struct {
	templateRepo interface {
		GetByProfileType(profileType string) (*domain.OnboardingTemplate, error)
	}
	progressRepo interface {
		GetByProfileID(profileID uuid.UUID) (*domain.OnboardingProgress, error)
		Update(progress *domain.OnboardingProgress) error
	}
	stepRepo interface {
		GetByOnboardingProgressID(progressID uuid.UUID) ([]domain.StepProgress, error)
	}
}

// NewActivationService creates a new activation service
func NewActivationService(
	templateRepo interface {
		GetByProfileType(profileType string) (*domain.OnboardingTemplate, error)
	},
	progressRepo interface {
		GetByProfileID(profileID uuid.UUID) (*domain.OnboardingProgress, error)
		Update(progress *domain.OnboardingProgress) error
	},
	stepRepo interface {
		GetByOnboardingProgressID(progressID uuid.UUID) ([]domain.StepProgress, error)
	},
) *ActivationService {
	return &ActivationService{
		templateRepo: templateRepo,
		progressRepo: progressRepo,
		stepRepo:     stepRepo,
	}
}

// ComputeActivationStatus computes the new activation status based on step completion
func (s *ActivationService) ComputeActivationStatus(progress *domain.OnboardingProgress) (string, error) {
	steps, err := s.stepRepo.GetByOnboardingProgressID(progress.ID)
	if err != nil {
		return "", err
	}

	// Get template for required step info
	template, err := s.templateRepo.GetByProfileType(progress.ProfileID.String())
	if err != nil {
		// If we can't get template, use existing progress status
		return progress.ActivationStatus, nil
	}

	// Count completed and required steps
	completedRequired := 0
	totalRequired := 0

	for _, step := range steps {
		// Find the template step to check if required
		for _, tStep := range template.Steps {
			if tStep.ID == step.StepID && tStep.Required {
				totalRequired++
				if step.Status == domain.StepStatusCompleted {
					completedRequired++
				}
			}
		}
	}

	// Determine new activation status
	switch progress.ActivationStatus {
	case domain.ActivationStatusNotStarted:
		// Any step started transitions to onboarding
		if len(steps) > 0 {
			return domain.ActivationStatusOnboarding, nil
		}
	case domain.ActivationStatusOnboarding:
		// All required steps completed/skipped transitions to pending_review
		if completedRequired == totalRequired && totalRequired > 0 {
			return domain.ActivationStatusPendingReview, nil
		}
	case domain.ActivationStatusPendingReview:
		// Only admin can transition from pending_review (handled separately)
		return domain.ActivationStatusPendingReview, nil
	case domain.ActivationStatusActivated:
		return domain.ActivationStatusActivated, nil
	}

	return progress.ActivationStatus, nil
}

// ValidateRequiredSteps checks if all required steps are complete
func (s *ActivationService) ValidateRequiredSteps(progress *domain.OnboardingProgress) (bool, error) {
	steps, err := s.stepRepo.GetByOnboardingProgressID(progress.ID)
	if err != nil {
		return false, err
	}

	template, err := s.templateRepo.GetByProfileType(progress.ProfileID.String())
	if err != nil {
		return false, err
	}

	for _, tStep := range template.Steps {
		if tStep.Required {
			found := false
			for _, step := range steps {
				if step.StepID == tStep.ID {
					if step.Status == domain.StepStatusCompleted || step.Status == domain.StepStatusSkipped {
						found = true
						break
					}
				}
			}
			if !found {
				return false, nil
			}
		}
	}

	return true, nil
}

// ActivateProfile transitions a profile from pending_review to activated
func (s *ActivationService) ActivateProfile(progress *domain.OnboardingProgress) error {
	if progress.ActivationStatus != domain.ActivationStatusPendingReview {
		return domain.ErrProfileNotPendingReview
	}

	now := time.Now()
	progress.ActivationStatus = domain.ActivationStatusActivated
	progress.LastActivityAt = &now

	return s.progressRepo.Update(progress)
}

// CalculatePercentage calculates the completion percentage
func (s *ActivationService) CalculatePercentage(progress *domain.OnboardingProgress) (int, error) {
	steps, err := s.stepRepo.GetByOnboardingProgressID(progress.ID)
	if err != nil {
		return 0, err
	}

	if len(steps) == 0 {
		return 0, nil
	}

	completed := 0
	for _, step := range steps {
		if step.Status == domain.StepStatusCompleted || step.Status == domain.StepStatusSkipped {
			completed++
		}
	}

	return (completed * 100) / len(steps), nil
}
