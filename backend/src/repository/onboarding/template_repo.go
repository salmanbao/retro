package onboarding

import (
	"github.com/google/uuid"
	"viralforge/backend/src/domain/onboarding"
	"gorm.io/gorm"
)

// TemplateRepo implements TemplateRepository interface
type TemplateRepo struct {
	db *gorm.DB
}

// NewTemplateRepo creates a new TemplateRepo
func NewTemplateRepo(db *gorm.DB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

// Create creates a new onboarding template
func (r *TemplateRepo) Create(template *onboarding.OnboardingTemplate) error {
	return r.db.Create(template).Error
}

// GetByID retrieves a template by ID
func (r *TemplateRepo) GetByID(id uuid.UUID) (*onboarding.OnboardingTemplate, error) {
	var template onboarding.OnboardingTemplate
	err := r.db.Preload("Steps").Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetByProfileType retrieves the latest template for a profile type
func (r *TemplateRepo) GetByProfileType(profileType string) (*onboarding.OnboardingTemplate, error) {
	var template onboarding.OnboardingTemplate
	err := r.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Where("profile_type = ?", profileType).Order("created_at DESC").First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetByProfileTypeAndVersion retrieves a specific template version
func (r *TemplateRepo) GetByProfileTypeAndVersion(profileType, version string) (*onboarding.OnboardingTemplate, error) {
	var template onboarding.OnboardingTemplate
	err := r.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Where("profile_type = ? AND version = ?", profileType, version).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetLatestByProfileType retrieves the latest version of a template for a profile type
func (r *TemplateRepo) GetLatestByProfileType(profileType string) (*onboarding.OnboardingTemplate, error) {
	return r.GetByProfileType(profileType)
}

// List retrieves all templates
func (r *TemplateRepo) List() ([]onboarding.OnboardingTemplate, error) {
	var templates []onboarding.OnboardingTemplate
	err := r.db.Preload("Steps").Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// Update updates an existing template
func (r *TemplateRepo) Update(template *onboarding.OnboardingTemplate) error {
	return r.db.Save(template).Error
}

// Delete deletes a template
func (r *TemplateRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&onboarding.OnboardingTemplate{}, "id = ?", id).Error
}