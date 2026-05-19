package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// CampaignRepository defines the interface for campaign data access.
type CampaignRepository interface {
	Create(ctx context.Context, campaign *domain.Campaign) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error)
	BySlug(ctx context.Context, slug string) (*domain.Campaign, error)
	ByBrandProfileID(ctx context.Context, brandProfileID uuid.UUID) ([]*domain.Campaign, error)
	Update(ctx context.Context, campaign *domain.Campaign) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, brandProfileID uuid.UUID, status string, page, pageSize int) ([]*domain.Campaign, int64, error)
	// ByStatusAndDeadline finds campaigns with given status where deadline has passed
	ByStatusAndDeadline(ctx context.Context, status domain.CampaignStatus, deadline time.Time) ([]*domain.Campaign, error)
}
