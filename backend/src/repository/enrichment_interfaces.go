package repository

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// ProfileEnrichmentRepository defines operations on ProfileEnrichment entities.
type ProfileEnrichmentRepository interface {
	Create(ctx context.Context, enrichment *domain.ProfileEnrichment) error
	ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.ProfileEnrichment, error)
	Update(ctx context.Context, enrichment *domain.ProfileEnrichment) error
}

// PortfolioItemRepository defines operations on PortfolioItem entities.
type PortfolioItemRepository interface {
	Create(ctx context.Context, item *domain.PortfolioItem) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.PortfolioItem, error)
	ByProfileID(ctx context.Context, profileID uuid.UUID, limit, offset int) ([]*domain.PortfolioItem, error)
	Update(ctx context.Context, item *domain.PortfolioItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error)
}

// AudienceDataRepository defines operations on AudienceData entities.
type AudienceDataRepository interface {
	Create(ctx context.Context, data *domain.AudienceData) error
	ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.AudienceData, error)
	Update(ctx context.Context, data *domain.AudienceData) error
}

// FollowerVerificationRepository defines operations on FollowerVerification entities.
type FollowerVerificationRepository interface {
	Create(ctx context.Context, verification *domain.FollowerVerification) error
	ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.FollowerVerification, error)
	Update(ctx context.Context, verification *domain.FollowerVerification) error
}

// PayoutPreferencesRepository defines operations on PayoutPreferences entities.
type PayoutPreferencesRepository interface {
	Create(ctx context.Context, prefs *domain.PayoutPreferences) error
	ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.PayoutPreferences, error)
	Update(ctx context.Context, prefs *domain.PayoutPreferences) error
}

// KYCStatusRepository defines operations on KYCStatus entities.
type KYCStatusRepository interface {
	Create(ctx context.Context, status *domain.KYCStatus) error
	ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.KYCStatus, error)
	Update(ctx context.Context, status *domain.KYCStatus) error
}