package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
)

// TestTOTPSecretGeneration tests TOTP secret generation
func TestTOTPSecretGeneration(t *testing.T) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ViralForge",
		AccountName: "test@example.com",
		Period:      30,
		Digits:      6,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, key.Secret())
	assert.NotEmpty(t, key.URL())
}

// TestTOTPCodeValidation tests TOTP code validation
func TestTOTPCodeValidation(t *testing.T) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ViralForge",
		AccountName: "test@example.com",
		Period:      30,
		Digits:      6,
	})
	require.NoError(t, err)

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	require.NoError(t, err)
	assert.Len(t, code, 6)

	valid, err := totp.ValidateCustom(code, key.Secret(), time.Now(), totp.ValidateOpts{
		Period: 30,
		Skew:   1,
		Digits: 6,
	})
	assert.NoError(t, err)
	assert.True(t, valid)
}

// TestBackupCodeGeneration tests backup code generation
func TestBackupCodeGeneration(t *testing.T) {
	codes := generateTestBackupCodes(8)
	assert.Len(t, codes, 8)

	for _, code := range codes {
		assert.Len(t, code, 10)
	}

	seen := make(map[string]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "Duplicate backup code found")
		seen[code] = true
	}
}

// TestTwoFactorSettingsEntity tests TwoFactorSettings entity
func TestTwoFactorSettingsEntity(t *testing.T) {
	settings := domain.NewTwoFactorSettings(
		testUserID,
		"encrypted-secret",
		`["hash1","hash2"]`,
	)

	assert.Equal(t, testUserID, settings.UserID)
	assert.Equal(t, "encrypted-secret", settings.TOTPSecretEncrypted)
	assert.False(t, settings.Enabled)

	settings.Enable()
	assert.True(t, settings.Enabled)

	settings.Disable()
	assert.False(t, settings.Enabled)
}

// TestLoginHistoryEntity tests LoginHistory entity
func TestLoginHistoryEntity(t *testing.T) {
	history := domain.NewLoginHistory(
		testUserID,
		"192.168.1.1",
		"Mozilla/5.0",
		"fingerprint123",
	)

	assert.Equal(t, testUserID, history.UserID)
	assert.Equal(t, "192.168.1.1", history.IPAddress)
	assert.Equal(t, "Mozilla/5.0", history.UserAgent)
	assert.Equal(t, "fingerprint123", history.DeviceFingerprint)

	history.SetGeolocation("New York", "US", 40.7128, -74.0060)
	assert.Equal(t, "New York", *history.City)
	assert.Equal(t, "US", *history.Country)
	assert.NotNil(t, history.Latitude)
	assert.NotNil(t, history.Longitude)
}

// generateTestBackupCodes generates backup codes for testing
func generateTestBackupCodes(count int) []string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const length = 10

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		result := make([]byte, length)
		for j := range result {
			result[j] = charset[(i+j)%len(charset)]
		}
		codes[i] = string(result)
	}
	return codes
}

// TestGenerateDeviceFingerprint tests device fingerprint generation
func TestGenerateDeviceFingerprint(t *testing.T) {
	fp1 := generateTestFingerprint("Mozilla/5.0", "en-US", "1920x1080", "America/New_York")
	fp2 := generateTestFingerprint("Mozilla/5.0", "en-US", "1920x1080", "America/New_York")
	fp3 := generateTestFingerprint("Mozilla/4.0", "en-US", "1920x1080", "America/New_York")

	assert.NotEmpty(t, fp1)
	assert.Equal(t, fp1, fp2, "Same inputs should produce same fingerprint")
	assert.NotEqual(t, fp1, fp3, "Different inputs should produce different fingerprint")
}

// generateTestFingerprint creates a test fingerprint
func generateTestFingerprint(ua, lang, screen, tz string) string {
	h := uint32(0)
	combined := ua + lang + screen + tz
	for i := range combined {
		h = h*31 + uint32(combined[i])
	}
	return string([]byte{
		byte((h >> 24) & 0xFF),
		byte((h >> 16) & 0xFF),
		byte((h >> 8) & 0xFF),
		byte(h & 0xFF),
	})
}

var testUserID = uuid.New()
