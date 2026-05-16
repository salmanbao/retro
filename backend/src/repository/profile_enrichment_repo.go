package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// ProfileEnrichmentDB implements ProfileEnrichmentRepository using GORM.
type ProfileEnrichmentDB struct {
	db *gorm.DB
}

// NewProfileEnrichmentDB creates a new ProfileEnrichmentDB.
func NewProfileEnrichmentDB(db *gorm.DB) *ProfileEnrichmentDB {
	return &ProfileEnrichmentDB{db: db}
}

// Create inserts a new profile enrichment.
func (r *ProfileEnrichmentDB) Create(ctx context.Context, enrichment *domain.ProfileEnrichment) error {
	return r.db.WithContext(ctx).Create(enrichment).Error
}

// ByProfileID retrieves profile enrichment by profile ID.
func (r *ProfileEnrichmentDB) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.ProfileEnrichment, error) {
	var enrichment domain.ProfileEnrichment
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).First(&enrichment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, err
	}
	return &enrichment, nil
}

// Update updates an existing profile enrichment.
func (r *ProfileEnrichmentDB) Update(ctx context.Context, enrichment *domain.ProfileEnrichment) error {
	return r.db.WithContext(ctx).Save(enrichment).Error
}