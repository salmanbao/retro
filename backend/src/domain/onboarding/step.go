package onboarding

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OnboardingStep defines a single step within an onboarding template
type OnboardingStep struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	TemplateID     uuid.UUID `gorm:"type:uuid;index" json:"template_id"`
	Title          string    `gorm:"type:varchar(100)" json:"title"`
	Description    string    `gorm:"type:text" json:"description,omitempty"`
	ActionURL      string    `gorm:"type:varchar(500)" json:"action_url,omitempty"`
	StepType       string    `gorm:"type:varchar(30)" json:"step_type"` // tutorial, checklist, verification, profile_completion
	Required       bool      `gorm:"default:false" json:"required"`
	DisplayOrder   int       `gorm:"default:0" json:"display_order"`
	AutoCompleteKey string   `gorm:"type:varchar(50)" json:"auto_complete_key,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName returns the table name for OnboardingStep
func (OnboardingStep) TableName() string {
	return "onboarding_steps"
}

// BeforeCreate generates a UUID before creating a new step
func (s *OnboardingStep) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ValidStepTypes returns the valid step type values
var ValidStepTypes = []string{"tutorial", "checklist", "verification", "profile_completion"}

// IsValidStepType checks if the step type is valid
func IsValidStepType(stepType string) bool {
	for _, valid := range ValidStepTypes {
		if stepType == valid {
			return true
		}
	}
	return false
}