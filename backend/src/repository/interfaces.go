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

// LoginHistoryRepository defines operations on LoginHistory entities.
type LoginHistoryRepository interface {
	Create(ctx context.Context, history *domain.LoginHistory) error
	ByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.LoginHistory, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// TwoFactorSettingsRepository defines operations on TwoFactorSettings entities.
type TwoFactorSettingsRepository interface {
	Create(ctx context.Context, settings *domain.TwoFactorSettings) error
	ByUserID(ctx context.Context, userID uuid.UUID) (*domain.TwoFactorSettings, error)
	Update(ctx context.Context, settings *domain.TwoFactorSettings) error
}

// PermissionRepository defines operations on Permission entities.
type PermissionRepository interface {
	Create(ctx context.Context, permission *domain.Permission) error
	ByKey(ctx context.Context, key string) (*domain.Permission, error)
	ListAll(ctx context.Context) ([]*domain.Permission, error)
}

// RoleRepository defines operations on Role entities.
type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	ByName(ctx context.Context, name string) (*domain.Role, error)
	ListAll(ctx context.Context) ([]*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RolePermissionRepository defines operations on RolePermission entities.
type RolePermissionRepository interface {
	Create(ctx context.Context, rp *domain.RolePermission) error
	ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error)
	ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error)
	Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error
	DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error
}

// ProfileRoleRepository defines operations on ProfileRole entities.
type ProfileRoleRepository interface {
	Create(ctx context.Context, pr *domain.ProfileRole) error
	ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error)
	ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error)
	Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error
	CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error)
	DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error
}