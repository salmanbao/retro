package service

import (
	"context"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// ProfileEnrichmentService handles profile enrichment business logic.
type ProfileEnrichmentService struct {
	enrichmentRepo repository.ProfileEnrichmentRepository
	profileRepo    repository.ProfileRepository
	logger         *slog.Logger
}

// NewProfileEnrichmentService creates a new ProfileEnrichmentService.
func NewProfileEnrichmentService(enrichmentRepo repository.ProfileEnrichmentRepository, profileRepo repository.ProfileRepository) *ProfileEnrichmentService {
	return &ProfileEnrichmentService{
		enrichmentRepo: enrichmentRepo,
		profileRepo:    profileRepo,
		logger:         slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

// NewProfileEnrichmentServiceWithLogger creates a new ProfileEnrichmentService with a logger.
func NewProfileEnrichmentServiceWithLogger(enrichmentRepo repository.ProfileEnrichmentRepository, profileRepo repository.ProfileRepository, logger *slog.Logger) *ProfileEnrichmentService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return &ProfileEnrichmentService{
		enrichmentRepo: enrichmentRepo,
		profileRepo:    profileRepo,
		logger:         logger,
	}
}

// GetDetails retrieves profile enrichment by profile ID.
func (s *ProfileEnrichmentService) GetDetails(ctx context.Context, profileID uuid.UUID) (*domain.ProfileEnrichment, error) {
	return s.enrichmentRepo.ByProfileID(ctx, profileID)
}

// UpdateDetails updates profile enrichment with partial data.
func (s *ProfileEnrichmentService) UpdateDetails(ctx context.Context, profileID uuid.UUID, bio, avatarURL, coverURL, websiteURL, location string, languages []string, timezone string, socialLinks *domain.SocialLinks) (*domain.ProfileEnrichment, error) {
	s.logger.Info("updating profile details",
		slog.String("profile_id", profileID.String()),
	)

	enrichment, err := s.enrichmentRepo.ByProfileID(ctx, profileID)
	if err != nil {
		if err == domain.ErrProfileNotFound {
			// Create new enrichment if not found
			enrichment = domain.NewProfileEnrichment(profileID)
		} else {
			s.logger.Error("failed to get profile enrichment",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Validate language codes
	if len(languages) > 0 {
		if err := domain.ValidateLanguageCodes(languages); err != nil {
			s.logger.Warn("invalid language codes",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Validate timezone
	if timezone != "" {
		if err := domain.ValidateTimezone(timezone); err != nil {
			s.logger.Warn("invalid timezone",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	// Update fields
	if err := enrichment.Update(bio, avatarURL, coverURL, websiteURL, location, languages, timezone, socialLinks); err != nil {
		s.logger.Error("failed to update profile enrichment",
			slog.String("profile_id", profileID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if enrichment.ID.String() == "00000000-0000-0000-0000-000000000000" {
		if err := s.enrichmentRepo.Create(ctx, enrichment); err != nil {
			s.logger.Error("failed to create profile enrichment",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	} else {
		if err := s.enrichmentRepo.Update(ctx, enrichment); err != nil {
			s.logger.Error("failed to update profile enrichment",
				slog.String("profile_id", profileID.String()),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
	}

	s.logger.Info("profile details updated",
		slog.String("profile_id", profileID.String()),
	)
	return enrichment, nil
}

// CreateIfNotExists creates profile enrichment if it doesn't exist.
func (s *ProfileEnrichmentService) CreateIfNotExists(ctx context.Context, profileID uuid.UUID) (*domain.ProfileEnrichment, error) {
	existing, err := s.enrichmentRepo.ByProfileID(ctx, profileID)
	if err == nil {
		return existing, nil
	}
	if err != domain.ErrProfileNotFound {
		return nil, err
	}

	// Create new
	enrichment := domain.NewProfileEnrichment(profileID)
	if err := s.enrichmentRepo.Create(ctx, enrichment); err != nil {
		return nil, err
	}
	return enrichment, nil
}
