package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// AssetMetadataDB implements AssetMetadataRepository using GORM.
type AssetMetadataDB struct {
	db *gorm.DB
}

// NewAssetMetadataDB creates a new AssetMetadataDB.
func NewAssetMetadataDB(db *gorm.DB) *AssetMetadataDB {
	return &AssetMetadataDB{db: db}
}

// Create inserts a new asset metadata record.
func (r *AssetMetadataDB) Create(ctx context.Context, asset *domain.AssetMetadata) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// ByID retrieves an asset by ID (excludes soft-deleted).
func (r *AssetMetadataDB) ByID(ctx context.Context, id uuid.UUID) (*domain.AssetMetadata, error) {
	var asset domain.AssetMetadata
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	return &asset, nil
}

// ListByCampaign retrieves all assets for a campaign with pagination (excludes soft-deleted).
func (r *AssetMetadataDB) ListByCampaign(ctx context.Context, campaignID uuid.UUID, page, pageSize int) ([]*domain.AssetMetadata, int64, error) {
	var assets []*domain.AssetMetadata
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AssetMetadata{}).
		Where("campaign_id = ? AND deleted_at IS NULL", campaignID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// Update updates an existing asset metadata record.
func (r *AssetMetadataDB) Update(ctx context.Context, asset *domain.AssetMetadata) error {
	asset.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(asset).Error
}

// SoftDelete soft-deletes an asset by setting deleted_at.
func (r *AssetMetadataDB) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.AssetMetadata{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}

// ByCampaignAndFilename retrieves the latest version of an asset by campaign and filename.
func (r *AssetMetadataDB) ByCampaignAndFilename(ctx context.Context, campaignID uuid.UUID, filename string) (*domain.AssetMetadata, error) {
	var asset domain.AssetMetadata
	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND original_filename = ? AND deleted_at IS NULL", campaignID, filename).
		Order("version DESC").
		First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	return &asset, nil
}

// ListVersions retrieves all versions of an asset by campaign and filename.
func (r *AssetMetadataDB) ListVersions(ctx context.Context, campaignID uuid.UUID, filename string) ([]*domain.AssetMetadata, error) {
	var assets []*domain.AssetMetadata
	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND original_filename = ?", campaignID, filename).
		Order("version DESC").
		Find(&assets).Error
	if err != nil {
		return nil, err
	}
	return assets, nil
}
