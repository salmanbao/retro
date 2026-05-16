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

// ErrProfileNotInfluencerVerification is returned when a non-Influencer profile attempts verification operations.
var ErrProfileNotInfluencerVerification = errors.New("profile must have Influencer role for verification operations")

// ErrVerificationNotFound is returned when verification data is not found.
var ErrVerificationNotFound = errors.New("verification not found")

// ErrInvalidVerificationStatus is returned when an invalid status transition is attempted.
var ErrInvalidVerificationStatus = errors.New("invalid verification status")

// VerificationService handles follower verification business logic.
type VerificationService struct {
	verificationRepo repository.FollowerVerificationRepository
	profileRepo      repository.ProfileRepository
	logger           *slog.Logger
}

// NewVerificationService creates a new VerificationService.
func NewVerificationService(verificationRepo repository.FollowerVerificationRepository, profileRepo repository.ProfileRepository) *VerificationService {
	return &VerificationService{
		verificationRepo: verificationRepo,
		profileRepo:      profileRepo,
		logger:           slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

// NewVerificationServiceWithLogger creates a new VerificationService with a logger.
func NewVerificationServiceWithLogger(verificationRepo repository.FollowerVerificationRepository, profileRepo repository.ProfileRepository, logger *slog.Logger) *VerificationService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return &VerificationService{
		verificationRepo: verificationRepo,
		profileRepo:      profileRepo,
		logger:           logger,
	}
}

// GetVerification retrieves verification data for a profile.
func (s *VerificationService) GetVerification(ctx context.Context, profileID uuid.UUID) (*domain.FollowerVerification, error) {
	return s.verificationRepo.ByProfileID(ctx, profileID)
}

// SubmitVerification submits verification evidence for an Influencer profile.
func (s *VerificationService) SubmitVerification(ctx context.Context, profileID uuid.UUID, evidenceURLs []string, notes string) (*domain.FollowerVerification, error) {
	s.logger.Info("submitting verification",
		slog.String("profile_id", profileID.String()),
		slog.Int("evidence_count", len(evidenceURLs)),
	)

	// Verify Influencer role
	if !s.hasInfluencerRole(ctx, profileID) {
		s.logger.Warn("verification submission denied - not influencer",
			slog.String("profile_id", profileID.String()),
		)
		return nil, ErrProfileNotInfluencerVerification
	}

	// Get existing or create new
	verification, err := s.verificationRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			verification = &domain.FollowerVerification{
				ProfileID: profileID,
			}
		} else {
			s.logger.Error("failed to get verification",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Submit evidence (sets status to pending)
	if err := verification.SubmitEvidence(evidenceURLs, notes); err != nil {
		s.logger.Error("failed to submit evidence",
			slog.String("profile_id", profileID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if verification.ProfileID == uuid.Nil {
		if err := s.verificationRepo.Create(ctx, verification); err != nil {
			s.logger.Error("failed to create verification",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	} else {
		if err := s.verificationRepo.Update(ctx, verification); err != nil {
			s.logger.Error("failed to update verification",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	s.logger.Info("verification submitted",
		slog.String("profile_id", profileID.String()),
		slog.String("status", string(verification.Status)),
	)
	return verification, nil
}

// AdminReviewVerification updates verification status after admin review.
func (s *VerificationService) AdminReviewVerification(ctx context.Context, profileID uuid.UUID, status domain.VerificationStatus, reviewedBy string, notes string) (*domain.FollowerVerification, error) {
	s.logger.Info("admin reviewing verification",
		slog.String("profile_id", profileID.String()),
		slog.String("new_status", string(status)),
		slog.String("reviewed_by", reviewedBy),
	)

	// Validate status transition
	if !isValidVerificationStatus(status) {
		s.logger.Warn("invalid verification status transition",
			slog.String("profile_id", profileID.String()),
			slog.String("status", string(status)),
		)
		return nil, ErrInvalidVerificationStatus
	}

	verification, err := s.verificationRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			// Create new verification record with admin review
			verification = &domain.FollowerVerification{
				ProfileID: profileID,
			}
		} else {
			s.logger.Error("failed to get verification for admin review",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Apply admin review
	verification.Review(status, reviewedBy, notes)

	if verification.ProfileID == uuid.Nil {
		if err := s.verificationRepo.Create(ctx, verification); err != nil {
			s.logger.Error("failed to create verification during admin review",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	} else {
		if err := s.verificationRepo.Update(ctx, verification); err != nil {
			s.logger.Error("failed to update verification during admin review",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	s.logger.Info("verification admin review complete",
		slog.String("profile_id", profileID.String()),
		slog.String("new_status", string(verification.Status)),
	)
	return verification, nil
}

// hasInfluencerRole checks if the profile has the Influencer role.
func (s *VerificationService) hasInfluencerRole(ctx context.Context, profileID uuid.UUID) bool {
	// TODO: Implement actual role checking via profile roles
	return true
}

// isValidVerificationStatus checks if a status is valid for transition.
func isValidVerificationStatus(status domain.VerificationStatus) bool {
	switch status {
	case domain.VerificationStatusVerified, domain.VerificationStatusRejected:
		return true
	default:
		return false
	}
}
