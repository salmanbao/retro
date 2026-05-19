package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// CampaignServiceInterface defines the interface for campaign service operations.
type CampaignServiceInterface interface {
	Create(ctx context.Context, input CreateCampaignInput) (*domain.Campaign, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Campaign, error)
	ListByBrandProfile(ctx context.Context, brandProfileID uuid.UUID, status string, page, pageSize int) ([]*domain.Campaign, int64, error)
	Update(ctx context.Context, input UpdateInput) (*domain.Campaign, error)
	Cancel(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error)
	Publish(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error)
	Pause(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error)
	Resume(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error)
	Complete(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error)
	TransitionPublishedToActive(ctx context.Context) (int, error)
}

// CampaignService handles campaign business logic.
type CampaignService struct {
	campaignRepo repository.CampaignRepository
	assetRepo    repository.CampaignAssetRepository
}

// NewCampaignService creates a new CampaignService.
func NewCampaignService(campaignRepo repository.CampaignRepository, assetRepo repository.CampaignAssetRepository) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		assetRepo:    assetRepo,
	}
}

// CreateCampaignInput contains the input for creating a campaign.
type CreateCampaignInput struct {
	BrandProfileID     uuid.UUID
	Title              string
	Summary            string
	Description        string
	Objective          string
	ProductName        string
	LandingURL         string
	TotalBudget        float64
	Currency           string
	TargetClips        int
	TargetPosts        int
	CPV                float64
	MinPayout          *float64
	MaxPayout          *float64
	SubmissionStart    time.Time
	SubmissionDeadline time.Time
	DistributionStart  time.Time
	CampaignEnd        time.Time
	AllowedCountries   []string
	AllowedLanguages   []string
	MinFollowers       int
	Platforms          []string
	CreatorCategories  []string
	MinDurationSecs    int
	MaxDurationSecs    int
	AspectRatio        string
	TalkingPoints      []string
	ProhibitedClaims   []string
	Hashtags           []string
	CTAInstructions    string
}

// GenerateSlug creates a URL-safe slug from a title.
func GenerateSlug(title string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(title) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			builder.WriteRune('-')
		default:
			// Skip special characters
		}
	}
	slug := builder.String()
	// Remove consecutive hyphens
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}

// ValidateTimeline validates the campaign timeline.
func ValidateTimeline(submissionStart, submissionDeadline, distributionStart, campaignEnd time.Time) error {
	if !submissionDeadline.After(submissionStart) {
		return domain.ErrInvalidTimeline
	}
	if !distributionStart.After(submissionDeadline) {
		return domain.ErrInvalidTimeline
	}
	if !campaignEnd.After(distributionStart) {
		return domain.ErrInvalidTimeline
	}
	return nil
}

// ValidatePayoutRange validates that min payout <= max payout.
func ValidatePayoutRange(min, max *float64) error {
	if min != nil && max != nil && *min > *max {
		return domain.ErrInvalidPayoutRange
	}
	return nil
}

// ValidateDuration validates that min duration <= max duration.
func ValidateDuration(minSecs, maxSecs int) error {
	if maxSecs < minSecs {
		return domain.ErrInvalidTimeline
	}
	return nil
}

// Create creates a new campaign.
func (s *CampaignService) Create(ctx context.Context, input CreateCampaignInput) (*domain.Campaign, error) {
	// Generate slug
	slug := GenerateSlug(input.Title)

	// Check for slug uniqueness
	existing, err := s.campaignRepo.BySlug(ctx, slug)
	if err != nil && !errors.Is(err, domain.ErrCampaignNotFound) {
		return nil, err
	}
	if existing != nil {
		// Append timestamp to make unique
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	// Validate timeline
	if err := ValidateTimeline(input.SubmissionStart, input.SubmissionDeadline, input.DistributionStart, input.CampaignEnd); err != nil {
		return nil, err
	}

	// Validate payout range
	if err := ValidatePayoutRange(input.MinPayout, input.MaxPayout); err != nil {
		return nil, err
	}

	// Validate duration
	if err := ValidateDuration(input.MinDurationSecs, input.MaxDurationSecs); err != nil {
		return nil, err
	}

	campaign := &domain.Campaign{
		ID:                 uuid.New(),
		BrandProfileID:     input.BrandProfileID,
		Title:              input.Title,
		Slug:               slug,
		Summary:            input.Summary,
		Description:        input.Description,
		Objective:          input.Objective,
		ProductName:        input.ProductName,
		LandingURL:         input.LandingURL,
		TotalBudget:        input.TotalBudget,
		Currency:           input.Currency,
		TargetClips:        input.TargetClips,
		TargetPosts:        input.TargetPosts,
		CPV:                input.CPV,
		MinPayout:          input.MinPayout,
		MaxPayout:          input.MaxPayout,
		SubmissionStart:    input.SubmissionStart,
		SubmissionDeadline: input.SubmissionDeadline,
		DistributionStart:  input.DistributionStart,
		CampaignEnd:        input.CampaignEnd,
		AllowedCountries:   input.AllowedCountries,
		AllowedLanguages:   input.AllowedLanguages,
		MinFollowers:       input.MinFollowers,
		Platforms:          input.Platforms,
		CreatorCategories:  input.CreatorCategories,
		MinDurationSecs:    input.MinDurationSecs,
		MaxDurationSecs:    input.MaxDurationSecs,
		AspectRatio:        input.AspectRatio,
		TalkingPoints:      input.TalkingPoints,
		ProhibitedClaims:   input.ProhibitedClaims,
		Hashtags:           input.Hashtags,
		CTAInstructions:    input.CTAInstructions,
		Status:             domain.CampaignStatusDraft,
		Version:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.campaignRepo.Create(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// GetByID retrieves a campaign by ID.
func (s *CampaignService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Campaign, error) {
	return s.campaignRepo.ByID(ctx, id)
}

// GetBySlug retrieves a campaign by slug.
func (s *CampaignService) GetBySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	return s.campaignRepo.BySlug(ctx, slug)
}

// ListByBrandProfile lists all campaigns for a brand profile with pagination.
func (s *CampaignService) ListByBrandProfile(ctx context.Context, brandProfileID uuid.UUID, status string, page, pageSize int) ([]*domain.Campaign, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.campaignRepo.List(ctx, brandProfileID, status, page, pageSize)
}

// UpdateInput contains the input for updating a campaign.
type UpdateInput struct {
	CampaignID         uuid.UUID
	BrandProfileID     uuid.UUID
	Title              *string
	Summary            *string
	Description        *string
	Objective          *string
	ProductName        *string
	LandingURL         *string
	TotalBudget        *float64
	Currency           *string
	TargetClips        *int
	TargetPosts        *int
	CPV                *float64
	MinPayout          *float64
	MaxPayout          *float64
	SubmissionStart    *time.Time
	SubmissionDeadline *time.Time
	DistributionStart  *time.Time
	CampaignEnd        *time.Time
	AllowedCountries   *[]string
	AllowedLanguages   *[]string
	MinFollowers       *int
	Platforms          *[]string
	CreatorCategories  *[]string
	MinDurationSecs    *int
	MaxDurationSecs    *int
	AspectRatio        *string
	TalkingPoints      *[]string
	ProhibitedClaims   *[]string
	Hashtags           *[]string
	CTAInstructions    *string
}

// GetEditableFields returns which fields can be edited based on campaign status.
func GetEditableFields(status domain.CampaignStatus) (allowed []string, rejected []string) {
	switch status {
	case domain.CampaignStatusDraft:
		return []string{"all"}, nil
	case domain.CampaignStatusPublished, domain.CampaignStatusActive:
		return []string{"summary", "description", "talking_points", "hashtags", "cta_instructions"},
			[]string{"title", "total_budget", "currency", "target_clips", "target_posts", "cpv", "min_payout", "max_payout", "submission_start", "submission_deadline", "distribution_start", "campaign_end", "allowed_countries", "allowed_languages", "min_followers", "platforms", "creator_categories", "min_duration_secs", "max_duration_secs", "aspect_ratio", "prohibited_claims"}
	case domain.CampaignStatusPaused:
		return []string{"summary", "description", "talking_points", "hashtags", "cta_instructions", "total_budget", "min_payout", "max_payout", "allowed_countries", "allowed_languages", "min_followers", "platforms", "creator_categories", "min_duration_secs", "max_duration_secs", "aspect_ratio", "prohibited_claims", "cpv", "target_clips", "target_posts"},
			[]string{"title", "submission_start", "submission_deadline", "distribution_start", "campaign_end"}
	case domain.CampaignStatusCompleted, domain.CampaignStatusCancelled:
		return nil, []string{"all"}
	default:
		return nil, []string{"all"}
	}
}

// Update updates a campaign.
func (s *CampaignService) Update(ctx context.Context, input UpdateInput) (*domain.Campaign, error) {
	campaign, err := s.campaignRepo.ByID(ctx, input.CampaignID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if campaign.BrandProfileID != input.BrandProfileID {
		return nil, domain.ErrCampaignNotOwned
	}

	// Get editable fields for current status
	allowed, _ := GetEditableFields(campaign.Status)
	if allowed == nil {
		return nil, domain.ErrRestrictedEdit
	}

	// Apply field updates (simplified - in production, check each field against allowed list)
	if input.Title != nil && contains(allowed, "all") || contains(allowed, "title") {
		campaign.Title = *input.Title
	}
	if input.Summary != nil && (contains(allowed, "all") || contains(allowed, "summary")) {
		campaign.Summary = *input.Summary
	}
	if input.Description != nil && (contains(allowed, "all") || contains(allowed, "description")) {
		campaign.Description = *input.Description
	}
	if input.Objective != nil && (contains(allowed, "all") || contains(allowed, "objective")) {
		campaign.Objective = *input.Objective
	}
	if input.ProductName != nil && (contains(allowed, "all") || contains(allowed, "product_name")) {
		campaign.ProductName = *input.ProductName
	}
	if input.LandingURL != nil && (contains(allowed, "all") || contains(allowed, "landing_url")) {
		campaign.LandingURL = *input.LandingURL
	}
	if input.TotalBudget != nil && (contains(allowed, "all") || contains(allowed, "total_budget")) {
		campaign.TotalBudget = *input.TotalBudget
	}
	if input.Currency != nil && (contains(allowed, "all") || contains(allowed, "currency")) {
		campaign.Currency = *input.Currency
	}
	if input.TargetClips != nil && (contains(allowed, "all") || contains(allowed, "target_clips")) {
		campaign.TargetClips = *input.TargetClips
	}
	if input.TargetPosts != nil && (contains(allowed, "all") || contains(allowed, "target_posts")) {
		campaign.TargetPosts = *input.TargetPosts
	}
	if input.CPV != nil && (contains(allowed, "all") || contains(allowed, "cpv")) {
		campaign.CPV = *input.CPV
	}
	if input.MinPayout != nil && (contains(allowed, "all") || contains(allowed, "min_payout")) {
		campaign.MinPayout = input.MinPayout
	}
	if input.MaxPayout != nil && (contains(allowed, "all") || contains(allowed, "max_payout")) {
		campaign.MaxPayout = input.MaxPayout
	}
	if input.SubmissionStart != nil && (contains(allowed, "all") || contains(allowed, "submission_start")) {
		campaign.SubmissionStart = *input.SubmissionStart
	}
	if input.SubmissionDeadline != nil && (contains(allowed, "all") || contains(allowed, "submission_deadline")) {
		campaign.SubmissionDeadline = *input.SubmissionDeadline
	}
	if input.DistributionStart != nil && (contains(allowed, "all") || contains(allowed, "distribution_start")) {
		campaign.DistributionStart = *input.DistributionStart
	}
	if input.CampaignEnd != nil && (contains(allowed, "all") || contains(allowed, "campaign_end")) {
		campaign.CampaignEnd = *input.CampaignEnd
	}
	if input.AllowedCountries != nil && (contains(allowed, "all") || contains(allowed, "allowed_countries")) {
		campaign.AllowedCountries = *input.AllowedCountries
	}
	if input.AllowedLanguages != nil && (contains(allowed, "all") || contains(allowed, "allowed_languages")) {
		campaign.AllowedLanguages = *input.AllowedLanguages
	}
	if input.MinFollowers != nil && (contains(allowed, "all") || contains(allowed, "min_followers")) {
		campaign.MinFollowers = *input.MinFollowers
	}
	if input.Platforms != nil && (contains(allowed, "all") || contains(allowed, "platforms")) {
		campaign.Platforms = *input.Platforms
	}
	if input.CreatorCategories != nil && (contains(allowed, "all") || contains(allowed, "creator_categories")) {
		campaign.CreatorCategories = *input.CreatorCategories
	}
	if input.MinDurationSecs != nil && (contains(allowed, "all") || contains(allowed, "min_duration_secs")) {
		campaign.MinDurationSecs = *input.MinDurationSecs
	}
	if input.MaxDurationSecs != nil && (contains(allowed, "all") || contains(allowed, "max_duration_secs")) {
		campaign.MaxDurationSecs = *input.MaxDurationSecs
	}
	if input.AspectRatio != nil && (contains(allowed, "all") || contains(allowed, "aspect_ratio")) {
		campaign.AspectRatio = *input.AspectRatio
	}
	if input.TalkingPoints != nil && (contains(allowed, "all") || contains(allowed, "talking_points")) {
		campaign.TalkingPoints = *input.TalkingPoints
	}
	if input.ProhibitedClaims != nil && (contains(allowed, "all") || contains(allowed, "prohibited_claims")) {
		campaign.ProhibitedClaims = *input.ProhibitedClaims
	}
	if input.Hashtags != nil && (contains(allowed, "all") || contains(allowed, "hashtags")) {
		campaign.Hashtags = *input.Hashtags
	}
	if input.CTAInstructions != nil && (contains(allowed, "all") || contains(allowed, "cta_instructions")) {
		campaign.CTAInstructions = *input.CTAInstructions
	}

	campaign.Version++
	campaign.UpdatedAt = time.Now()

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Cancel soft-deletes a campaign.
func (s *CampaignService) Cancel(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error) {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	if campaign.BrandProfileID != brandProfileID {
		return nil, domain.ErrCampaignNotOwned
	}

	if !campaign.Status.IsValidTransition(domain.CampaignStatusCancelled) {
		return nil, domain.ErrInvalidTransition
	}

	now := time.Now()
	campaign.Status = domain.CampaignStatusCancelled
	campaign.DeletedAt = &now
	campaign.Version++
	campaign.UpdatedAt = now

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// Publish transitions a campaign from draft to published status.
func (s *CampaignService) Publish(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error) {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	if campaign.BrandProfileID != brandProfileID {
		return nil, domain.ErrCampaignNotOwned
	}

	if !campaign.Status.IsValidTransition(domain.CampaignStatusPublished) {
		return nil, domain.ErrInvalidTransition
	}

	// Validate budget
	if campaign.TotalBudget <= 0 {
		return nil, domain.ErrBudgetRequired
	}

	now := time.Now()
	campaign.Status = domain.CampaignStatusPublished
	campaign.Version++
	campaign.UpdatedAt = now

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// Pause transitions a campaign from active to paused status.
func (s *CampaignService) Pause(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error) {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	if campaign.BrandProfileID != brandProfileID {
		return nil, domain.ErrCampaignNotOwned
	}

	if !campaign.Status.IsValidTransition(domain.CampaignStatusPaused) {
		return nil, domain.ErrInvalidTransition
	}

	now := time.Now()
	campaign.Status = domain.CampaignStatusPaused
	campaign.Version++
	campaign.UpdatedAt = now

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// Resume transitions a campaign from paused to active status.
func (s *CampaignService) Resume(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error) {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	if campaign.BrandProfileID != brandProfileID {
		return nil, domain.ErrCampaignNotOwned
	}

	if !campaign.Status.IsValidTransition(domain.CampaignStatusActive) {
		return nil, domain.ErrInvalidTransition
	}

	now := time.Now()
	campaign.Status = domain.CampaignStatusActive
	campaign.Version++
	campaign.UpdatedAt = now

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// Complete transitions a campaign from active to completed status.
func (s *CampaignService) Complete(ctx context.Context, campaignID, brandProfileID uuid.UUID) (*domain.Campaign, error) {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	if campaign.BrandProfileID != brandProfileID {
		return nil, domain.ErrCampaignNotOwned
	}

	if !campaign.Status.IsValidTransition(domain.CampaignStatusCompleted) {
		return nil, domain.ErrInvalidTransition
	}

	now := time.Now()
	campaign.Status = domain.CampaignStatusCompleted
	campaign.Version++
	campaign.UpdatedAt = now

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

// TransitionPublishedToActive automatically transitions published campaigns to active
// when their submission_deadline has passed. Returns the number of campaigns transitioned.
func (s *CampaignService) TransitionPublishedToActive(ctx context.Context) (int, error) {
	now := time.Now()
	campaigns, err := s.campaignRepo.ByStatusAndDeadline(ctx, domain.CampaignStatusPublished, now)
	if err != nil {
		return 0, err
	}

	transitioned := 0
	for _, campaign := range campaigns {
		if campaign.Status.IsValidTransition(domain.CampaignStatusActive) {
			campaign.Status = domain.CampaignStatusActive
			campaign.Version++
			campaign.UpdatedAt = now

			if err := s.campaignRepo.Update(ctx, campaign); err != nil {
				// Log error but continue with other campaigns
				continue
			}
			transitioned++
		}
	}

	return transitioned, nil
}
