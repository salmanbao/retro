package onboarding

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivationStatus constants
const (
	ActivationStatusNotStarted    = "not_started"
	ActivationStatusOnboarding    = "onboarding"
	ActivationStatusPendingReview = "pending_review"
	ActivationStatusActivated     = "activated"
)

// OnboardingProgress tracks onboarding progress for a specific profile
type OnboardingProgress struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	ProfileID        uuid.UUID      `gorm:"type:uuid;uniqueIndex" json:"profile_id"`
	ProfileType      string         `gorm:"type:varchar(20)" json:"profile_type"`
	TemplateID       uuid.UUID      `gorm:"type:uuid" json:"template_id"`
	TemplateVersion  string         `gorm:"type:varchar(10)" json:"template_version"`
	ActivationStatus string         `gorm:"type:varchar(20);index" json:"activation_status"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	LastActivityAt   *time.Time     `json:"last_activity_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	StepProgresses   []StepProgress `gorm:"foreignKey:OnboardingProgressID" json:"step_progresses,omitempty"`
}

// TableName returns the table name for OnboardingProgress
func (OnboardingProgress) TableName() string {
	return "onboarding_progresses"
}

// BeforeCreate generates a UUID before creating a new progress record
func (p *OnboardingProgress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ValidActivationStatuses returns the valid activation status values
var ValidActivationStatuses = []string{ActivationStatusNotStarted, ActivationStatusOnboarding, ActivationStatusPendingReview, ActivationStatusActivated}

// IsValidActivationStatus checks if the activation status is valid
func IsValidActivationStatus(status string) bool {
	for _, valid := range ValidActivationStatuses {
		if status == valid {
			return true
		}
	}
	return false
}
