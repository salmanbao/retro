package domain

import (
	"time"

	"github.com/google/uuid"
)

// LoginHistory records a login event for security auditing.
type LoginHistory struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	IPAddress         string     `gorm:"type:varchar(45);not null" json:"ip_address"`
	UserAgent         string     `gorm:"type:varchar(512);not null" json:"user_agent"`
	DeviceFingerprint string     `gorm:"type:varchar(255);not null" json:"device_fingerprint"`
	City              *string    `gorm:"type:varchar(100)" json:"city,omitempty"`
	Country           *string    `gorm:"type:varchar(100)" json:"country,omitempty"`
	Latitude          *float64   `gorm:"type:decimal(9,6)" json:"latitude,omitempty"`
	Longitude         *float64   `gorm:"type:decimal(9,6)" json:"longitude,omitempty"`
	LoggedInAt        time.Time  `gorm:"type:timestamptz;not null" json:"logged_in_at"`
}

// TableName sets the table name for LoginHistory.
func (LoginHistory) TableName() string { return "login_history" }

// NewLoginHistory creates a new login history record.
func NewLoginHistory(userID uuid.UUID, ipAddress, userAgent, deviceFingerprint string) *LoginHistory {
	return &LoginHistory{
		ID:                uuid.New(),
		UserID:            userID,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		DeviceFingerprint: deviceFingerprint,
		LoggedInAt:        time.Now(),
	}
}

// SetGeolocation sets the geolocation data.
func (lh *LoginHistory) SetGeolocation(city, country string, lat, lng float64) {
	lh.City = &city
	lh.Country = &country
	lh.Latitude = &lat
	lh.Longitude = &lng
}