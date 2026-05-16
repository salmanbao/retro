package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AudienceData represents an Influencer's audience information and demographics.
type AudienceData struct {
	ProfileID            uuid.UUID       `gorm:"type:uuid;primary_key;not null" json:"profile_id"`
	PlatformHandles      json.RawMessage `gorm:"type:jsonb" json:"platform_handles,omitempty"`
	ClaimedFollowers     json.RawMessage `gorm:"type:jsonb" json:"claimed_followers,omitempty"`
	EngagementRate       float64         `gorm:"type:decimal(5,2)" json:"engagement_rate,omitempty"`
	AudienceDemographics json.RawMessage `gorm:"type:jsonb" json:"audience_demographics,omitempty"`
	UpdatedAt            time.Time       `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for AudienceData.
func (AudienceData) TableName() string { return "audience_data" }

// PlatformHandlesFromJSON parses the platform handles JSON.
func (a *AudienceData) GetPlatformHandles() (map[string]string, error) {
	if a.PlatformHandles == nil || len(a.PlatformHandles) == 0 {
		return map[string]string{}, nil
	}
	var result map[string]string
	err := json.Unmarshal(a.PlatformHandles, &result)
	return result, err
}

// SetPlatformHandles serializes platform handles.
func (a *AudienceData) SetPlatformHandles(handles map[string]string) error {
	data, err := json.Marshal(handles)
	if err != nil {
		return err
	}
	a.PlatformHandles = data
	a.UpdatedAt = time.Now()
	return nil
}

// GetClaimedFollowers parses the claimed followers JSON.
func (a *AudienceData) GetClaimedFollowers() (map[string]int, error) {
	if a.ClaimedFollowers == nil || len(a.ClaimedFollowers) == 0 {
		return map[string]int{}, nil
	}
	var result map[string]int
	err := json.Unmarshal(a.ClaimedFollowers, &result)
	return result, err
}

// SetClaimedFollowers serializes claimed followers.
func (a *AudienceData) SetClaimedFollowers(followers map[string]int) error {
	data, err := json.Marshal(followers)
	if err != nil {
		return err
	}
	a.ClaimedFollowers = data
	a.UpdatedAt = time.Now()
	return nil
}

// GetAudienceDemographics parses audience demographics JSON.
func (a *AudienceData) GetAudienceDemographics() (json.RawMessage, error) {
	return a.AudienceDemographics, nil
}

// SetAudienceDemographics sets demographics JSON with size validation (max 10KB).
func (a *AudienceData) SetAudienceDemographics(data json.RawMessage) error {
	if len(data) > 10*1024 {
		return ErrDemographicsTooLarge
	}
	a.AudienceDemographics = data
	a.UpdatedAt = time.Now()
	return nil
}

// Update updates audience data fields.
func (a *AudienceData) Update(handles map[string]string, followers map[string]int, engagementRate float64, demographics json.RawMessage) error {
	if err := a.SetPlatformHandles(handles); err != nil {
		return err
	}
	if err := a.SetClaimedFollowers(followers); err != nil {
		return err
	}
	a.EngagementRate = engagementRate
	if demographics != nil {
		if err := a.SetAudienceDemographics(demographics); err != nil {
			return err
		}
	}
	a.UpdatedAt = time.Now()
	return nil
}
