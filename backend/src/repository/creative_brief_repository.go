package repository

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// CreativeBriefRepository defines the interface for creative brief data access.
type CreativeBriefRepository interface {
	Create(ctx context.Context, brief *domain.CreativeBrief) error
	ByCampaignID(ctx context.Context, campaignID uuid.UUID) (*domain.CreativeBrief, error)
	Update(ctx context.Context, brief *domain.CreativeBrief) error
}
