package domain

import (
	"time"

	"github.com/google/uuid"
)

// PortfolioItem represents an Editor's portfolio piece with media and links.
type PortfolioItem struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"profile_id"`
	Title        string     `gorm:"type:varchar(200);not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description,omitempty"`
	ThumbnailURL string     `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	VideoURL     string     `gorm:"type:varchar(500)" json:"video_url,omitempty"`
	ExternalLink string     `gorm:"type:varchar(500)" json:"external_link,omitempty"`
	DisplayOrder int        `gorm:"type:int;not null;default:0" json:"display_order"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"type:timestamptz;index" json:"deleted_at,omitempty"`
}

// TableName sets the table name for PortfolioItem.
func (PortfolioItem) TableName() string { return "portfolio_items" }

// NewPortfolioItem creates a new portfolio item with the given fields.
func NewPortfolioItem(profileID uuid.UUID, title, description, thumbnailURL, videoURL, externalLink string, displayOrder int) *PortfolioItem {
	return &PortfolioItem{
		ID:           uuid.New(),
		ProfileID:    profileID,
		Title:        title,
		Description:  description,
		ThumbnailURL: thumbnailURL,
		VideoURL:     videoURL,
		ExternalLink: externalLink,
		DisplayOrder: displayOrder,
	}
}

// IsDeleted returns true if the portfolio item has been soft-deleted.
func (p *PortfolioItem) IsDeleted() bool {
	return p.DeletedAt != nil
}

// SoftDelete marks the portfolio item as deleted.
func (p *PortfolioItem) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}

// Update updates portfolio item fields.
func (p *PortfolioItem) Update(title, description, thumbnailURL, videoURL, externalLink string, displayOrder int) {
	p.Title = title
	p.Description = description
	p.ThumbnailURL = thumbnailURL
	p.VideoURL = videoURL
	p.ExternalLink = externalLink
	p.DisplayOrder = displayOrder
	p.UpdatedAt = time.Now()
}
