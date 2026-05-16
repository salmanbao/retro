package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// AudienceDataDB implements AudienceDataRepository using GORM.
type AudienceDataDB struct {
	db *gorm.DB
}

// NewAudienceDataDB creates a new AudienceDataDB.
func NewAudienceDataDB(db *gorm.DB) *AudienceDataDB {
	return &AudienceDataDB{db: db}
}

// Create inserts new audience data.
func (r *AudienceDataDB) Create(ctx context.Context, data *domain.AudienceData) error {
	return r.db.WithContext(ctx).Create(data).Error
}

// ByProfileID retrieves audience data by profile ID.
func (r *AudienceDataDB) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.AudienceData, error) {
	var data domain.AudienceData
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &data, nil
}

// Update updates existing audience data.
func (r *AudienceDataDB) Update(ctx context.Context, data *domain.AudienceData) error {
	return r.db.WithContext(ctx).Save(data).Error
}