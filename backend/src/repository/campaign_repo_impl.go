package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// CampaignDB implements CampaignRepository using GORM.
type CampaignDB struct {
	db *gorm.DB
}

// NewCampaignDB creates a new CampaignDB.
func NewCampaignDB(db *gorm.DB) *CampaignDB {
	return &CampaignDB{db: db}
}

// Create inserts a new campaign.
func (r *CampaignDB) Create(ctx context.Context, campaign *domain.Campaign) error {
	return r.db.WithContext(ctx).Create(campaign).Error
}

// ByID retrieves a campaign by ID.
func (r *CampaignDB) ByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error) {
	var campaign domain.Campaign
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&campaign).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCampaignNotFound
		}
		return nil, err
	}
	return &campaign, nil
}

// BySlug retrieves a campaign by slug.
func (r *CampaignDB) BySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	var campaign domain.Campaign
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&campaign).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCampaignNotFound
		}
		return nil, err
	}
	return &campaign, nil
}

// ByBrandProfileID retrieves all campaigns for a brand profile.
func (r *CampaignDB) ByBrandProfileID(ctx context.Context, brandProfileID uuid.UUID) ([]*domain.Campaign, error) {
	var campaigns []*domain.Campaign
	err := r.db.WithContext(ctx).
		Where("brand_profile_id = ? AND deleted_at IS NULL", brandProfileID).
		Order("created_at DESC").
		Find(&campaigns).Error
	if err != nil {
		return nil, err
	}
	return campaigns, nil
}

// Update updates an existing campaign.
func (r *CampaignDB) Update(ctx context.Context, campaign *domain.Campaign) error {
	return r.db.WithContext(ctx).Save(campaign).Error
}

// Delete soft-deletes a campaign.
func (r *CampaignDB) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Campaign{}, "id = ?", id).Error
}

// List retrieves campaigns with pagination and optional status filter.
func (r *CampaignDB) List(ctx context.Context, brandProfileID uuid.UUID, status string, page, pageSize int) ([]*domain.Campaign, int64, error) {
	var campaigns []*domain.Campaign
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Campaign{}).
		Where("brand_profile_id = ? AND deleted_at IS NULL", brandProfileID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&campaigns).Error
	if err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// ByStatusAndDeadline finds campaigns that need automatic state transition.
// Used for deadline-based transitions (e.g., published -> active when submission_deadline passes).
func (r *CampaignDB) ByStatusAndDeadline(ctx context.Context, status domain.CampaignStatus, deadline time.Time) ([]*domain.Campaign, error) {
	var campaigns []*domain.Campaign
	err := r.db.WithContext(ctx).
		Where("status = ? AND submission_deadline <= ? AND deleted_at IS NULL", status, deadline).
		Find(&campaigns).Error
	if err != nil {
		return nil, err
	}
	return campaigns, nil
}
