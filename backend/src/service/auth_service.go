package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
	"viralforge/backend/src/adapter"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

const bcryptCost = 12
const sessionDuration = 24 * time.Hour

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// AuthService handles user authentication operations.
type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	tokenRepo   repository.TokenRepository
	emailSvc    adapter.EmailService
	baseURL     string
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, tokenRepo repository.TokenRepository, emailSvc adapter.EmailService, baseURL string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokenRepo:   tokenRepo,
		emailSvc:    emailSvc,
		baseURL:     baseURL,
	}
}

// Register creates a new user account and sends a verification email.
func (s *AuthService) Register(ctx context.Context, email, password string) error {
	if !isValidEmail(email) {
		return domain.ErrInvalidEmailFormat
	}
	if !isValidPassword(password) {
		return domain.ErrInvalidPasswordFormat
	}

	existing, _ := s.userRepo.ByEmail(ctx, email)
	if existing != nil {
		return domain.ErrEmailAlreadyRegistered
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.NewUser(email, passwordHash)
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	verificationToken := domain.NewAuthToken(user.ID, domain.TokenTypeVerification, generateSecureToken(), time.Now().Add(24*time.Hour))
	if err := s.tokenRepo.Create(ctx, verificationToken); err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.baseURL, verificationToken.TokenHash)
	emailBody := fmt.Sprintf("Click the link to verify your email: %s", verificationURL)
	if err := s.emailSvc.SendEmail(ctx, email, "Verify your email", emailBody); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// VerifyEmail verifies a user's email address using the provided token.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	authToken, err := s.tokenRepo.ByTokenHash(ctx, token)
	if err != nil {
		return err
	}

	if authToken.IsExpired() {
		return domain.ErrTokenExpired
	}
	if authToken.IsUsed() {
		return domain.ErrTokenAlreadyUsed
	}

	user, err := s.userRepo.ByID(ctx, authToken.UserID)
	if err != nil {
		return err
	}

	user.Verify()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	authToken.MarkUsed()
	if err := s.tokenRepo.Update(ctx, authToken); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// Login authenticates a user and creates a new session.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*domain.Session, error) {
	if !isValidEmail(email) {
		return nil, domain.ErrInvalidEmailFormat
	}

	user, err := s.userRepo.ByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.Verified {
		return nil, domain.ErrEmailNotVerified
	}

	if !verifyPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token := generateSecureToken()
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(sessionDuration)

	session := domain.NewSession(user.ID, tokenHash, userAgent, ipAddress, expiresAt)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// Logout invalidates the current session.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	session, err := s.sessionRepo.ByTokenHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrSessionNotFound
	}

	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email) && len(email) <= 255
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c == '!' || c == '@' || c == '#' || c == '$' || c == '%' || c == '^' || c == '&' || c == '*' || c == '(' || c == ')' || c == '-' || c == '+' || c == '=' || c == '[' || c == ']' || c == '{' || c == '}' || c == '|' || c == ';' || c == ':' || c == '"' || c == '\'' || c == ',' || c == '.' || c == '/' || c == '<' || c == '>' || c == '?':
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hash), nil
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	// In production, use crypto/sha256
	return token
}
