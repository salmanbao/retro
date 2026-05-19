package repository

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// CampaignAssetRepository defines the interface for campaign asset data access.
type CampaignAssetRepository interface {
	Create(ctx context.Context, asset *domain.CampaignAsset) error
	ByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*domain.CampaignAsset, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCampaignID(ctx context.Context, campaignID uuid.UUID) error
}
