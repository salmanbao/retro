package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// ErrProfileNotInfluencer is returned when a non-Influencer profile attempts audience operations.
var ErrProfileNotInfluencer = errors.New("profile must have Influencer role for audience operations")

// ErrDemographicsTooLarge is returned when audience demographics exceed 10KB.
var ErrDemographicsTooLarge = errors.New("audience demographics exceed maximum size of 10KB")

// AudienceService handles audience data business logic.
type AudienceService struct {
	audienceDataRepo repository.AudienceDataRepository
	profileRepo      repository.ProfileRepository
	logger           *slog.Logger
}

// NewAudienceService creates a new AudienceService.
func NewAudienceService(audienceDataRepo repository.AudienceDataRepository, profileRepo repository.ProfileRepository) *AudienceService {
	return &AudienceService{
		audienceDataRepo: audienceDataRepo,
		profileRepo:      profileRepo,
		logger:           slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

// NewAudienceServiceWithLogger creates a new AudienceService with a logger.
func NewAudienceServiceWithLogger(audienceDataRepo repository.AudienceDataRepository, profileRepo repository.ProfileRepository, logger *slog.Logger) *AudienceService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return &AudienceService{
		audienceDataRepo: audienceDataRepo,
		profileRepo:      profileRepo,
		logger:           logger,
	}
}

// GetAudience retrieves audience data for a profile.
func (s *AudienceService) GetAudience(ctx context.Context, profileID uuid.UUID) (*domain.AudienceData, error) {
	return s.audienceDataRepo.ByProfileID(ctx, profileID)
}

// UpdateAudience updates audience data for an Influencer profile.
func (s *AudienceService) UpdateAudience(ctx context.Context, profileID uuid.UUID, platformHandles map[string]string, claimedFollowers map[string]int, engagementRate float64, demographics json.RawMessage) (*domain.AudienceData, error) {
	s.logger.Info("updating audience data",
		slog.String("profile_id", profileID.String()),
		slog.Float64("engagement_rate", engagementRate),
	)

	// Verify Influencer role
	if !s.hasInfluencerRole(ctx, profileID) {
		s.logger.Warn("audience update denied - not influencer",
			slog.String("profile_id", profileID.String()),
		)
		return nil, ErrProfileNotInfluencer
	}

	// Validate engagement rate range
	if engagementRate < 0 || engagementRate > 100 {
		engagementRate = 0
	}

	// Get existing or create new
	data, err := s.audienceDataRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			data = &domain.AudienceData{
				ProfileID: profileID,
			}
		} else {
			s.logger.Error("failed to get audience data",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Update fields
	if err := data.Update(platformHandles, claimedFollowers, engagementRate, demographics); err != nil {
		if err == domain.ErrDemographicsTooLarge {
			s.logger.Warn("audience demographics too large",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, ErrDemographicsTooLarge
		}
		s.logger.Error("failed to update audience data",
			slog.String("profile_id", profileID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if data.ProfileID == uuid.Nil {
		if err := s.audienceDataRepo.Create(ctx, data); err != nil {
			s.logger.Error("failed to create audience data",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	} else {
		if err := s.audienceDataRepo.Update(ctx, data); err != nil {
			s.logger.Error("failed to update audience data",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	s.logger.Info("audience data updated",
		slog.String("profile_id", profileID.String()),
	)
	return data, nil
}

// hasInfluencerRole checks if the profile has the Influencer role.
func (s *AudienceService) hasInfluencerRole(ctx context.Context, profileID uuid.UUID) bool {
	profile, err := s.profileRepo.ByID(ctx, profileID)
	if err != nil {
		return false
	}
	return profile.Type == domain.ProfileTypeInfluencer
}
