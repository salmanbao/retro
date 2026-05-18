package onboarding

import (
	"github.com/google/uuid"
	"viralforge/backend/src/domain/onboarding"
	"gorm.io/gorm"
)

// ProgressRepo implements ProgressRepository interface
type ProgressRepo struct {
	db *gorm.DB
}

// NewProgressRepo creates a new ProgressRepo
func NewProgressRepo(db *gorm.DB) *ProgressRepo {
	return &ProgressRepo{db: db}
}

// Create creates a new onboarding progress
func (r *ProgressRepo) Create(progress *onboarding.OnboardingProgress) error {
	return r.db.Create(progress).Error
}

// GetByID retrieves progress by ID
func (r *ProgressRepo) GetByID(id uuid.UUID) (*onboarding.OnboardingProgress, error) {
	var progress onboarding.OnboardingProgress
	err := r.db.Preload("StepProgresses").Where("id = ?", id).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetByProfileID retrieves progress by profile ID
func (r *ProgressRepo) GetByProfileID(profileID uuid.UUID) (*onboarding.OnboardingProgress, error) {
	var progress onboarding.OnboardingProgress
	err := r.db.Preload("StepProgresses").Where("profile_id = ?", profileID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// Update updates an existing progress
func (r *ProgressRepo) Update(progress *onboarding.OnboardingProgress) error {
	return r.db.Save(progress).Error
}

// Delete deletes progress
func (r *ProgressRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&onboarding.OnboardingProgress{}, "id = ?", id).Error
}