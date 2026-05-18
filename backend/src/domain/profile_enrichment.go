package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ProfileEnrichment represents public profile information for marketplace participants.
type ProfileEnrichment struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID   uuid.UUID       `gorm:"type:uuid;uniqueIndex;not null" json:"profile_id"`
	Bio         string          `gorm:"type:text" json:"bio,omitempty"`
	AvatarURL   string          `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	CoverURL    string          `gorm:"type:varchar(500)" json:"cover_url,omitempty"`
	WebsiteURL  string          `gorm:"type:varchar(500)" json:"website_url,omitempty"`
	Location    string          `gorm:"type:varchar(255)" json:"location,omitempty"`
	Languages   pq.StringArray  `gorm:"type:text[]" json:"languages,omitempty"`
	Timezone    string          `gorm:"type:varchar(100)" json:"timezone,omitempty"`
	SocialLinks json.RawMessage `gorm:"type:jsonb" json:"social_links,omitempty"`
	CreatedAt   time.Time       `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for ProfileEnrichment.
func (ProfileEnrichment) TableName() string { return "profile_enrichments" }

// NewProfileEnrichment creates a new profile enrichment for a profile.
func NewProfileEnrichment(profileID uuid.UUID) *ProfileEnrichment {
	now := time.Now()
	return &ProfileEnrichment{
		ID:        uuid.New(),
		ProfileID: profileID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetSocialLinks parses and returns the social links struct.
func (p *ProfileEnrichment) GetSocialLinks() (*SocialLinks, error) {
	return SocialLinksFromJSON(p.SocialLinks)
}

// SetSocialLinks serializes and stores the social links.
func (p *ProfileEnrichment) SetSocialLinks(sl *SocialLinks) error {
	data, err := sl.ToJSON()
	if err != nil {
		return err
	}
	p.SocialLinks = data
	p.UpdatedAt = time.Now()
	return nil
}

// Update applies partial updates to profile enrichment.
func (p *ProfileEnrichment) Update(bio, avatarURL, coverURL, websiteURL, location string, languages []string, timezone string, socialLinks *SocialLinks) error {
	p.Bio = bio
	p.AvatarURL = avatarURL
	p.CoverURL = coverURL
	p.WebsiteURL = websiteURL
	p.Location = location
	p.Languages = languages
	p.Timezone = timezone
	if socialLinks != nil {
		if err := p.SetSocialLinks(socialLinks); err != nil {
			return err
		}
	}
	p.UpdatedAt = time.Now()
	return nil
}
