package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// CreativeBriefDB implements CreativeBriefRepository using GORM.
type CreativeBriefDB struct {
	db *gorm.DB
}

// NewCreativeBriefDB creates a new CreativeBriefDB.
func NewCreativeBriefDB(db *gorm.DB) *CreativeBriefDB {
	return &CreativeBriefDB{db: db}
}

// Create inserts a new creative brief.
func (r *CreativeBriefDB) Create(ctx context.Context, brief *domain.CreativeBrief) error {
	return r.db.WithContext(ctx).Create(brief).Error
}

// ByCampaignID retrieves a creative brief by campaign ID.
func (r *CreativeBriefDB) ByCampaignID(ctx context.Context, campaignID uuid.UUID) (*domain.CreativeBrief, error) {
	var brief domain.CreativeBrief
	err := r.db.WithContext(ctx).Where("campaign_id = ?", campaignID).First(&brief).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBriefNotFound
		}
		return nil, err
	}
	return &brief, nil
}

// Update updates an existing creative brief.
func (r *CreativeBriefDB) Update(ctx context.Context, brief *domain.CreativeBrief) error {
	brief.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(brief).Error
}
