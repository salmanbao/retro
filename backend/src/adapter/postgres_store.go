package adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// PostgresStore implements all repository interfaces using GORM.
type PostgresStore struct {
	db *gorm.DB
}

// NewPostgresStore creates a new PostgreSQL store.
func NewPostgresStore(db *gorm.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Connect creates a new GORM database connection.
func Connect(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// User repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateUser persists a new user record.
func (s *PostgresStore) CreateUser(ctx context.Context, user *domain.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

// UserByID retrieves a user by ID.
func (s *PostgresStore) UserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UserByEmail retrieves a user by email address.
func (s *PostgresStore) UserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user record.
func (s *PostgresStore) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.db.WithContext(ctx).Save(user).Error
}

// ─────────────────────────────────────────────────────────────────────────────
// Session repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateSession persists a new session record.
func (s *PostgresStore) CreateSession(ctx context.Context, session *domain.Session) error {
	return s.db.WithContext(ctx).Create(session).Error
}

// SessionByID retrieves a session by ID.
func (s *PostgresStore) SessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// SessionByTokenHash retrieves a session by its token hash.
func (s *PostgresStore) SessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	var session domain.Session
	err := s.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// SessionsByUserID retrieves all sessions for a user.
func (s *PostgresStore) SessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// UpdateSession updates an existing session record.
func (s *PostgresStore) UpdateSession(ctx context.Context, session *domain.Session) error {
	return s.db.WithContext(ctx).Save(session).Error
}

// DeleteSession removes a session by ID.
func (s *PostgresStore) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&domain.Session{}, "id = ?", id).Error
}

// DeleteSessionsByUserID removes all sessions for a user.
func (s *PostgresStore) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&domain.Session{}, "user_id = ?", userID).Error
}

// ─────────────────────────────────────────────────────────────────────────────
// Profile repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateProfile persists a new profile record.
func (s *PostgresStore) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	return s.db.WithContext(ctx).Create(profile).Error
}

// ProfileByID retrieves a profile by ID (excludes soft-deleted).
func (s *PostgresStore) ProfileByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	var profile domain.Profile
	err := s.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// ProfilesByUserID retrieves all non-deleted profiles for a user.
func (s *PostgresStore) ProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	var profiles []*domain.Profile
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at ASC").
		Find(&profiles).Error
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// UpdateProfile updates an existing profile record.
func (s *PostgresStore) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	return s.db.WithContext(ctx).Save(profile).Error
}

// DeleteProfile soft-deletes a profile by ID.
func (s *PostgresStore) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&domain.Profile{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// ─────────────────────────────────────────────────────────────────────────────
// Token repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateToken persists a new auth token record.
func (s *PostgresStore) CreateToken(ctx context.Context, token *domain.AuthToken) error {
	return s.db.WithContext(ctx).Create(token).Error
}

// TokenByTokenHash retrieves a token by its hash.
func (s *PostgresStore) TokenByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthToken, error) {
	var token domain.AuthToken
	err := s.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// TokenByUserIDAndType retrieves the most recent unused token of a given type for a user.
func (s *PostgresStore) TokenByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.AuthToken, error) {
	var token domain.AuthToken
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND token_type = ? AND used_at IS NULL AND expires_at > ?", userID, tokenType, time.Now()).
		Order("created_at DESC").
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateToken marks a token as used.
func (s *PostgresStore) UpdateToken(ctx context.Context, token *domain.AuthToken) error {
	return s.db.WithContext(ctx).Save(token).Error
}

// DeleteTokensByUserIDAndType removes all tokens of a given type for a user.
func (s *PostgresStore) DeleteTokensByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	return s.db.WithContext(ctx).
		Delete(&domain.AuthToken{}, "user_id = ? AND token_type = ?", userID, tokenType).Error
}

// ─────────────────────────────────────────────────────────────────────────────
// Repository interface wrappers
// ─────────────────────────────────────────────────────────────────────────────

// userRepo adapts PostgresStore to satisfy repository.UserRepository.
type userRepo struct{ *PostgresStore }

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	return r.CreateUser(ctx, user)
}
func (r *userRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.UserByID(ctx, id)
}
func (r *userRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.UserByEmail(ctx, email)
}
func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	return r.UpdateUser(ctx, user)
}

// sessionRepo adapts PostgresStore to satisfy repository.SessionRepository.
type sessionRepo struct{ *PostgresStore }

func (r *sessionRepo) Create(ctx context.Context, session *domain.Session) error {
	return r.CreateSession(ctx, session)
}
func (r *sessionRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return r.SessionByID(ctx, id)
}
func (r *sessionRepo) ByTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	return r.SessionByTokenHash(ctx, hash)
}
func (r *sessionRepo) ByUserID(ctx context.Context, uid uuid.UUID) ([]*domain.Session, error) {
	return r.SessionsByUserID(ctx, uid)
}
func (r *sessionRepo) Update(ctx context.Context, session *domain.Session) error {
	return r.UpdateSession(ctx, session)
}
func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DeleteSession(ctx, id)
}
func (r *sessionRepo) DeleteByUserID(ctx context.Context, uid uuid.UUID) error {
	return r.DeleteSessionsByUserID(ctx, uid)
}

// profileRepo adapts PostgresStore to satisfy repository.ProfileRepository.
type profileRepo struct{ *PostgresStore }

func (r *profileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	return r.CreateProfile(ctx, profile)
}
func (r *profileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	return r.ProfileByID(ctx, id)
}
func (r *profileRepo) ByUserID(ctx context.Context, uid uuid.UUID) ([]*domain.Profile, error) {
	return r.ProfilesByUserID(ctx, uid)
}
func (r *profileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	return r.UpdateProfile(ctx, profile)
}
func (r *profileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DeleteProfile(ctx, id)
}

// tokenRepo adapts PostgresStore to satisfy repository.TokenRepository.
type tokenRepo struct{ *PostgresStore }

func (r *tokenRepo) Create(ctx context.Context, token *domain.AuthToken) error {
	return r.CreateToken(ctx, token)
}
func (r *tokenRepo) ByTokenHash(ctx context.Context, hash string) (*domain.AuthToken, error) {
	return r.TokenByTokenHash(ctx, hash)
}
func (r *tokenRepo) ByUserIDAndType(ctx context.Context, uid uuid.UUID, tt domain.TokenType) (*domain.AuthToken, error) {
	return r.TokenByUserIDAndType(ctx, uid, tt)
}
func (r *tokenRepo) Update(ctx context.Context, token *domain.AuthToken) error {
	return r.UpdateToken(ctx, token)
}
func (r *tokenRepo) DeleteByUserIDAndType(ctx context.Context, uid uuid.UUID, tt domain.TokenType) error {
	return r.DeleteTokensByUserIDAndType(ctx, uid, tt)
}

// Compile-time interface checks.
var (
	_ repository.UserRepository   = (*userRepo)(nil)
	_ repository.SessionRepository = (*sessionRepo)(nil)
	_ repository.ProfileRepository = (*profileRepo)(nil)
	_ repository.TokenRepository   = (*tokenRepo)(nil)
)

// UserRepository returns a repository.UserRepository backed by PostgresStore.
func (s *PostgresStore) UserRepository() repository.UserRepository { return &userRepo{s} }

// SessionRepository returns a repository.SessionRepository backed by PostgresStore.
func (s *PostgresStore) SessionRepository() repository.SessionRepository { return &sessionRepo{s} }

// ProfileRepository returns a repository.ProfileRepository backed by PostgresStore.
func (s *PostgresStore) ProfileRepository() repository.ProfileRepository { return &profileRepo{s} }

// TokenRepository returns a repository.TokenRepository backed by PostgresStore.
func (s *PostgresStore) TokenRepository() repository.TokenRepository { return &tokenRepo{s} }

// DB returns the underlying GORM DB instance for advanced operations.
func (s *PostgresStore) DB() *gorm.DB {
	return s.db
}

// AutoMigrate runs GORM automigration for all domain models.
func (s *PostgresStore) AutoMigrate() error {
	return s.db.AutoMigrate(&domain.User{}, &domain.Session{}, &domain.Profile{}, &domain.AuthToken{})
}

// Transaction executes a function within a database transaction.
func (s *PostgresStore) Transaction(ctx context.Context, fn func(repo repository.UserRepository) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new PostgresStore with the transaction
		txStore := &PostgresStore{db: tx}
		return fn(txStore.UserRepository())
	})
}