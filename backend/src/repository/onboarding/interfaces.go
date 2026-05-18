package onboarding

import (
	"github.com/google/uuid"
	"viralforge/backend/src/domain/onboarding"
)

// TemplateRepository defines operations for onboarding templates
type TemplateRepository interface {
	Create(template *onboarding.OnboardingTemplate) error
	GetByID(id uuid.UUID) (*onboarding.OnboardingTemplate, error)
	GetByProfileType(profileType string) (*onboarding.OnboardingTemplate, error)
	GetByProfileTypeAndVersion(profileType, version string) (*onboarding.OnboardingTemplate, error)
	GetLatestByProfileType(profileType string) (*onboarding.OnboardingTemplate, error)
	List() ([]onboarding.OnboardingTemplate, error)
	Update(template *onboarding.OnboardingTemplate) error
	Delete(id uuid.UUID) error
}

// ProgressRepository defines operations for onboarding progress
type ProgressRepository interface {
	Create(progress *onboarding.OnboardingProgress) error
	GetByID(id uuid.UUID) (*onboarding.OnboardingProgress, error)
	GetByProfileID(profileID uuid.UUID) (*onboarding.OnboardingProgress, error)
	Update(progress *onboarding.OnboardingProgress) error
	Delete(id uuid.UUID) error
}

// StepRepository defines operations for step progress
type StepRepository interface {
	Create(stepProgress *onboarding.StepProgress) error
	GetByID(id uuid.UUID) (*onboarding.StepProgress, error)
	GetByOnboardingProgressID(progressID uuid.UUID) ([]onboarding.StepProgress, error)
	GetByProgressIDAndStepID(progressID, stepID uuid.UUID) (*onboarding.StepProgress, error)
	Update(stepProgress *onboarding.StepProgress) error
	Delete(id uuid.UUID) error
}
