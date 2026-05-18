package onboarding

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OnboardingTemplate defines the onboarding flow for a profile type
type OnboardingTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ProfileType string    `gorm:"type:varchar(20);index" json:"profile_type"` // brand, editor, influencer
	Version     string    `gorm:"type:varchar(10)" json:"version"`           // semantic version
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Steps       []OnboardingStep `gorm:"foreignKey:TemplateID" json:"steps,omitempty"`
}

// TableName returns the table name for OnboardingTemplate
func (OnboardingTemplate) TableName() string {
	return "onboarding_templates"
}

// BeforeCreate generates a UUID before creating a new template
func (t *OnboardingTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}