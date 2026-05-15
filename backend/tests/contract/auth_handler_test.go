package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/service"
)

// mockUserRepo implements repository.UserRepository for contract testing.
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

// mockSessionRepo implements repository.SessionRepository for contract testing.
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

// mockTokenRepo implements repository.TokenRepository for contract testing.
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

// mockEmailSvc implements adapter.EmailService for contract testing.
type mockEmailSvc struct {
	sent []mockEmail
}

type mockEmail struct {
	To      string
	Subject string
	Body    string
}

func (m *mockEmailSvc) SendEmail(ctx context.Context, to, subject, body string) error {
	m.sent = append(m.sent, mockEmail{To: to, Subject: subject, Body: body})
	return nil
}

// setupTestHandler creates a handler configured for testing.
func setupTestHandler() (*handler.AuthHandler, *mockUserRepo, *mockTokenRepo, *mockEmailSvc) {
	userRepo := newMockUserRepo()
	sessionRepo := newMockSessionRepo()
	tokenRepo := newMockTokenRepo()
	emailSvc := &mockEmailSvc{}
	authSvc := service.NewAuthService(userRepo, sessionRepo, tokenRepo, emailSvc, "http://localhost:8080")
	authHandler := handler.NewAuthHandler(authSvc, nil, nil)
	return authHandler, userRepo, tokenRepo, emailSvc
}

// TestContractRegisterEndpoint tests POST /api/v1/auth/register contract.
func TestContractRegisterEndpoint(t *testing.T) {
	t.Run("201 Created - valid registration", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.RegisterRequest{Email: "test@example.com", Password: "Password123!"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "Account created. Please check your email to verify your address.", resp["message"])
	})

	t.Run("400 Bad Request - invalid JSON", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "invalid_request", resp.Error)
	})

	t.Run("400 Bad Request - missing email", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.RegisterRequest{Email: "", Password: "Password123!"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "validation_error", resp.Error)
	})

	t.Run("400 Bad Request - missing password", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.RegisterRequest{Email: "test@example.com", Password: ""}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("409 Conflict - email already registered", func(t *testing.T) {
		h, userRepo, _, _ := setupTestHandler()

		existing := domain.NewUser("test@example.com", "hash")
		userRepo.Create(context.Background(), existing)

		reqBody := handler.RegisterRequest{Email: "test@example.com", Password: "Password123!"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "email_exists", resp.Error)
	})
}

// TestContractVerifyEmailEndpoint tests POST /api/v1/auth/verify-email contract.
func TestContractVerifyEmailEndpoint(t *testing.T) {
	t.Run("200 OK - valid token", func(t *testing.T) {
		h, userRepo, tokenRepo, _ := setupTestHandler()

		user := domain.NewUser("verify@example.com", "hash")
		userRepo.Create(context.Background(), user)

		tokenHash := "valid-token-hash"
		token := domain.NewAuthToken(user.ID, domain.TokenTypeVerification, tokenHash, time.Now().Add(24*time.Hour))
		tokenRepo.Create(context.Background(), token)

		reqBody := handler.VerifyEmailRequest{Token: tokenHash}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "Email verified successfully", resp["message"])
	})

	t.Run("400 Bad Request - invalid JSON", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - missing token", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.VerifyEmailRequest{Token: ""}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - token not found", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.VerifyEmailRequest{Token: "nonexistent-token"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "invalid_token", resp.Error)
	})
}

// TestContractForgotPasswordEndpoint tests POST /api/v1/auth/forgot-password contract.
func TestContractForgotPasswordEndpoint(t *testing.T) {
	t.Run("200 OK - email exists, reset sent", func(t *testing.T) {
		h, userRepo, _, emailSvc := setupTestHandler()

		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), 12)
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "reset-test@example.com",
			PasswordHash: string(passwordHash),
			Verified:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		userRepo.Create(context.Background(), user)

		reqBody := handler.ForgotPasswordRequest{Email: "reset-test@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ForgotPassword(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Contains(t, resp["message"], "reset link has been sent")
		// Verify email was sent
		assert.Len(t, emailSvc.sent, 1)
	})

	t.Run("200 OK - email not found returns same response (security)", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.ForgotPasswordRequest{Email: "nonexistent@example.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ForgotPassword(w, req)

		// Should return same 200 response to prevent email enumeration
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Contains(t, resp["message"], "reset link has been sent")
	})

	t.Run("400 Bad Request - invalid JSON", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ForgotPassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - missing email", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.ForgotPasswordRequest{Email: ""}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ForgotPassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestContractResetPasswordEndpoint tests POST /api/v1/auth/reset-password contract.
func TestContractResetPasswordEndpoint(t *testing.T) {
	t.Run("200 OK - valid token and password", func(t *testing.T) {
		h, userRepo, tokenRepo, _ := setupTestHandler()

		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("OldPassword123!"), 12)
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "reset-confirm@example.com",
			PasswordHash: string(passwordHash),
			Verified:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		userRepo.Create(context.Background(), user)

		tokenHash := "valid-reset-token"
		token := domain.NewAuthToken(user.ID, domain.TokenTypePasswordReset, tokenHash, time.Now().Add(1*time.Hour))
		tokenRepo.Create(context.Background(), token)

		reqBody := handler.ResetPasswordRequest{Token: tokenHash, NewPassword: "NewPassword123!"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ResetPassword(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Contains(t, resp["message"], "Password reset successfully")
	})

	t.Run("400 Bad Request - missing token", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.ResetPasswordRequest{Token: "", NewPassword: "NewPassword123!"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ResetPassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - missing new password", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.ResetPasswordRequest{Token: "some-token", NewPassword: ""}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ResetPassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - invalid password format", func(t *testing.T) {
		h, userRepo, tokenRepo, _ := setupTestHandler()

		user := domain.NewUser("weak@example.com", "hash")
		userRepo.Create(context.Background(), user)

		tokenHash := "weak-token"
		token := domain.NewAuthToken(user.ID, domain.TokenTypePasswordReset, tokenHash, time.Now().Add(1*time.Hour))
		tokenRepo.Create(context.Background(), token)

		reqBody := handler.ResetPasswordRequest{Token: tokenHash, NewPassword: "weak"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ResetPassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "invalid_password", resp.Error)
	})

	t.Run("400 Bad Request - invalid token", func(t *testing.T) {
		h, _, _, _ := setupTestHandler()

		reqBody := handler.ResetPasswordRequest{Token: "nonexistent-token", NewPassword: "NewPassword123!"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ResetPassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "invalid_token", resp.Error)
	})
}
