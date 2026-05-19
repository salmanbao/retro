package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

var (
	ErrNotEligible          = errors.New("editor not eligible to create submission")
	ErrDeadlinePassed       = errors.New("submission deadline has passed")
	ErrCampaignNotAccepting = errors.New("campaign is not accepting submissions")
	ErrNoCreativeBrief      = errors.New("campaign has no creative brief")
	ErrNoAssets             = errors.New("campaign has no assets")
	ErrDuplicateSubmission  = errors.New("editor already has a non-draft submission for this campaign")
	ErrCannotEditNonDraft   = errors.New("cannot edit non-draft submission")
	ErrCannotSubmitNonDraft = errors.New("cannot submit non-draft submission")
	ErrCannotWithdraw       = errors.New("cannot withdraw submission in current state")
	ErrNotOwner             = errors.New("editor does not own this submission")
)

type SubmissionService struct {
	submissionRepo repository.SubmissionRepository
	campaignRepo   repository.CampaignRepository
	briefRepo      repository.CreativeBriefRepository
	assetMetaRepo  repository.AssetMetadataRepository
}

func NewSubmissionService(
	submissionRepo repository.SubmissionRepository,
	campaignRepo repository.CampaignRepository,
	briefRepo repository.CreativeBriefRepository,
	assetMetaRepo repository.AssetMetadataRepository,
) *SubmissionService {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		campaignRepo:   campaignRepo,
		briefRepo:      briefRepo,
		assetMetaRepo:  assetMetaRepo,
	}
}

// CheckEligibility validates whether an editor can create a submission for a campaign.
func (s *SubmissionService) CheckEligibility(ctx context.Context, editorProfileID, campaignID uuid.UUID) error {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return err
	}
	if campaign == nil {
		return domain.ErrCampaignNotFound
	}

	// Check campaign is published or active
	if campaign.Status != domain.CampaignStatusPublished && campaign.Status != domain.CampaignStatusActive {
		return ErrCampaignNotAccepting
	}

	// Check deadline has not passed
	if time.Now().After(campaign.SubmissionDeadline) {
		return ErrDeadlinePassed
	}

	// Check creative brief exists
	brief, err := s.briefRepo.ByCampaignID(ctx, campaignID)
	if err != nil {
		return err
	}
	if brief == nil {
		return ErrNoCreativeBrief
	}

	// Check at least one asset exists
	assets, _, err := s.assetMetaRepo.ListByCampaign(ctx, campaignID, 1, 1)
	if err != nil {
		return err
	}
	if len(assets) == 0 {
		return ErrNoAssets
	}

	// Check no duplicate non-draft submission exists
	existing, err := s.submissionRepo.FindNonDraftByEditorAndCampaign(ctx, editorProfileID, campaignID)
	if err != nil && !errors.Is(err, repository.ErrSubmissionNotFound) {
		return err
	}
	if existing != nil {
		return ErrDuplicateSubmission
	}

	return nil
}

// CreateDraft creates a new draft submission for an eligible editor.
func (s *SubmissionService) CreateDraft(ctx context.Context, editorProfileID uuid.UUID, input CreateSubmissionInput) (*domain.Submission, error) {
	if err := s.CheckEligibility(ctx, editorProfileID, input.CampaignID); err != nil {
		return nil, err
	}

	submission := &domain.Submission{
		CampaignID:      input.CampaignID,
		EditorProfileID: editorProfileID,
		Title:           input.Title,
		Description:     input.Description,
		VideoURL:        input.VideoURL,
		ThumbnailURL:    input.ThumbnailURL,
		DurationSeconds: input.DurationSeconds,
		Notes:           input.Notes,
		Tags:            input.Tags,
		Status:          domain.SubmissionStatusDraft,
	}

	if err := submission.Validate(); err != nil {
		return nil, err
	}

	if err := s.submissionRepo.Create(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}

// CreateSubmissionInput contains fields needed to create a submission.
type CreateSubmissionInput struct {
	CampaignID      uuid.UUID
	Title           string
	Description     string
	VideoURL        string
	ThumbnailURL    string
	DurationSeconds int
	Notes           string
	Tags            []string
}

// GetByCampaignID retrieves all submissions for a campaign (for Brand listing).
func (s *SubmissionService) GetByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*domain.Submission, error) {
	campaign, err := s.campaignRepo.ByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, domain.ErrCampaignNotFound
	}
	return s.submissionRepo.GetByCampaignID(ctx, campaignID)
}

// GetByID retrieves a submission by ID.
func (s *SubmissionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return s.submissionRepo.GetByID(ctx, id)
}

// UpdateDraft updates a draft submission.
func (s *SubmissionService) UpdateDraft(ctx context.Context, editorProfileID uuid.UUID, submission *domain.Submission, input UpdateSubmissionInput) error {
	// Only the owning editor can edit
	if submission.EditorProfileID != editorProfileID {
		return ErrNotOwner
	}

	// Only draft submissions are editable
	if !submission.CanEdit() {
		return ErrCannotEditNonDraft
	}

	// Apply updates
	if input.Title != "" {
		submission.Title = input.Title
	}
	if input.Description != "" {
		submission.Description = input.Description
	}
	if input.VideoURL != "" {
		submission.VideoURL = input.VideoURL
	}
	if input.ThumbnailURL != "" {
		submission.ThumbnailURL = input.ThumbnailURL
	}
	if input.DurationSeconds > 0 {
		submission.DurationSeconds = input.DurationSeconds
	}
	if input.Notes != "" {
		submission.Notes = input.Notes
	}
	if input.Tags != nil {
		submission.Tags = input.Tags
	}

	if err := submission.Validate(); err != nil {
		return err
	}

	return s.submissionRepo.Update(ctx, submission)
}

// UpdateSubmissionInput contains fields that can be updated on a draft submission.
type UpdateSubmissionInput struct {
	Title           string
	Description     string
	VideoURL        string
	ThumbnailURL    string
	DurationSeconds int
	Notes           string
	Tags            []string
}

// Submit transitions a draft submission to submitted state.
func (s *SubmissionService) Submit(ctx context.Context, editorProfileID uuid.UUID, submission *domain.Submission) error {
	// Only the owning editor can submit
	if submission.EditorProfileID != editorProfileID {
		return ErrNotOwner
	}

	// Only draft submissions can be submitted
	if !submission.CanSubmit() {
		return ErrCannotSubmitNonDraft
	}

	// Validate state transition
	if !submission.Status.CanTransitionTo(domain.SubmissionStatusSubmitted) {
		return domain.ErrInvalidTransition
	}

	now := time.Now()
	submission.Status = domain.SubmissionStatusSubmitted
	submission.SubmittedAt = &now

	return s.submissionRepo.Update(ctx, submission)
}

// Withdraw transitions a submitted or under_review submission to withdrawn state.
func (s *SubmissionService) Withdraw(ctx context.Context, editorProfileID uuid.UUID, submission *domain.Submission) error {
	// Only the owning editor can withdraw
	if submission.EditorProfileID != editorProfileID {
		return ErrNotOwner
	}

	// Only submitted or under_review submissions can be withdrawn
	if !submission.CanWithdraw() {
		return ErrCannotWithdraw
	}

	// Validate state transition
	if !submission.Status.CanTransitionTo(domain.SubmissionStatusWithdrawn) {
		return domain.ErrInvalidTransition
	}

	now := time.Now()
	submission.Status = domain.SubmissionStatusWithdrawn
	submission.WithdrawnAt = &now

	return s.submissionRepo.Update(ctx, submission)
}
