package repository

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// UserRepository defines operations on User entities.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// SessionRepository defines operations on Session entities.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// ProfileRepository defines operations on Profile entities.
type ProfileRepository interface {
	Create(ctx context.Context, profile *domain.Profile) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error)
	Update(ctx context.Context, profile *domain.Profile) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TokenRepository defines operations on AuthToken entities.
type TokenRepository interface {
	Create(ctx context.Context, token *domain.AuthToken) error
	ByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthToken, error)
	ByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.AuthToken, error)
	Update(ctx context.Context, token *domain.AuthToken) error
	DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error
}

// CampaignRepository defines operations on Campaign entities.
type CampaignRepository interface {
	Create(ctx context.Context, campaign *domain.Campaign) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error)
	ByBrandProfileID(ctx context.Context, brandProfileID uuid.UUID) ([]*domain.Campaign, error)
	Update(ctx context.Context, campaign *domain.Campaign) error
}
