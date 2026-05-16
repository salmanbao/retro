package service

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// ErrKYCNotFound is returned when KYC status is not found.
var ErrKYCNotFound = errors.New("KYC status not found")

// ErrInvalidKYCStatus is returned when an invalid KYC status is provided.
var ErrInvalidKYCStatus = errors.New("invalid KYC status")

// KYCService handles KYC status business logic.
type KYCService struct {
	kycStatusRepo repository.KYCStatusRepository
	profileRepo   repository.ProfileRepository
	logger        *slog.Logger
}

// NewKYCService creates a new KYCService.
func NewKYCService(kycStatusRepo repository.KYCStatusRepository, profileRepo repository.ProfileRepository) *KYCService {
	return &KYCService{
		kycStatusRepo: kycStatusRepo,
		profileRepo:   profileRepo,
		logger:        slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

// NewKYCServiceWithLogger creates a new KYCService with a logger.
func NewKYCServiceWithLogger(kycStatusRepo repository.KYCStatusRepository, profileRepo repository.ProfileRepository, logger *slog.Logger) *KYCService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return &KYCService{
		kycStatusRepo: kycStatusRepo,
		profileRepo:   profileRepo,
		logger:        logger,
	}
}

// GetKYCStatus retrieves KYC status for a profile.
func (s *KYCService) GetKYCStatus(ctx context.Context, profileID uuid.UUID) (*domain.KYCStatus, error) {
	return s.kycStatusRepo.ByProfileID(ctx, profileID)
}

// AdminUpdateKYC updates KYC status (admin only operation).
func (s *KYCService) AdminUpdateKYC(ctx context.Context, profileID uuid.UUID, status domain.KYCStatusValue, reviewedBy string, notes string) (*domain.KYCStatus, error) {
	s.logger.Info("admin updating KYC status",
		slog.String("profile_id", profileID.String()),
		slog.String("new_status", string(status)),
		slog.String("reviewed_by", reviewedBy),
	)

	// Validate status
	if !isValidKYCStatus(status) {
		s.logger.Warn("invalid KYC status",
			slog.String("profile_id", profileID.String()),
			slog.String("status", string(status)),
		)
		return nil, ErrInvalidKYCStatus
	}

	kycStatus, err := s.kycStatusRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			// Create new KYC status record
			kycStatus = &domain.KYCStatus{
				ProfileID: profileID,
			}
		} else {
			s.logger.Error("failed to get KYC status",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Apply admin update
	kycStatus.UpdateStatus(status, reviewedBy, notes)

	if kycStatus.ProfileID == uuid.Nil {
		if err := s.kycStatusRepo.Create(ctx, kycStatus); err != nil {
			s.logger.Error("failed to create KYC status",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	} else {
		if err := s.kycStatusRepo.Update(ctx, kycStatus); err != nil {
			s.logger.Error("failed to update KYC status",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	s.logger.Info("KYC status updated",
		slog.String("profile_id", profileID.String()),
		slog.String("new_status", string(kycStatus.Status)),
	)
	return kycStatus, nil
}

// isValidKYCStatus checks if a KYC status value is valid.
func isValidKYCStatus(status domain.KYCStatusValue) bool {
	switch status {
	case domain.KYCStatusNotStarted, domain.KYCStatusPending, domain.KYCStatusApproved, domain.KYCStatusRejected, domain.KYCStatusSuspended:
		return true
	default:
		return false
	}
}
