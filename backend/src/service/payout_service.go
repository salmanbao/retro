package service

import (
	"context"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// PayoutService handles payout preferences business logic.
type PayoutService struct {
	payoutPrefsRepo repository.PayoutPreferencesRepository
	profileRepo     repository.ProfileRepository
	logger          *slog.Logger
}

// NewPayoutService creates a new PayoutService.
func NewPayoutService(payoutPrefsRepo repository.PayoutPreferencesRepository, profileRepo repository.ProfileRepository) *PayoutService {
	return &PayoutService{
		payoutPrefsRepo: payoutPrefsRepo,
		profileRepo:     profileRepo,
		logger:          slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

// NewPayoutServiceWithLogger creates a new PayoutService with a logger.
func NewPayoutServiceWithLogger(payoutPrefsRepo repository.PayoutPreferencesRepository, profileRepo repository.ProfileRepository, logger *slog.Logger) *PayoutService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return &PayoutService{
		payoutPrefsRepo: payoutPrefsRepo,
		profileRepo:     profileRepo,
		logger:          logger,
	}
}

// GetPayoutPreferences retrieves payout preferences for a profile.
// Returns masked data - encrypted_details is never included.
func (s *PayoutService) GetPayoutPreferences(ctx context.Context, profileID uuid.UUID) (*domain.PayoutPreferencesMasked, error) {
	prefs, err := s.payoutPrefsRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			// Return empty masked preferences if not found
			return &domain.PayoutPreferencesMasked{
				ProfileID: profileID,
			}, nil
		}
		return nil, err
	}
	return prefs.GetMasked(), nil
}

// UpdatePayoutPreferences updates payout preferences for a profile.
// The encryptedDetails field is stored but never returned in responses.
func (s *PayoutService) UpdatePayoutPreferences(ctx context.Context, profileID uuid.UUID, method domain.PayoutMethod, beneficiaryName, country, currency, encryptedDetails string, payoutReady bool) (*domain.PayoutPreferencesMasked, error) {
	s.logger.Info("updating payout preferences",
		slog.String("profile_id", profileID.String()),
		slog.String("method", string(method)),
	)

	// Validate country code (ISO 3166-1 alpha-2)
	if err := domain.ValidateCountryCode(country); err != nil {
		s.logger.Warn("invalid country code",
			slog.String("profile_id", profileID.String()),
			slog.String("country", country),
		)
		return nil, domain.ErrInvalidCountryCode
	}

	// Validate currency code (ISO 4217)
	if err := domain.ValidateCurrencyCode(currency); err != nil {
		s.logger.Warn("invalid currency code",
			slog.String("profile_id", profileID.String()),
			slog.String("currency", currency),
		)
		return nil, domain.ErrInvalidCurrencyCode
	}

	// Get existing or create new
	prefs, err := s.payoutPrefsRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			prefs = &domain.PayoutPreferences{
				ProfileID: profileID,
			}
		} else {
			s.logger.Error("failed to get payout preferences",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Update fields
	if err := prefs.Update(method, beneficiaryName, country, currency, encryptedDetails, payoutReady); err != nil {
		s.logger.Error("failed to update payout preferences",
			slog.String("profile_id", profileID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if prefs.ProfileID == uuid.Nil {
		if err := s.payoutPrefsRepo.Create(ctx, prefs); err != nil {
			s.logger.Error("failed to create payout preferences",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	} else {
		if err := s.payoutPrefsRepo.Update(ctx, prefs); err != nil {
			s.logger.Error("failed to update payout preferences",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	s.logger.Info("payout preferences updated",
		slog.String("profile_id", profileID.String()),
	)
	return prefs.GetMasked(), nil
}

// IsPayoutReady checks if a profile has complete payout information.
func (s *PayoutService) IsPayoutReady(ctx context.Context, profileID uuid.UUID) (bool, error) {
	prefs, err := s.payoutPrefsRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			return false, nil
		}
		return false, err
	}
	return prefs.PayoutReady, nil
}
