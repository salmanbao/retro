package domain

import (
	"time"

	"github.com/google/uuid"
)

// AssetType represents the type of campaign asset.
type AssetType string

const (
	AssetTypeReference AssetType = "reference"
	AssetTypeRawMedia  AssetType = "raw_media"
	AssetTypeDocument  AssetType = "document"
)

// CampaignAsset represents a reference to an uploaded asset or document linked to a campaign.
type CampaignAsset struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CampaignID  uuid.UUID `json:"campaign_id" gorm:"type:uuid;not null;index"`
	URL         string    `json:"url" gorm:"size:2048;not null"`
	AssetType   AssetType `json:"asset_type" gorm:"size:20;not null"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name for CampaignAsset.
func (CampaignAsset) TableName() string {
	return "campaign_assets"
}
