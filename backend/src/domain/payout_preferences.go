package domain

import (
	"time"

	"github.com/google/uuid"
)

// PayoutMethod represents the preferred payout delivery method.
type PayoutMethod string

const (
	PayoutMethodBankTransfer PayoutMethod = "bank_transfer"
	PayoutMethodPayPal       PayoutMethod = "paypal"
	PayoutMethodCrypto       PayoutMethod = "crypto"
)

// PayoutPreferences represents a user's payout configuration and sensitive details.
type PayoutPreferences struct {
	ProfileID        uuid.UUID    `gorm:"type:uuid;primary_key;not null" json:"profile_id"`
	PreferredMethod  PayoutMethod `gorm:"type:varchar(50);not null" json:"preferred_method"`
	BeneficiaryName  string       `gorm:"type:varchar(255);not null" json:"beneficiary_name"`
	Country          string       `gorm:"type:varchar(2);not null" json:"country"`
	Currency         string       `gorm:"type:varchar(3);not null" json:"currency"`
	EncryptedDetails string       `gorm:"type:text" json:"-"` // DB-layer encryption, never returned in plaintext
	PayoutReady      bool         `gorm:"type:bool;default:false" json:"payout_ready"`
	UpdatedAt        time.Time    `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for PayoutPreferences.
func (PayoutPreferences) TableName() string { return "payout_preferences" }

// Update updates payout preference fields.
func (p *PayoutPreferences) Update(method PayoutMethod, beneficiaryName, country, currency string, encryptedDetails string, payoutReady bool) error {
	p.PreferredMethod = method
	p.BeneficiaryName = beneficiaryName
	p.Country = country
	p.Currency = currency
	p.EncryptedDetails = encryptedDetails
	p.PayoutReady = payoutReady
	p.UpdatedAt = time.Now()
	return nil
}

// GetMasked returns a safe representation for API responses (never exposes encrypted_details).
func (p *PayoutPreferences) GetMasked() *PayoutPreferencesMasked {
	return &PayoutPreferencesMasked{
		ProfileID:       p.ProfileID,
		PreferredMethod: p.PreferredMethod,
		BeneficiaryName: p.BeneficiaryName,
		Country:         p.Country,
		Currency:        p.Currency,
		PayoutReady:     p.PayoutReady,
		UpdatedAt:       p.UpdatedAt,
	}
}

// PayoutPreferencesMasked is the safe API response type (encrypted_details excluded).
type PayoutPreferencesMasked struct {
	ProfileID       uuid.UUID    `json:"profile_id"`
	PreferredMethod PayoutMethod `json:"preferred_method"`
	BeneficiaryName string       `json:"beneficiary_name"`
	Country         string       `json:"country"`
	Currency        string       `json:"currency"`
	PayoutReady     bool         `json:"payout_ready"`
	UpdatedAt       time.Time    `json:"updated_at"`
}
