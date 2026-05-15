package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	adapter "viralforge/backend/src/adapter"
)

// LoginHistoryService handles login history operations.
type LoginHistoryService struct {
	store *adapter.PostgresStore
}

// NewLoginHistoryService creates a new LoginHistoryService.
func NewLoginHistoryService(store *adapter.PostgresStore) *LoginHistoryService {
	return &LoginHistoryService{store: store}
}

// RecordLogin records a new login event.
func (s *LoginHistoryService) RecordLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent, deviceFingerprint string) (*domain.LoginHistory, error) {
	history := domain.NewLoginHistory(userID, ipAddress, userAgent, deviceFingerprint)

	if err := s.store.CreateLoginHistory(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

// RecordLoginWithGeo records a new login event with geolocation data.
func (s *LoginHistoryService) RecordLoginWithGeo(ctx context.Context, userID uuid.UUID, ipAddress, userAgent, deviceFingerprint, city, country string, lat, lng float64) (*domain.LoginHistory, error) {
	history := domain.NewLoginHistory(userID, ipAddress, userAgent, deviceFingerprint)
	history.SetGeolocation(city, country, lat, lng)

	if err := s.store.CreateLoginHistory(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

// GetLoginHistory retrieves login history for a user with pagination.
func (s *LoginHistoryService) GetLoginHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.LoginHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	histories, err := s.store.LoginHistoriesByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.store.CountLoginHistoriesByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GenerateDeviceFingerprint generates a device fingerprint from browser characteristics.
func GenerateDeviceFingerprint(userAgent, acceptLanguage, screenResolution, timezone string) string {
	// Simple fingerprint - in production use a more robust approach
	combined := userAgent + acceptLanguage + screenResolution + timezone
	return hashString(combined)
}

// hashString creates a simple hash for fingerprinting (not cryptographic).
func hashString(s string) string {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return fmt.Sprintf("%08x", h)
}