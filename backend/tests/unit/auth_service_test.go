package unit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"viralforge/backend/src/adapter"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
	"viralforge/backend/src/service"
)

// mockUserRepo implements repository.UserRepository for testing.
type mockUserRepo struct {
	users   map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (r *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, exists := r.byEmail[user.Email]; exists {
		return domain.ErrEmailAlreadyRegistered
	}
	r.users[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *mockUserRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	r.users[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

// mockSessionRepo implements repository.SessionRepository for testing.
type mockSessionRepo struct {
	sessions map[uuid.UUID]*domain.Session
	byHash   map[string]*domain.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[uuid.UUID]*domain.Session),
		byHash:   make(map[string]*domain.Session),
	}
}

func (r *mockSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	r.byHash[session.TokenHash] = session
	return nil
}

func (r *mockSessionRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSessionRepo) ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	if s, ok := r.byHash[tokenHash]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSessionRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *mockSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	if _, ok := r.sessions[session.ID]; !ok {
		return domain.ErrSessionNotFound
	}
	r.sessions[session.ID] = session
	r.byHash[session.TokenHash] = session
	return nil
}

func (r *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if s, ok := r.sessions[id]; ok {
		delete(r.byHash, s.TokenHash)
		delete(r.sessions, id)
	}
	return nil
}

func (r *mockSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.byHash, s.TokenHash)
			delete(r.sessions, id)
		}
	}
	return nil
}

// mockTokenRepo implements repository.TokenRepository for testing.
type mockTokenRepo struct {
	tokens map[uuid.UUID]*domain.AuthToken
	byHash map[string]*domain.AuthToken
	byUser map[uuid.UUID][]*domain.AuthToken
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		tokens: make(map[uuid.UUID]*domain.AuthToken),
		byHash: make(map[string]*domain.AuthToken),
		byUser: make(map[uuid.UUID][]*domain.AuthToken),
	}
}

func (r *mockTokenRepo) Create(ctx context.Context, token *domain.AuthToken) error {
	r.tokens[token.ID] = token
	r.byHash[token.TokenHash] = token
	r.byUser[token.UserID] = append(r.byUser[token.UserID], token)
	return nil
}

func (r *mockTokenRepo) ByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthToken, error) {
	if t, ok := r.byHash[tokenHash]; ok {
		return t, nil
	}
	return nil, domain.ErrTokenNotFound
}

func (r *mockTokenRepo) ByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.AuthToken, error) {
	for _, t := range r.byUser[userID] {
		if t.TokenType == tokenType && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, domain.ErrTokenNotFound
}

func (r *mockTokenRepo) Update(ctx context.Context, token *domain.AuthToken) error {
	if _, ok := r.tokens[token.ID]; !ok {
		return domain.ErrTokenNotFound
	}
	r.tokens[token.ID] = token
	r.byHash[token.TokenHash] = token
	return nil
}

func (r *mockTokenRepo) DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	delete(r.byUser, userID)
	return nil
}

// mockEmailService implements adapter.EmailService for testing.
type mockEmailService struct {
	sent []mockEmail
}

type mockEmail struct {
	To      string
	Subject string
	Body    string
}

func (m *mockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	m.sent = append(m.sent, mockEmail{To: to, Subject: subject, Body: body})
	return nil
}

func generateTokenHash() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// testableAuthService wraps AuthService for testing.
type testableAuthService struct {
	*service.AuthService
}

func newTestableAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, tokenRepo repository.TokenRepository, emailSvc adapter.EmailService, baseURL string) *testableAuthService {
	svc := service.NewAuthService(userRepo, sessionRepo, tokenRepo, emailSvc, baseURL)
	return &testableAuthService{svc}
}

// TestLogin tests the Login method.
func TestLogin(t *testing.T) {
	ctx := context.Background()
	userRepo := newMockUserRepo()
	sessionRepo := newMockSessionRepo()
	tokenRepo := newMockTokenRepo()
	emailSvc := &mockEmailService{}

	svc := newTestableAuthService(userRepo, sessionRepo, tokenRepo, emailSvc, "http://localhost:3000")

	t.Run("successful login", func(t *testing.T) {
		password := "Password123!"
		passwordHash, _ := hashPassword(password)
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "login@example.com",
			PasswordHash: passwordHash,
			Verified:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		userRepo.Create(ctx, user)

		session, err := svc.Login(ctx, "login@example.com", password, "TestAgent", "127.0.0.1")
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, user.ID, session.UserID)
		assert.NotEmpty(t, session.TokenHash)
		assert.False(t, session.IsExpired())
	})

	t.Run("invalid email format", func(t *testing.T) {
		_, err := svc.Login(ctx, "not-an-email", "Password123!", "", "")
		assert.ErrorIs(t, err, domain.ErrInvalidEmailFormat)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := svc.Login(ctx, "nonexistent@example.com", "Password123!", "", "")
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("wrong password", func(t *testing.T) {
		passwordHash, _ := hashPassword("CorrectPassword123!")
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "wrongpass@example.com",
			PasswordHash: passwordHash,
			Verified:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		userRepo.Create(ctx, user)

		_, err := svc.Login(ctx, "wrongpass@example.com", "WrongPassword123!", "", "")
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("email not verified", func(t *testing.T) {
		passwordHash, _ := hashPassword("Password123!")
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "unverified@example.com",
			PasswordHash: passwordHash,
			Verified:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		userRepo.Create(ctx, user)

		_, err := svc.Login(ctx, "unverified@example.com", "Password123!", "", "")
		assert.ErrorIs(t, err, domain.ErrEmailNotVerified)
	})
}

// TestLogout tests the Logout method.
func TestLogout(t *testing.T) {
	ctx := context.Background()
	userRepo := newMockUserRepo()
	sessionRepo := newMockSessionRepo()
	tokenRepo := newMockTokenRepo()
	emailSvc := &mockEmailService{}

	svc := newTestableAuthService(userRepo, sessionRepo, tokenRepo, emailSvc, "http://localhost:3000")

	t.Run("successful logout", func(t *testing.T) {
		user := &domain.User{
			ID:        uuid.New(),
			Email:     "logout@example.com",
			Verified:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		userRepo.Create(ctx, user)

		token := generateTokenHash()
		session := domain.NewSession(user.ID, token, "TestAgent", "127.0.0.1", time.Now().Add(24*time.Hour))
		sessionRepo.Create(ctx, session)

		err := svc.Logout(ctx, token)
		require.NoError(t, err)

		_, err = sessionRepo.ByTokenHash(ctx, token)
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("session not found", func(t *testing.T) {
		err := svc.Logout(ctx, "nonexistent-token")
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})
}
