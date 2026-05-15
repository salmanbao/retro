package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

func TestNewUser(t *testing.T) {
	email := "test@example.com"
	hash := "$2a$12$hashedpassword"
	user := domain.NewUser(email, hash)

	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, hash, user.PasswordHash)
	assert.False(t, user.Verified)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestUserVerify(t *testing.T) {
	user := domain.NewUser("test@example.com", "hash")
	assert.False(t, user.Verified)

	user.Verify()
	assert.True(t, user.Verified)
	assert.True(t, user.UpdatedAt.After(user.CreatedAt))
}

func TestNewSession(t *testing.T) {
	userID := uuid.New()
	tokenHash := "tokenhash"
	expiresAt := time.Now().Add(24 * time.Hour)
	session := domain.NewSession(userID, tokenHash, "Mozilla/5.0", "192.168.1.1", expiresAt)

	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, tokenHash, session.TokenHash)
	assert.Equal(t, "Mozilla/5.0", session.UserAgent)
	assert.Equal(t, "192.168.1.1", session.IPAddress)
	assert.Equal(t, expiresAt, session.ExpiresAt)
	assert.Nil(t, session.ActiveProfileID)
}

func TestSessionIsExpired(t *testing.T) {
	userID := uuid.New()
	tokenHash := "tokenhash"
	expiredSession := domain.NewSession(userID, tokenHash, "", "", time.Now().Add(-1*time.Hour))
	assert.True(t, expiredSession.IsExpired())

	validSession := domain.NewSession(userID, tokenHash, "", "", time.Now().Add(1*time.Hour))
	assert.False(t, validSession.IsExpired())
}

func TestSessionSetActiveProfile(t *testing.T) {
	session := domain.NewSession(uuid.New(), "tokenhash", "", "", time.Now().Add(24*time.Hour))
	assert.Nil(t, session.ActiveProfileID)

	profileID := uuid.New()
	session.SetActiveProfile(profileID)
	assert.NotNil(t, session.ActiveProfileID)
	assert.Equal(t, profileID, *session.ActiveProfileID)
}

func TestSessionClearActiveProfile(t *testing.T) {
	session := domain.NewSession(uuid.New(), "tokenhash", "", "", time.Now().Add(24*time.Hour))
	session.SetActiveProfile(uuid.New())
	assert.NotNil(t, session.ActiveProfileID)

	session.ClearActiveProfile()
	assert.Nil(t, session.ActiveProfileID)
}

func TestNewProfile(t *testing.T) {
	userID := uuid.New()
	profileType := domain.ProfileTypeBrand
	name := "My Company"
	details := json.RawMessage(`{"company_name":"Acme Corp","size":"enterprise","industry":"tech"}`)

	profile := domain.NewProfile(userID, profileType, name, details)

	assert.NotEqual(t, uuid.Nil, profile.ID)
	assert.Equal(t, userID, profile.UserID)
	assert.Equal(t, profileType, profile.Type)
	assert.Equal(t, name, profile.Name)
	assert.NotNil(t, profile.Details)
	assert.Nil(t, profile.DeletedAt)
	assert.False(t, profile.CreatedAt.IsZero())
}

func TestProfileIsDeleted(t *testing.T) {
	profile := domain.NewProfile(uuid.New(), domain.ProfileTypeEditor, "Editor Profile", nil)
	assert.False(t, profile.IsDeleted())

	profile.SoftDelete()
	assert.True(t, profile.IsDeleted())
	assert.NotNil(t, profile.DeletedAt)
}

func TestProfileSoftDelete(t *testing.T) {
	profile := domain.NewProfile(uuid.New(), domain.ProfileTypeInfluencer, "Influencer Profile", nil)
	oldUpdatedAt := profile.UpdatedAt

	profile.SoftDelete()
	assert.NotNil(t, profile.DeletedAt)
	assert.True(t, profile.UpdatedAt.After(oldUpdatedAt))
}

func TestProfileUpdate(t *testing.T) {
	profile := domain.NewProfile(uuid.New(), domain.ProfileTypeBrand, "Old Name", json.RawMessage(`{}`))
	oldUpdatedAt := profile.UpdatedAt

	newName := "New Name"
	newDetails := json.RawMessage(`{"company_name":"New Corp"}`)
	profile.Update(newName, newDetails)

	assert.Equal(t, newName, profile.Name)
	assert.JSONEq(t, string(newDetails), string(profile.Details))
	assert.True(t, profile.UpdatedAt.After(oldUpdatedAt))
}

func TestNewAuthToken(t *testing.T) {
	userID := uuid.New()
	tokenType := domain.TokenTypeVerification
	tokenHash := "tokenhash"
	expiresAt := time.Now().Add(24 * time.Hour)

	token := domain.NewAuthToken(userID, tokenType, tokenHash, expiresAt)

	assert.NotEqual(t, uuid.Nil, token.ID)
	assert.Equal(t, userID, token.UserID)
	assert.Equal(t, tokenType, token.TokenType)
	assert.Equal(t, tokenHash, token.TokenHash)
	assert.Equal(t, expiresAt, token.ExpiresAt)
	assert.Nil(t, token.UsedAt)
}

func TestAuthTokenIsExpired(t *testing.T) {
	token := domain.NewAuthToken(uuid.New(), domain.TokenTypeVerification, "hash", time.Now().Add(-1*time.Hour))
	assert.True(t, token.IsExpired())

	token2 := domain.NewAuthToken(uuid.New(), domain.TokenTypePasswordReset, "hash", time.Now().Add(1*time.Hour))
	assert.False(t, token2.IsExpired())
}

func TestAuthTokenIsUsed(t *testing.T) {
	token := domain.NewAuthToken(uuid.New(), domain.TokenTypeVerification, "hash", time.Now().Add(1*time.Hour))
	assert.False(t, token.IsUsed())

	token.MarkUsed()
	assert.True(t, token.IsUsed())
	assert.NotNil(t, token.UsedAt)
}

func TestAuthTokenMarkUsed(t *testing.T) {
	token := domain.NewAuthToken(uuid.New(), domain.TokenTypePasswordReset, "hash", time.Now().Add(1*time.Hour))
	oldUpdatedAt := token.CreatedAt

	token.MarkUsed()
	assert.NotNil(t, token.UsedAt)
	assert.True(t, token.UsedAt.After(oldUpdatedAt))
}

func TestProfileTypeConstants(t *testing.T) {
	assert.Equal(t, domain.ProfileType("brand"), domain.ProfileTypeBrand)
	assert.Equal(t, domain.ProfileType("editor"), domain.ProfileTypeEditor)
	assert.Equal(t, domain.ProfileType("influencer"), domain.ProfileTypeInfluencer)
}

func TestTokenTypeConstants(t *testing.T) {
	assert.Equal(t, domain.TokenType("verification"), domain.TokenTypeVerification)
	assert.Equal(t, domain.TokenType("password_reset"), domain.TokenTypePasswordReset)
}
