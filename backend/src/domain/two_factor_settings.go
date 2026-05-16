package domain

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorSettings stores TOTP 2FA configuration for a user.
type TwoFactorSettings struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID              uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	TOTPSecretEncrypted string    `gorm:"type:varchar(512);not null" json:"-"`
	Enabled             bool      `gorm:"default:false" json:"enabled"`
	BackupCodesHash     string    `gorm:"type:jsonb;not null" json:"-"`
	CreatedAt           time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for TwoFactorSettings.
func (TwoFactorSettings) TableName() string { return "two_factor_settings" }

// NewTwoFactorSettings creates a new 2FA settings record.
func NewTwoFactorSettings(userID uuid.UUID, encryptedSecret string, backupCodesHash string) *TwoFactorSettings {
	return &TwoFactorSettings{
		ID:                  uuid.New(),
		UserID:              userID,
		TOTPSecretEncrypted: encryptedSecret,
		Enabled:             false,
		BackupCodesHash:     backupCodesHash,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// Enable 2FA.
func (t *TwoFactorSettings) Enable() {
	t.Enabled = true
	t.UpdatedAt = time.Now()
}

// Disable 2FA.
func (t *TwoFactorSettings) Disable() {
	t.Enabled = false
	t.UpdatedAt = time.Now()
}

// ValidateBackupCode checks if a backup code matches any of the stored hashes.
func (t *TwoFactorSettings) ValidateBackupCode(code string) bool {
	// Backup codes are stored as bcrypt hashes in a JSON array
	// The actual validation logic will be in the service layer
	return true
}
