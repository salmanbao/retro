package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// ProfileService handles profile operations.
type ProfileService struct {
	profileRepo repository.ProfileRepository
	userRepo    repository.UserRepository
}

// NewProfileService creates a new ProfileService.
func NewProfileService(profileRepo repository.ProfileRepository, userRepo repository.UserRepository) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
		userRepo:    userRepo,
	}
}

// CreateProfileRequest represents a profile creation request.
type CreateProfileRequest struct {
	Type    domain.ProfileType     `json:"profile_type"`
	Name    string                 `json:"name"`
	Details map[string]interface{} `json:"details"`
}

// ValidateProfileType validates the profile type and required details.
func ValidateProfileType(profileType domain.ProfileType, details map[string]interface{}) error {
	switch profileType {
	case domain.ProfileTypeBrand:
		if _, ok := details["company_name"]; !ok || details["company_name"] == "" {
			return fmt.Errorf("company_name is required for Brand profile")
		}
		if _, ok := details["size"]; !ok || details["size"] == "" {
			return fmt.Errorf("size is required for Brand profile")
		}
		if _, ok := details["industry"]; !ok || details["industry"] == "" {
			return fmt.Errorf("industry is required for Brand profile")
		}
	case domain.ProfileTypeEditor:
		if _, ok := details["specializations"]; !ok {
			return fmt.Errorf("specializations is required for Editor profile")
		}
		specs, ok := details["specializations"].([]interface{})
		if !ok || len(specs) == 0 {
			return fmt.Errorf("specializations must be a non-empty array for Editor profile")
		}
		if _, ok := details["portfolio_url"]; !ok || details["portfolio_url"] == "" {
			return fmt.Errorf("portfolio_url is required for Editor profile")
		}
	case domain.ProfileTypeInfluencer:
		if _, ok := details["platforms"]; !ok {
			return fmt.Errorf("platforms is required for Influencer profile")
		}
		plats, ok := details["platforms"].([]interface{})
		if !ok || len(plats) == 0 {
			return fmt.Errorf("platforms must be a non-empty array for Influencer profile")
		}
		if _, ok := details["follower_counts"]; !ok {
			return fmt.Errorf("follower_counts is required for Influencer profile")
		}
	default:
		return domain.ErrInvalidProfileType
	}
	return nil
}

// CreateProfile creates a new profile for a user.
func (s *ProfileService) CreateProfile(ctx context.Context, userID uuid.UUID, req *CreateProfileRequest) (*domain.Profile, error) {
	// Validate profile type
	if req.Type != domain.ProfileTypeBrand && req.Type != domain.ProfileTypeEditor && req.Type != domain.ProfileTypeInfluencer {
		return nil, domain.ErrInvalidProfileType
	}

	// Validate type-specific details
	if err := ValidateProfileType(req.Type, req.Details); err != nil {
		return nil, err
	}

	// Check for duplicate active profile of same type
	existing, err := s.profileRepo.ByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, p := range existing {
		if p.Type == req.Type && !p.IsDeleted() {
			return nil, fmt.Errorf("active %s profile already exists", req.Type)
		}
	}

	// Marshal details to JSON
	detailsJSON, err := json.Marshal(req.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile details: %w", err)
	}

	// Create profile
	profile := domain.NewProfile(userID, req.Type, req.Name, detailsJSON)
	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// ListProfiles returns all active profiles for a user.
func (s *ProfileService) ListProfiles(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	profiles, err := s.profileRepo.ByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter out deleted profiles
	var active []*domain.Profile
	for _, p := range profiles {
		if !p.IsDeleted() {
			active = append(active, p)
		}
	}

	return active, nil
}

// GetProfile returns a profile by ID for a specific user.
func (s *ProfileService) GetProfile(ctx context.Context, userID, profileID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.profileRepo.ByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// Ensure profile belongs to user
	if profile.UserID != userID {
		return nil, domain.ErrProfileNotOwned
	}

	if profile.IsDeleted() {
		return nil, domain.ErrProfileNotFound
	}

	return profile, nil
}
