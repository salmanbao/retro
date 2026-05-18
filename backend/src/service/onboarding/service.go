package onboarding

import (
	"time"

	"github.com/google/uuid"
	domain "viralforge/backend/src/domain/onboarding"
	"viralforge/backend/src/repository/onboarding"
)

// Service provides onboarding business logic
type Service struct {
	templateRepo *onboarding.TemplateRepo
	progressRepo *onboarding.ProgressRepo
	stepRepo     *onboarding.StepRepo
}

// NewService creates a new onboarding service
func NewService(
	templateRepo *onboarding.TemplateRepo,
	progressRepo *onboarding.ProgressRepo,
	stepRepo *onboarding.StepRepo,
) *Service {
	return &Service{
		templateRepo: templateRepo,
		progressRepo: progressRepo,
		stepRepo:     stepRepo,
	}
}

// GetTemplateByProfileType retrieves the latest template for a profile type
func (s *Service) GetTemplateByProfileType(profileType string) (*domain.OnboardingTemplate, error) {
	if !domain.IsValidProfileType(profileType) {
		return nil, domain.ErrInvalidProfileType
	}

	template, err := s.templateRepo.GetByProfileType(profileType)
	if err != nil {
		return nil, domain.ErrTemplateNotFound
	}

	return template, nil
}

// GetTemplateByID retrieves a template by ID
func (s *Service) GetTemplateByID(id uuid.UUID) (*domain.OnboardingTemplate, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrTemplateNotFound
	}
	return template, nil
}

// ListTemplates retrieves all templates
func (s *Service) ListTemplates() ([]domain.OnboardingTemplate, error) {
	return s.templateRepo.List()
}

// GetOrCreateProgress retrieves or creates onboarding progress for a profile
func (s *Service) GetOrCreateProgress(profileID uuid.UUID, profileType string) (*domain.OnboardingProgress, error) {
	// Try to get existing progress
	progress, err := s.progressRepo.GetByProfileID(profileID)
	if err == nil {
		return progress, nil
	}

	// If not found, check if it's a "not found" error
	if err != domain.ErrProgressNotFound {
		// For other errors, return as-is
		// But first check if it's just gorm returning no rows
		progress, getErr := s.progressRepo.GetByProfileID(profileID)
		if getErr == nil {
			return progress, nil
		}
	}

	// Get the template for this profile type
	template, err := s.templateRepo.GetByProfileType(profileType)
	if err != nil {
		return nil, domain.ErrTemplateNotFound
	}

	// Create new progress with snapshot
	now := time.Now()
	newProgress := &domain.OnboardingProgress{
		ID:               uuid.New(),
		ProfileID:        profileID,
		ProfileType:      profileType,
		TemplateID:       template.ID,
		TemplateVersion:  template.Version,
		ActivationStatus: domain.ActivationStatusNotStarted,
		StartedAt:        nil,
		LastActivityAt:   &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.progressRepo.Create(newProgress); err != nil {
		return nil, err
	}

	// Create step progress records for each template step
	for _, step := range template.Steps {
		stepProgress := &domain.StepProgress{
			ID:                   uuid.New(),
			OnboardingProgressID: newProgress.ID,
			StepID:               step.ID,
			Status:               domain.StepStatusNotStarted,
			StartedAt:            nil,
			CompletedAt:          nil,
		}
		if err := s.stepRepo.Create(stepProgress); err != nil {
			return nil, err
		}
	}

	return newProgress, nil
}

// GetProgressByProfileID retrieves onboarding progress for a profile
func (s *Service) GetProgressByProfileID(profileID uuid.UUID) (*domain.OnboardingProgress, error) {
	return s.progressRepo.GetByProfileID(profileID)
}

// UpdateStepStatus updates the status of a step
func (s *Service) UpdateStepStatus(progressID, stepID uuid.UUID, newStatus string) (*domain.StepProgress, error) {
	// Get current step progress
	stepProgress, err := s.stepRepo.GetByProgressIDAndStepID(progressID, stepID)
	if err != nil {
		return nil, domain.ErrStepNotFound
	}

	// Validate the transition
	if !domain.IsValidStepStatusTransition(stepProgress.Status, newStatus) {
		return nil, domain.ErrInvalidStepStatus
	}

	// Check if trying to skip a required step
	if newStatus == domain.StepStatusSkipped {
		// Get the progress to find template
		progress, err := s.progressRepo.GetByID(progressID)
		if err != nil {
			return nil, err
		}

		// Get template to check if step is required
		template, err := s.templateRepo.GetByID(progress.TemplateID)
		if err != nil {
			return nil, err
		}

		for _, step := range template.Steps {
			if step.ID == stepID && step.Required {
				return nil, domain.ErrStepNotSkippable
			}
		}
	}

	// Update the step progress
	now := time.Now()
	stepProgress.Status = newStatus
	if newStatus == domain.StepStatusInProgress {
		stepProgress.StartedAt = &now
	} else if newStatus == domain.StepStatusCompleted {
		stepProgress.CompletedAt = &now
	}

	if err := s.stepRepo.Update(stepProgress); err != nil {
		return nil, err
	}

	// Update progress last_activity_at
	progress, err := s.progressRepo.GetByID(progressID)
	if err == nil {
		progress.LastActivityAt = &now
		// Update activation status if needed
		if progress.ActivationStatus == domain.ActivationStatusNotStarted {
			progress.ActivationStatus = domain.ActivationStatusOnboarding
			progress.StartedAt = &now
		}
		s.progressRepo.Update(progress)
	}

	return stepProgress, nil
}

// RecalculateProgress recalculates activation status and auto-completes steps
func (s *Service) RecalculateProgress(profileID uuid.UUID) (*domain.OnboardingProgress, error) {
	progress, err := s.progressRepo.GetByProfileID(profileID)
	if err != nil {
		return nil, err
	}

	template, err := s.templateRepo.GetByProfileType(progress.ProfileType)
	if err != nil {
		return nil, err
	}

	// Get current step progress
	stepProgresses, err := s.stepRepo.GetByOnboardingProgressID(progress.ID)
	if err != nil {
		return nil, err
	}

	// Create a map of existing step progress
	existingSteps := make(map[uuid.UUID]domain.StepProgress)
	for _, sp := range stepProgresses {
		existingSteps[sp.StepID] = sp
	}

	// Add missing steps from template (new in newer version)
	for _, step := range template.Steps {
		if _, exists := existingSteps[step.ID]; !exists {
			// New step from newer template version
			newStepProgress := &domain.StepProgress{
				ID:                   uuid.New(),
				OnboardingProgressID: progress.ID,
				StepID:               step.ID,
				Status:               domain.StepStatusNotStarted,
				StartedAt:            nil,
				CompletedAt:          nil,
			}
			s.stepRepo.Create(newStepProgress)
		}
	}

	// Recalculate activation status
	steps, _ := s.stepRepo.GetByOnboardingProgressID(progress.ID)
	completedRequired := 0
	totalRequired := 0

	for _, step := range steps {
		for _, tStep := range template.Steps {
			if tStep.ID == step.StepID && tStep.Required {
				totalRequired++
				if step.Status == domain.StepStatusCompleted || step.Status == domain.StepStatusSkipped {
					completedRequired++
				}
			}
		}
	}

	now := time.Now()
	// Update activation status based on completion
	switch progress.ActivationStatus {
	case domain.ActivationStatusNotStarted:
		if len(steps) > 0 {
			progress.ActivationStatus = domain.ActivationStatusOnboarding
			progress.StartedAt = &now
		}
	case domain.ActivationStatusOnboarding:
		if completedRequired == totalRequired && totalRequired > 0 {
			progress.ActivationStatus = domain.ActivationStatusPendingReview
		}
	}
	progress.LastActivityAt = &now
	progress.UpdatedAt = now

	if err := s.progressRepo.Update(progress); err != nil {
		return nil, err
	}

	return progress, nil
}

// GetNextStep returns the first incomplete step for a profile's onboarding progress
func (s *Service) GetNextStep(profileID uuid.UUID) (*domain.OnboardingStep, error) {
	progress, err := s.progressRepo.GetByProfileID(profileID)
	if err != nil {
		return nil, err
	}

	template, err := s.templateRepo.GetByProfileType(progress.ProfileType)
	if err != nil {
		return nil, err
	}

	steps, err := s.stepRepo.GetByOnboardingProgressID(progress.ID)
	if err != nil {
		return nil, err
	}

	// Create a map of step progress by step ID
	stepProgressMap := make(map[uuid.UUID]domain.StepProgress)
	for _, sp := range steps {
		stepProgressMap[sp.StepID] = sp
	}

	// Find first incomplete step ordered by display_order
	for _, step := range template.Steps {
		sp, exists := stepProgressMap[step.ID]
		if !exists {
			// Step not in progress yet - it would be added by recalculate
			continue
		}
		if sp.Status == domain.StepStatusNotStarted || sp.Status == domain.StepStatusInProgress {
			return &step, nil
		}
	}

	return nil, nil // All steps completed
}
