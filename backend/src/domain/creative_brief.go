package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CreativeBrief struct {
	ID                     uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CampaignID             uuid.UUID  `json:"campaign_id" gorm:"type:uuid;not null;uniqueIndex"`
	KeyMessages            JSONBArray `json:"key_messages" gorm:"type:jsonb;not null;default:'[]'"`
	ProductBenefits        JSONBArray `json:"product_benefits" gorm:"type:jsonb;not null;default:'[]'"`
	MandatoryTalkingPoints JSONBArray `json:"mandatory_talking_points" gorm:"type:jsonb;not null;default:'[]'"`
	ProhibitedClaims       JSONBArray `json:"prohibited_claims" gorm:"type:jsonb;default:'[]'"`
	RequiredHashtags       JSONBArray `json:"required_hashtags" gorm:"type:jsonb;not null;default:'[]'"`
	CallToActionText       *string    `json:"call_to_action_text,omitempty"`
	ToneAndStyleGuidelines *string    `json:"tone_and_style_guidelines,omitempty"`
	TargetAudienceDesc     *string    `json:"target_audience_description,omitempty"`
	CompetitorReferences   JSONBArray `json:"competitor_references" gorm:"type:jsonb;default:'[]'"`
	ExampleVideoLinks      JSONBArray `json:"example_video_links" gorm:"type:jsonb;default:'[]'"`
	CreatedAt              time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt              time.Time  `json:"updated_at" gorm:"not null;default:now()"`
}

func (CreativeBrief) TableName() string {
	return "creative_briefs"
}

type JSONBArray []string

func (j *JSONBArray) Scan(value interface{}) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONBArray")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBArray) Value() (interface{}, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

var (
	ErrBriefCampaignRequired = errors.New("campaign_id is required")
	ErrBriefKeyMessagesEmpty = errors.New("key_messages cannot be empty")
	ErrBriefHashtagsRequired = errors.New("required_hashtags cannot be empty")
	ErrBriefTalkingPointsReq = errors.New("mandatory_talking_points cannot be empty")
)

func (b *CreativeBrief) Validate() error {
	if b.CampaignID == uuid.Nil {
		return ErrBriefCampaignRequired
	}
	if len(b.KeyMessages) == 0 {
		return ErrBriefKeyMessagesEmpty
	}
	if len(b.RequiredHashtags) == 0 {
		return ErrBriefHashtagsRequired
	}
	if len(b.MandatoryTalkingPoints) == 0 {
		return ErrBriefTalkingPointsReq
	}
	return nil
}

func (b *CreativeBrief) CanEditFull(campaignStatus CampaignStatus) bool {
	return campaignStatus == CampaignStatusDraft || campaignStatus == CampaignStatusPaused
}

func (b *CreativeBrief) RestrictedEditFields() []string {
	return []string{
		"tone_and_style_guidelines",
		"target_audience_description",
		"example_video_links",
	}
}
