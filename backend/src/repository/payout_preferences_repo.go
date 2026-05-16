package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// PayoutPreferencesDB implements PayoutPreferencesRepository using GORM.
type PayoutPreferencesDB struct {
	db *gorm.DB
}

// NewPayoutPreferencesDB creates a new PayoutPreferencesDB.
func NewPayoutPreferencesDB(db *gorm.DB) *PayoutPreferencesDB {
	return &PayoutPreferencesDB{db: db}
}

// Create inserts new payout preferences.
func (r *PayoutPreferencesDB) Create(ctx context.Context, prefs *domain.PayoutPreferences) error {
	return r.db.WithContext(ctx).Create(prefs).Error
}

// ByProfileID retrieves payout preferences by profile ID.
func (r *PayoutPreferencesDB) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.PayoutPreferences, error) {
	var prefs domain.PayoutPreferences
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).First(&prefs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &prefs, nil
}

// Update updates existing payout preferences.
func (r *PayoutPreferencesDB) Update(ctx context.Context, prefs *domain.PayoutPreferences) error {
	return r.db.WithContext(ctx).Save(prefs).Error
}