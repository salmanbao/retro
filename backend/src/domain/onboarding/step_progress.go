package onboarding

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StepStatus constants
const (
	StepStatusNotStarted = "not_started"
	StepStatusInProgress = "in_progress"
	StepStatusCompleted  = "completed"
	StepStatusSkipped    = "skipped"
)

// StepProgress tracks status of individual steps within onboarding progress
type StepProgress struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	OnboardingProgressID uuid.UUID  `gorm:"type:uuid;index" json:"onboarding_progress_id"`
	StepID               uuid.UUID  `gorm:"type:uuid" json:"step_id"`
	Status               string     `gorm:"type:varchar(20)" json:"status"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	LastViewedAt         *time.Time `json:"last_viewed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// TableName returns the table name for StepProgress
func (StepProgress) TableName() string {
	return "step_progresses"
}

// BeforeCreate generates a UUID before creating a new step progress
func (sp *StepProgress) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == uuid.Nil {
		sp.ID = uuid.New()
	}
	return nil
}

// ValidStepStatuses returns the valid step status values
var ValidStepStatuses = []string{StepStatusNotStarted, StepStatusInProgress, StepStatusCompleted, StepStatusSkipped}

// IsValidStepStatus checks if the step status is valid
func IsValidStepStatus(status string) bool {
	for _, valid := range ValidStepStatuses {
		if status == valid {
			return true
		}
	}
	return false
}
