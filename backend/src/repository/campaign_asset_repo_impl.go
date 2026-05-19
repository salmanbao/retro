package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// CampaignAssetDB implements CampaignAssetRepository using GORM.
type CampaignAssetDB struct {
	db *gorm.DB
}

// NewCampaignAssetDB creates a new CampaignAssetDB.
func NewCampaignAssetDB(db *gorm.DB) *CampaignAssetDB {
	return &CampaignAssetDB{db: db}
}

// Create inserts a new campaign asset.
func (r *CampaignAssetDB) Create(ctx context.Context, asset *domain.CampaignAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// ByCampaignID retrieves all assets for a campaign.
func (r *CampaignAssetDB) ByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*domain.CampaignAsset, error) {
	var assets []*domain.CampaignAsset
	err := r.db.WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Order("created_at ASC").
		Find(&assets).Error
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// Delete removes an asset by ID.
func (r *CampaignAssetDB) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CampaignAsset{}, "id = ?", id).Error
}

// DeleteByCampaignID removes all assets for a campaign.
func (r *CampaignAssetDB) DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CampaignAsset{}, "campaign_id = ?", campaignID).Error
}
