package repository

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// AssetMetadataRepository defines the interface for asset metadata data access.
type AssetMetadataRepository interface {
	Create(ctx context.Context, asset *domain.AssetMetadata) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.AssetMetadata, error)
	ListByCampaign(ctx context.Context, campaignID uuid.UUID, page, pageSize int) ([]*domain.AssetMetadata, int64, error)
	Update(ctx context.Context, asset *domain.AssetMetadata) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ByCampaignAndFilename(ctx context.Context, campaignID uuid.UUID, filename string) (*domain.AssetMetadata, error)
	ListVersions(ctx context.Context, campaignID uuid.UUID, filename string) ([]*domain.AssetMetadata, error)
}
