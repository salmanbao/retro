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

const maxPortfolioItems = 50

// ErrProfileNotEditor is returned when a non-Editor profile attempts portfolio operations.
var ErrProfileNotEditor = errors.New("profile must have Editor role for portfolio operations")

// ErrPortfolioLimitReached is returned when the maximum portfolio items limit is reached.
var ErrPortfolioLimitReached = errors.New("maximum portfolio items (50) reached for this profile")

// PortfolioService handles portfolio business logic.
type PortfolioService struct {
	portfolioItemRepo repository.PortfolioItemRepository
	profileRepo       repository.ProfileRepository
	logger            *slog.Logger
}

// NewPortfolioService creates a new PortfolioService.
func NewPortfolioService(portfolioItemRepo repository.PortfolioItemRepository, profileRepo repository.ProfileRepository) *PortfolioService {
	return &PortfolioService{
		portfolioItemRepo: portfolioItemRepo,
		profileRepo:       profileRepo,
		logger:            slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})),
	}
}

// NewPortfolioServiceWithLogger creates a new PortfolioService with a logger.
func NewPortfolioServiceWithLogger(portfolioItemRepo repository.PortfolioItemRepository, profileRepo repository.ProfileRepository, logger *slog.Logger) *PortfolioService {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
	}
	return &PortfolioService{
		portfolioItemRepo: portfolioItemRepo,
		profileRepo:       profileRepo,
		logger:            logger,
	}
}

// ListByProfileID retrieves all portfolio items for a profile.
func (s *PortfolioService) ListByProfileID(ctx context.Context, profileID uuid.UUID, limit, offset int) ([]*domain.PortfolioItem, error) {
	return s.portfolioItemRepo.ByProfileID(ctx, profileID, limit, offset)
}

// Create creates a new portfolio item for an Editor profile.
func (s *PortfolioService) Create(ctx context.Context, profileID uuid.UUID, title, description, thumbnailURL, videoURL, externalLink string, displayOrder int) (*domain.PortfolioItem, error) {
	s.logger.Info("creating portfolio item",
		slog.String("profile_id", profileID.String()),
		slog.String("title", title),
	)

	// Verify Editor role
	if !s.hasEditorRole(ctx, profileID) {
		s.logger.Warn("portfolio creation denied - not editor",
			slog.String("profile_id", profileID.String()),
		)
		return nil, ErrProfileNotEditor
	}

	// Check item limit
	count, err := s.portfolioItemRepo.CountByProfileID(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to count portfolio items",
			slog.String("profile_id", profileID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	if count >= maxPortfolioItems {
		s.logger.Warn("portfolio limit reached",
			slog.String("profile_id", profileID.String()),
			slog.Int64("current_count", count),
		)
		return nil, ErrPortfolioLimitReached
	}

	item := domain.NewPortfolioItem(profileID, title, description, thumbnailURL, videoURL, externalLink, displayOrder)
	if err := s.portfolioItemRepo.Create(ctx, item); err != nil {
		s.logger.Error("failed to create portfolio item",
			slog.String("profile_id", profileID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("portfolio item created",
		slog.String("profile_id", profileID.String()),
		slog.String("item_id", item.ID.String()),
	)
	return item, nil
}

// Update updates an existing portfolio item.
func (s *PortfolioService) Update(ctx context.Context, itemID uuid.UUID, profileID uuid.UUID, title, description, thumbnailURL, videoURL, externalLink string, displayOrder int) (*domain.PortfolioItem, error) {
	s.logger.Info("updating portfolio item",
		slog.String("item_id", itemID.String()),
		slog.String("profile_id", profileID.String()),
	)

	// Verify Editor role
	if !s.hasEditorRole(ctx, profileID) {
		s.logger.Warn("portfolio update denied - not editor",
			slog.String("profile_id", profileID.String()),
		)
		return nil, ErrProfileNotEditor
	}

	item, err := s.portfolioItemRepo.ByID(ctx, itemID)
	if err != nil {
		s.logger.Error("failed to find portfolio item",
			slog.String("item_id", itemID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Verify ownership
	if item.ProfileID != profileID {
		s.logger.Warn("portfolio update denied - not owner",
			slog.String("item_id", itemID.String()),
			slog.String("profile_id", profileID.String()),
			slog.String("owner_id", item.ProfileID.String()),
		)
		return nil, domain.ErrPortfolioItemNotFound
	}

	item.Update(title, description, thumbnailURL, videoURL, externalLink, displayOrder)
	if err := s.portfolioItemRepo.Update(ctx, item); err != nil {
		s.logger.Error("failed to update portfolio item",
			slog.String("item_id", itemID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("portfolio item updated",
		slog.String("item_id", itemID.String()),
	)
	return item, nil
}

// Delete soft-deletes a portfolio item.
func (s *PortfolioService) Delete(ctx context.Context, itemID uuid.UUID, profileID uuid.UUID) error {
	s.logger.Info("deleting portfolio item",
		slog.String("item_id", itemID.String()),
		slog.String("profile_id", profileID.String()),
	)

	// Verify Editor role
	if !s.hasEditorRole(ctx, profileID) {
		s.logger.Warn("portfolio delete denied - not editor",
			slog.String("profile_id", profileID.String()),
		)
		return ErrProfileNotEditor
	}

	item, err := s.portfolioItemRepo.ByID(ctx, itemID)
	if err != nil {
		s.logger.Error("failed to find portfolio item for delete",
			slog.String("item_id", itemID.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	// Verify ownership
	if item.ProfileID != profileID {
		s.logger.Warn("portfolio delete denied - not owner",
			slog.String("item_id", itemID.String()),
			slog.String("profile_id", profileID.String()),
		)
		return domain.ErrPortfolioItemNotFound
	}

	if err := s.portfolioItemRepo.Delete(ctx, itemID); err != nil {
		s.logger.Error("failed to delete portfolio item",
			slog.String("item_id", itemID.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("portfolio item deleted",
		slog.String("item_id", itemID.String()),
	)
	return nil
}

// CountByProfileID returns the number of portfolio items for a profile.
func (s *PortfolioService) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return s.portfolioItemRepo.CountByProfileID(ctx, profileID)
}

// hasEditorRole checks if the profile has the Editor role.
func (s *PortfolioService) hasEditorRole(ctx context.Context, profileID uuid.UUID) bool {
	profile, err := s.profileRepo.ByID(ctx, profileID)
	if err != nil {
		return false
	}
	return profile.Type == domain.ProfileTypeEditor
}
