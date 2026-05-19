package repository

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

type SubmissionRepository interface {
	Create(ctx context.Context, submission *domain.Submission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	GetByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*domain.Submission, error)
	GetByEditorProfileID(ctx context.Context, editorProfileID uuid.UUID) ([]*domain.Submission, error)
	Update(ctx context.Context, submission *domain.Submission) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	FindNonDraftByEditorAndCampaign(ctx context.Context, editorProfileID, campaignID uuid.UUID) (*domain.Submission, error)
	ExistsByCampaignID(ctx context.Context, campaignID uuid.UUID) (bool, error)
}
