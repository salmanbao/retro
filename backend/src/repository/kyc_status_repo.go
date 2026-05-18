package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// KYCStatusDB implements KYCStatusRepository using GORM.
type KYCStatusDB struct {
	db *gorm.DB
}

// NewKYCStatusDB creates a new KYCStatusDB.
func NewKYCStatusDB(db *gorm.DB) *KYCStatusDB {
	return &KYCStatusDB{db: db}
}

// Create inserts new KYC status.
func (r *KYCStatusDB) Create(ctx context.Context, status *domain.KYCStatus) error {
	return r.db.WithContext(ctx).Create(status).Error
}

// ByProfileID retrieves KYC status by profile ID.
func (r *KYCStatusDB) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.KYCStatus, error) {
	var status domain.KYCStatus
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &status, nil
}

// Update updates existing KYC status.
func (r *KYCStatusDB) Update(ctx context.Context, status *domain.KYCStatus) error {
	return r.db.WithContext(ctx).Save(status).Error
}
