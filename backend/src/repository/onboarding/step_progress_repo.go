package onboarding

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain/onboarding"
)

// StepRepo implements StepRepository interface
type StepRepo struct {
	db *gorm.DB
}

// NewStepRepo creates a new StepRepo
func NewStepRepo(db *gorm.DB) *StepRepo {
	return &StepRepo{db: db}
}

// Create creates a new step progress
func (r *StepRepo) Create(stepProgress *onboarding.StepProgress) error {
	return r.db.Create(stepProgress).Error
}

// GetByID retrieves step progress by ID
func (r *StepRepo) GetByID(id uuid.UUID) (*onboarding.StepProgress, error) {
	var stepProgress onboarding.StepProgress
	err := r.db.Where("id = ?", id).First(&stepProgress).Error
	if err != nil {
		return nil, err
	}
	return &stepProgress, nil
}

// GetByOnboardingProgressID retrieves all step progress for a progress ID
func (r *StepRepo) GetByOnboardingProgressID(progressID uuid.UUID) ([]onboarding.StepProgress, error) {
	var stepProgresses []onboarding.StepProgress
	err := r.db.Where("onboarding_progress_id = ?", progressID).Find(&stepProgresses).Error
	if err != nil {
		return nil, err
	}
	return stepProgresses, nil
}

// GetByProgressIDAndStepID retrieves step progress by progress ID and step ID
func (r *StepRepo) GetByProgressIDAndStepID(progressID, stepID uuid.UUID) (*onboarding.StepProgress, error) {
	var stepProgress onboarding.StepProgress
	err := r.db.Where("onboarding_progress_id = ? AND step_id = ?", progressID, stepID).First(&stepProgress).Error
	if err != nil {
		return nil, err
	}
	return &stepProgress, nil
}

// Update updates an existing step progress
func (r *StepRepo) Update(stepProgress *onboarding.StepProgress) error {
	return r.db.Save(stepProgress).Error
}

// Delete deletes step progress
func (r *StepRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&onboarding.StepProgress{}, "id = ?", id).Error
}
