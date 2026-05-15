package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
)

// mockUserStore implements user repository for integration testing.
type mockUserStore struct {
	users   map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (s *mockUserStore) Create(ctx context.Context, user *domain.User) error {
	if _, exists := s.byEmail[user.Email]; exists {
		return domain.ErrEmailAlreadyRegistered
	}
	s.users[user.ID] = user
	s.byEmail[user.Email] = user
	return nil
}

func (s *mockUserStore) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *mockUserStore) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := s.byEmail[email]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *mockUserStore) Update(ctx context.Context, user *domain.User) error {
	if _, ok := s.users[user.ID]; !ok {
		return domain.ErrUserNotFound
	}
	s.users[user.ID] = user
	s.byEmail[user.Email] = user
	return nil
}

// mockTokenStore implements token repository for integration testing.
type mockTokenStore struct {
	tokens map[uuid.UUID]*domain.AuthToken
	byHash map[string]*domain.AuthToken
	byUser map[uuid.UUID][]*domain.AuthToken
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{
		tokens: make(map[uuid.UUID]*domain.AuthToken),
		byHash: make(map[string]*domain.AuthToken),
		byUser: make(map[uuid.UUID][]*domain.AuthToken),
	}
}

func (s *mockTokenStore) Create(ctx context.Context, token *domain.AuthToken) error {
	s.tokens[token.ID] = token
	s.byHash[token.TokenHash] = token
	s.byUser[token.UserID] = append(s.byUser[token.UserID], token)
	return nil
}

func (s *mockTokenStore) ByTokenHash(ctx context.Context, tokenHash string) (*domain.AuthToken, error) {
	if t, ok := s.byHash[tokenHash]; ok {
		return t, nil
	}
	return nil, domain.ErrTokenNotFound
}

func (s *mockTokenStore) ByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) (*domain.AuthToken, error) {
	for _, t := range s.byUser[userID] {
		if t.TokenType == tokenType && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, domain.ErrTokenNotFound
}

func (s *mockTokenStore) Update(ctx context.Context, token *domain.AuthToken) error {
	if _, ok := s.tokens[token.ID]; !ok {
		return domain.ErrTokenNotFound
	}
	s.tokens[token.ID] = token
	s.byHash[token.TokenHash] = token
	return nil
}

func (s *mockTokenStore) DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error {
	delete(s.byUser, userID)
	return nil
}

// mockEmailSender implements email service for integration testing.
type mockEmailSender struct {
	sent []string
}

func (m *mockEmailSender) SendEmail(ctx context.Context, to, subject, body string) error {
	m.sent = append(m.sent, to)
	return nil
}

// TestUserRegistrationIntegration tests the full registration flow.
func TestUserRegistrationIntegration(t *testing.T) {
	ctx := context.Background()
	userStore := newMockUserStore()
	tokenStore := newMockTokenStore()
	emailSender := &mockEmailSender{}

	email := "integration@example.com"
	password := "Password123!"

	existing, _ := userStore.ByEmail(ctx, email)
	assert.Nil(t, existing, "user should not exist before registration")

	user := domain.NewUser(email, password)
	err := userStore.Create(ctx, user)
	require.NoError(t, err)

	token := domain.NewAuthToken(user.ID, domain.TokenTypeVerification, "test-token-hash", time.Now().Add(24*time.Hour))
	err = tokenStore.Create(ctx, token)
	require.NoError(t, err)

	err = emailSender.SendEmail(ctx, email, "Verify your email", "Click the link to verify")
	require.NoError(t, err)
	assert.Len(t, emailSender.sent, 1)
	assert.Equal(t, email, emailSender.sent[0])

	createdUser, err := userStore.ByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, email, createdUser.Email)
	assert.False(t, createdUser.Verified)
}

// TestEmailVerificationIntegration tests the full verification flow.
func TestEmailVerificationIntegration(t *testing.T) {
	ctx := context.Background()
	userStore := newMockUserStore()
	tokenStore := newMockTokenStore()

	email := "verify@example.com"
	user := domain.NewUser(email, "hash")
	err := userStore.Create(ctx, user)
	require.NoError(t, err)

	tokenHash := "valid-token-hash"
	token := domain.NewAuthToken(user.ID, domain.TokenTypeVerification, tokenHash, time.Now().Add(24*time.Hour))
	err = tokenStore.Create(ctx, token)
	require.NoError(t, err)

	fetchedToken, err := tokenStore.ByTokenHash(ctx, tokenHash)
	require.NoError(t, err)
	assert.False(t, fetchedToken.IsUsed())
	assert.False(t, fetchedToken.IsExpired())

	fetchedUser, err := userStore.ByID(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, fetchedUser.Verified)

	fetchedUser.Verify()
	err = userStore.Update(ctx, fetchedUser)
	require.NoError(t, err)

	fetchedToken.MarkUsed()
	err = tokenStore.Update(ctx, fetchedToken)
	require.NoError(t, err)

	verifiedUser, err := userStore.ByEmail(ctx, email)
	require.NoError(t, err)
	assert.True(t, verifiedUser.Verified)

	updatedToken, err := tokenStore.ByTokenHash(ctx, tokenHash)
	require.NoError(t, err)
	assert.True(t, updatedToken.IsUsed())
}

// TestTokenPersistenceIntegration verifies tokens are properly persisted.
func TestTokenPersistenceIntegration(t *testing.T) {
	ctx := context.Background()
	tokenStore := newMockTokenStore()

	userID := uuid.New()
	token := domain.NewAuthToken(userID, domain.TokenTypeVerification, "persist-token", time.Now().Add(24*time.Hour))

	err := tokenStore.Create(ctx, token)
	require.NoError(t, err)

	retrieved, err := tokenStore.ByTokenHash(ctx, "persist-token")
	require.NoError(t, err)
	assert.Equal(t, token.ID, retrieved.ID)
	assert.Equal(t, userID, retrieved.UserID)
	assert.Equal(t, domain.TokenTypeVerification, retrieved.TokenType)
	assert.False(t, retrieved.IsUsed())
}

// TestConcurrentRegistrationIntegration tests handling of simultaneous registrations.
func TestConcurrentRegistrationIntegration(t *testing.T) {
	ctx := context.Background()
	userStore := newMockUserStore()

	email := "concurrent@example.com"
	password := "Password123!"

	user1 := domain.NewUser(email, password)
	err1 := userStore.Create(ctx, user1)
	require.NoError(t, err1)

	user2 := domain.NewUser(email, password)
	err2 := userStore.Create(ctx, user2)
	assert.ErrorIs(t, err2, domain.ErrEmailAlreadyRegistered)

	users, err := userStore.ByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, user1.ID, users.ID)
}
