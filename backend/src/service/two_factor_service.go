package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"viralforge/backend/src/domain"
	adapter "viralforge/backend/src/adapter"
)

const (
	// TOTP parameters
	totpDigits    = 6
	totpPeriod    = 30
	totpAlgorithm = otp.AlgorithmSHA1

	// Backup code parameters
	backupCodeCount  = 8
	backupCodeLength = 10
	bcryptCostValue  = 12
)

// TwoFactorService handles 2FA operations.
type TwoFactorService struct {
	store         *adapter.PostgresStore
	encryptionKey []byte // 32 bytes for AES-256
}

// NewTwoFactorService creates a new TwoFactorService.
func NewTwoFactorService(store *adapter.PostgresStore, encryptionKey []byte) *TwoFactorService {
	return &TwoFactorService{
		store:         store,
		encryptionKey: encryptionKey,
	}
}

// Setup2FA generates a new TOTP secret and backup codes for a user.
func (s *TwoFactorService) Setup2FA(ctx context.Context, userID uuid.UUID, issuer string) (secret string, qrCodeURL string, backupCodes []string, err error) {
	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: userID.String(),
		Period:      totpPeriod,
		Digits:      otp.DigitsSix,
		Algorithm:   totpAlgorithm,
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	secret = key.Secret()
	qrCodeURL = key.URL()

	// Encrypt the secret for storage
	encryptedSecret, err := s.encryptTOTPSecret(secret)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	// Generate backup codes
	backupCodes, backupCodesHash, err := s.generateBackupCodes()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Store 2FA settings (disabled until verified)
	settings := domain.NewTwoFactorSettings(userID, encryptedSecret, backupCodesHash)
	if err := s.store.CreateTwoFactorSettings(ctx, settings); err != nil {
		return "", "", nil, fmt.Errorf("failed to store 2FA settings: %w", err)
	}

	return secret, qrCodeURL, backupCodes, nil
}

// Verify2FASetup validates a TOTP code to confirm 2FA setup and enables it.
func (s *TwoFactorService) Verify2FASetup(ctx context.Context, userID uuid.UUID, code string) error {
	settings, err := s.store.TwoFactorSettingsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get 2FA settings: %w", err)
	}

	// Decrypt the secret
	secret, err := s.decryptTOTPSecret(settings.TOTPSecretEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	// Validate the TOTP code
	if !s.validateTOTP(secret, code) {
		return errors.New("invalid TOTP code")
	}

	// Enable 2FA
	settings.Enable()
	if err := s.store.UpdateTwoFactorSettings(ctx, settings); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// ValidateTOTP validates a TOTP code for login.
func (s *TwoFactorService) ValidateTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	settings, err := s.store.TwoFactorSettingsByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get 2FA settings: %w", err)
	}

	if !settings.Enabled {
		return false, errors.New("2FA is not enabled")
	}

	// Decrypt the secret
	secret, err := s.decryptTOTPSecret(settings.TOTPSecretEncrypted)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	// Validate with ±1 window tolerance
	return s.validateTOTP(secret, code), nil
}

// Disable2FAWithBackupCode disables 2FA using a backup code.
func (s *TwoFactorService) Disable2FAWithBackupCode(ctx context.Context, userID uuid.UUID, backupCode string) error {
	settings, err := s.store.TwoFactorSettingsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get 2FA settings: %w", err)
	}

	// Parse and validate backup codes
	var codes []string
	if err := json.Unmarshal([]byte(settings.BackupCodesHash), &codes); err != nil {
		return fmt.Errorf("failed to parse backup codes: %w", err)
	}

	// Check each backup code
	for i, storedHash := range codes {
		if storedHash == "" {
			continue // Already used
		}
		// Verify the backup code matches this hash
		if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(backupCode)); err == nil {
			// Code is valid - mark all codes as used (invalidate all backup codes)
			codes[i] = "" // Mark as used
			updatedHash, _ := json.Marshal(codes)
			settings.BackupCodesHash = string(updatedHash)
			settings.Disable()

			if err := s.store.UpdateTwoFactorSettings(ctx, settings); err != nil {
				return fmt.Errorf("failed to disable 2FA: %w", err)
			}
			return nil
		}
	}

	return errors.New("invalid backup code")
}

// encryptTOTPSecret encrypts a TOTP secret using AES-256-GCM.
func (s *TwoFactorService) encryptTOTPSecret(secret string) (string, error) {
	if len(s.encryptionKey) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptTOTPSecret decrypts an encrypted TOTP secret.
func (s *TwoFactorService) decryptTOTPSecret(encrypted string) (string, error) {
	if len(s.encryptionKey) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// validateTOTP validates a TOTP code with ±1 window tolerance.
func (s *TwoFactorService) validateTOTP(secret, code string) bool {
	// Validate with time tolerance
	valid, _ := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    uint(totpPeriod),
		Skew:      1, // Allow ±1 period (90 seconds total)
		Digits:    otp.DigitsSix,
		Algorithm: totpAlgorithm,
	})
	return valid
}

// generateBackupCodes generates 8 backup codes and their bcrypt hashes.
func (s *TwoFactorService) generateBackupCodes() ([]string, string, error) {
	codes := make([]string, backupCodeCount)
	hashes := make([]string, backupCodeCount)

	for i := 0; i < backupCodeCount; i++ {
		code := generateBackupCode()
		codes[i] = code

		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcryptCostValue)
		if err != nil {
			return nil, "", err
		}
		hashes[i] = string(hash)
	}

	hashJSON, err := json.Marshal(hashes)
	if err != nil {
		return nil, "", err
	}

	return codes, string(hashJSON), nil
}

// generateBackupCode generates a single backup code.
func generateBackupCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excluding confusing chars
	const length = backupCodeLength

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}

// Get2FASettings retrieves 2FA settings for a user.
func (s *TwoFactorService) Get2FASettings(ctx context.Context, userID uuid.UUID) (*domain.TwoFactorSettings, error) {
	return s.store.TwoFactorSettingsByUserID(ctx, userID)
}

// Is2FAEnabled checks if 2FA is enabled for a user.
func (s *TwoFactorService) Is2FAEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	settings, err := s.store.TwoFactorSettingsByUserID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return settings.Enabled, nil
}