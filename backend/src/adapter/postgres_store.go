package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// PostgresStore implements all repository interfaces using PostgreSQL.
// Each entity's repository methods are named with entity-specific prefixes to avoid
// method signature collisions across the four interfaces.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a new PostgreSQL store.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// Connect returns a new pgxpool connection pool.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return pool, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// User repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateUser persists a new user record.
func (s *PostgresStore) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.PasswordHash, user.Verified, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// UserByID retrieves a user by ID.
func (s *PostgresStore) UserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, verified, created_at, updated_at FROM users WHERE id = $1`,
		id,
	)
	return scanUser(row)
}

// UserByEmail retrieves a user by email address.
func (s *PostgresStore) UserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, verified, created_at, updated_at FROM users WHERE email = $1`,
		email,
	)
	return scanUser(row)
}

// UpdateUser updates an existing user record.
func (s *PostgresStore) UpdateUser(ctx context.Context, user *domain.User) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET email = $2, password_hash = $3, verified = $4, updated_at = $5 WHERE id = $1`,
		user.ID, user.Email, user.PasswordHash, user.Verified, user.UpdatedAt,
	)
	return err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Verified, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Session repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateSession persists a new session record.
func (s *PostgresStore) CreateSession(ctx context.Context, session *domain.Session) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, active_profile_id, token_hash, user_agent, ip_address, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		session.ID, session.UserID, session.ActiveProfileID, session.TokenHash,
		session.UserAgent, session.IPAddress, session.ExpiresAt, session.CreatedAt,
	)
	return err
}

// SessionByID retrieves a session by ID.
func (s *PostgresStore) SessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, user_id, active_profile_id, token_hash, user_agent, ip_address, expires_at, created_at
		 FROM sessions WHERE id = $1`,
		id,
	)
	return scanSession(row)
}

// SessionByTokenHash retrieves a session by its token hash.
func (s *PostgresStore) SessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, user_id, active_profile_id, token_hash, user_agent, ip_address, expires_at, created_at
		 FROM sessions WHERE token_hash = $1`,
		tokenHash,
	)
	return scanSession(row)
}

// SessionsByUserID retrieves all sessions for a user.
func (s *PostgresStore) SessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, active_profile_id, token_hash, user_agent, ip_address, expires_at, created_at
		 FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session, err := scanSessionRows(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

// UpdateSession updates an existing session record.
func (s *PostgresStore) UpdateSession(ctx context.Context, session *domain.Session) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET active_profile_id = $2, token_hash = $3, user_agent = $4, ip_address = $5, expires_at = $6 WHERE id = $1`,
		session.ID, session.ActiveProfileID, session.TokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt,
	)
	return err
}

// DeleteSession removes a session by ID.
func (s *PostgresStore) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

// DeleteSessionsByUserID removes all sessions for a user.
func (s *PostgresStore) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func scanSession(row pgx.Row) (*domain.Session, error) {
	var sess domain.Session
	err := row.Scan(&sess.ID, &sess.UserID, &sess.ActiveProfileID, &sess.TokenHash,
		&sess.UserAgent, &sess.IPAddress, &sess.ExpiresAt, &sess.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func scanSessionRows(rows pgx.Rows) (*domain.Session, error) {
	var sess domain.Session
	err := rows.Scan(&sess.ID, &sess.UserID, &sess.ActiveProfileID, &sess.TokenHash,
		&sess.UserAgent, &sess.IPAddress, &sess.ExpiresAt, &sess.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Profile repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateProfile persists a new profile record.
func (s *PostgresStore) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO profiles (id, user_id, profile_type, name, details, created_at, updated_at, deleted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		profile.ID, profile.UserID, profile.Type, profile.Name, profile.Details,
		profile.CreatedAt, profile.UpdatedAt, profile.DeletedAt,
	)
	return err
}

// ProfileByID retrieves a profile by ID (excludes soft-deleted).
func (s *PostgresStore) ProfileByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, user_id, profile_type, name, details, created_at, updated_at, deleted_at
		 FROM profiles WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	return scanProfile(row)
}

// ProfilesByUserID retrieves all non-deleted profiles for a user.
func (s *PostgresStore) ProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, profile_type, name, details, created_at, updated_at, deleted_at
		 FROM profiles WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*domain.Profile
	for rows.Next() {
		profile, err := scanProfileRows(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

// UpdateProfile updates an existing profile record.
func (s *PostgresStore) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE profiles SET name = $2, details = $3, updated_at = $4 WHERE id = $1`,
		profile.ID, profile.Name, profile.Details, profile.UpdatedAt,
	)
	return err
}

// DeleteProfile soft-deletes a profile by ID.
func (s *PostgresStore) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE profiles SET deleted_at = $2 WHERE id = $1`,
		id, time.Now(),
	)
	return err
}

func scanProfile(row pgx.Row) (*domain.Profile, error) {
	var p domain.Profile
	var details []byte
	err := row.Scan(&p.ID, &p.UserID, &p.Type, &p.Name, &details, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Details = json.RawMessage(details)
	return &p, nil
}

func scanProfileRows(rows pgx.Rows) (*domain.Profile, error) {
	var p domain.Profile
	var details []byte
	err := rows.Scan(&p.ID, &p.UserID, &p.Type, &p.Name, &details, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	p.Details = json.RawMessage(details)
	return &p, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Token repository methods
// ─────────────────────────────────────────────────────────────────────────────

// CreateToken persists a new auth token record.
func (s *PostgresStore) CreateToken(ctx context.Context, token *domain.AuthToken) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO auth_tokens (id, user_id, token_type, token_hash, expires_at, used_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		token.ID, token.UserID, token.TokenType, token.TokenHash, token.ExpiresAt, token.UsedAt, token.CreatedAt,
	)
	return err
}

// TokenByTokenHash retrieves a token by its hash.
func (s *PostgresStore) TokenByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthToken, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, user_id, token_type, token_hash, expires_at, used_at, created_at
		 FROM auth_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	return scanToken(row)
}

// TokenByUserIDAndType retrieves the most recent unused token of a given type for a user.
func (s *PostgresStore) TokenByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.AuthToken, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, user_id, token_type, token_hash, expires_at, used_at, created_at
		 FROM auth_tokens WHERE user_id = $1 AND token_type = $2 AND used_at IS NULL AND expires_at > NOW()
		 ORDER BY created_at DESC LIMIT 1`,
		userID, tokenType,
	)
	return scanToken(row)
}

// UpdateToken marks a token as used.
func (s *PostgresStore) UpdateToken(ctx context.Context, token *domain.AuthToken) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE auth_tokens SET used_at = $2 WHERE id = $1`,
		token.ID, token.UsedAt,
	)
	return err
}

// DeleteTokensByUserIDAndType removes all tokens of a given type for a user.
func (s *PostgresStore) DeleteTokensByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM auth_tokens WHERE user_id = $1 AND token_type = $2`,
		userID, tokenType,
	)
	return err
}

func scanToken(row pgx.Row) (*domain.AuthToken, error) {
	var t domain.AuthToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenType, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
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
	_ repository.UserRepository    = (*userRepo)(nil)
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
