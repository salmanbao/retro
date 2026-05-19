package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// CampaignStatus represents the lifecycle state of a campaign.
type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusPublished CampaignStatus = "published"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

// Campaign represents a brand's campaign for short-form video content.
type Campaign struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BrandProfileID     uuid.UUID       `json:"brand_profile_id" gorm:"type:uuid;not null;index"`
	Title              string          `json:"title" gorm:"size:255;not null"`
	Slug               string          `json:"slug" gorm:"size:255;not null;uniqueIndex"`
	Summary            string          `json:"summary" gorm:"size:500"`
	Description        string          `json:"description" gorm:"type:text;not null"`
	Objective          string          `json:"objective" gorm:"size:255"`
	ProductName        string          `json:"product_name" gorm:"size:255"`
	LandingURL         string          `json:"landing_url" gorm:"size:2048"`
	TotalBudget        float64         `json:"total_budget" gorm:"type:decimal(12,2);not null;default:0"`
	Currency           string          `json:"currency" gorm:"size:3;not null;default:USD"`
	TargetClips        int             `json:"target_clips" gorm:"not null;default:0"`
	TargetPosts        int             `json:"target_posts" gorm:"not null;default:0"`
	CPV                float64         `json:"cpv" gorm:"type:decimal(10,4);not null;default:0"`
	MinPayout          *float64        `json:"min_payout" gorm:"type:decimal(12,2)"`
	MaxPayout          *float64        `json:"max_payout" gorm:"type:decimal(12,2)"`
	SubmissionStart    time.Time       `json:"submission_start" gorm:"not null"`
	SubmissionDeadline time.Time       `json:"submission_deadline" gorm:"not null"`
	DistributionStart  time.Time       `json:"distribution_start" gorm:"not null"`
	CampaignEnd        time.Time       `json:"campaign_end" gorm:"not null"`
	AllowedCountries   []string        `json:"allowed_countries" gorm:"type:text[];not null;default:'{}'"`
	AllowedLanguages   []string        `json:"allowed_languages" gorm:"type:text[];not null;default:'{}'"`
	MinFollowers       int             `json:"min_followers" gorm:"not null;default:0"`
	Platforms          []string        `json:"platforms" gorm:"type:text[];not null;default:'{}'"`
	CreatorCategories  []string        `json:"creator_categories" gorm:"type:text[]"`
	MinDurationSecs    int             `json:"min_duration_secs" gorm:"not null;default:15"`
	MaxDurationSecs    int             `json:"max_duration_secs" gorm:"not null;default:60"`
	AspectRatio        string          `json:"aspect_ratio" gorm:"size:10;not null;default:'9:16'"`
	TalkingPoints      []string        `json:"talking_points" gorm:"type:text[]"`
	ProhibitedClaims   []string        `json:"prohibited_claims" gorm:"type:text[]"`
	Hashtags           []string        `json:"hashtags" gorm:"type:text[]"`
	CTAInstructions    string          `json:"cta_instructions" gorm:"type:text"`
	Status             CampaignStatus  `json:"status" gorm:"size:20;not null;default:'draft';index"`
	Version            int             `json:"version" gorm:"not null;default:1"`
	CreatedAt          time.Time       `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time       `json:"updated_at" gorm:"not null;default:now()"`
	DeletedAt          *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
	Assets             []CampaignAsset `json:"assets,omitempty" gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Campaign.
func (Campaign) TableName() string {
	return "campaigns"
}

// IsValidTransition checks if a status transition is valid.
func (s CampaignStatus) IsValidTransition(target CampaignStatus) bool {
	transitions := map[CampaignStatus][]CampaignStatus{
		CampaignStatusDraft:     {CampaignStatusPublished, CampaignStatusCancelled},
		CampaignStatusPublished: {CampaignStatusActive, CampaignStatusCancelled},
		CampaignStatusActive:    {CampaignStatusPaused, CampaignStatusCompleted, CampaignStatusCancelled},
		CampaignStatusPaused:    {CampaignStatusActive, CampaignStatusCancelled},
		CampaignStatusCompleted: {},
		CampaignStatusCancelled: {},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}

// Domain errors for Campaign operations.
var (
	ErrSlugAlreadyExists  = errors.New("slug already exists")
	ErrInvalidTransition  = errors.New("invalid transition")
	ErrCampaignNotOwned   = errors.New("campaign not owned")
	ErrRestrictedEdit     = errors.New("restricted edit")
	ErrReadinessFailed    = errors.New("readiness failed")
	ErrBudgetRequired     = errors.New("budget required")
	ErrInvalidTimeline    = errors.New("invalid timeline")
	ErrInvalidPayoutRange = errors.New("invalid payout range")
)
