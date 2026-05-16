package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// FollowerVerificationDB implements FollowerVerificationRepository using GORM.
type FollowerVerificationDB struct {
	db *gorm.DB
}

// NewFollowerVerificationDB creates a new FollowerVerificationDB.
func NewFollowerVerificationDB(db *gorm.DB) *FollowerVerificationDB {
	return &FollowerVerificationDB{db: db}
}

// Create inserts new follower verification.
func (r *FollowerVerificationDB) Create(ctx context.Context, verification *domain.FollowerVerification) error {
	return r.db.WithContext(ctx).Create(verification).Error
}

// ByProfileID retrieves follower verification by profile ID.
func (r *FollowerVerificationDB) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.FollowerVerification, error) {
	var verification domain.FollowerVerification
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &verification, nil
}

// Update updates existing follower verification.
func (r *FollowerVerificationDB) Update(ctx context.Context, verification *domain.FollowerVerification) error {
	return r.db.WithContext(ctx).Save(verification).Error
}