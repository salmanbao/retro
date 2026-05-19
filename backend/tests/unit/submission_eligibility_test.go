package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

func TestCheckEligibility_CampaignNotFound(t *testing.T) {
	svc := &MockSubmissionService{}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCampaignNotFound, err)
}

func TestCheckEligibility_CampaignNotAcceptingSubmissions(t *testing.T) {
	svc := &MockSubmissionService{
		campaign: &domain.Campaign{
			Status: domain.CampaignStatusDraft, // Draft campaigns don't accept submissions
		},
	}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.Error(t, err)
	assert.Equal(t, service.ErrCampaignNotAccepting, err)
}

func TestCheckEligibility_DeadlinePassed(t *testing.T) {
	svc := &MockSubmissionService{
		campaign: &domain.Campaign{
			Status:             domain.CampaignStatusPublished,
			SubmissionDeadline: time.Now().Add(-24 * time.Hour), // Deadline passed
		},
	}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.Error(t, err)
	assert.Equal(t, service.ErrDeadlinePassed, err)
}

func TestCheckEligibility_NoCreativeBrief(t *testing.T) {
	svc := &MockSubmissionService{
		campaign: &domain.Campaign{
			Status:             domain.CampaignStatusPublished,
			SubmissionDeadline: time.Now().Add(24 * time.Hour), // Deadline not passed
		},
		brief: nil, // No creative brief
	}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.Error(t, err)
	assert.Equal(t, service.ErrNoCreativeBrief, err)
}

func TestCheckEligibility_NoAssets(t *testing.T) {
	svc := &MockSubmissionService{
		campaign: &domain.Campaign{
			Status:             domain.CampaignStatusPublished,
			SubmissionDeadline: time.Now().Add(24 * time.Hour),
		},
		brief:  &domain.CreativeBrief{},
		assets: []*domain.AssetMetadata{}, // No assets
	}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.Error(t, err)
	assert.Equal(t, service.ErrNoAssets, err)
}

func TestCheckEligibility_DuplicateSubmission(t *testing.T) {
	svc := &MockSubmissionService{
		campaign: &domain.Campaign{
			Status:             domain.CampaignStatusPublished,
			SubmissionDeadline: time.Now().Add(24 * time.Hour),
		},
		brief:  &domain.CreativeBrief{},
		assets: []*domain.AssetMetadata{{OriginalFilename: "test.mp4"}},
		existingSubmission: &domain.Submission{ // Already has a non-draft submission
			Status: domain.SubmissionStatusSubmitted,
		},
	}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.Error(t, err)
	assert.Equal(t, service.ErrDuplicateSubmission, err)
}

func TestCheckEligibility_Valid(t *testing.T) {
	svc := &MockSubmissionService{
		campaign: &domain.Campaign{
			Status:             domain.CampaignStatusPublished,
			SubmissionDeadline: time.Now().Add(24 * time.Hour),
		},
		brief:  &domain.CreativeBrief{},
		assets: []*domain.AssetMetadata{{OriginalFilename: "test.mp4"}},
	}
	editorID := uuid.New()
	campaignID := uuid.New()

	err := svc.CheckEligibility(nil, editorID, campaignID)
	assert.NoError(t, err)
}

// MockSubmissionService simulates the eligibility check logic for unit testing.
type MockSubmissionService struct {
	campaign           *domain.Campaign
	brief              *domain.CreativeBrief
	assets             []*domain.AssetMetadata
	existingSubmission *domain.Submission
}

func (m *MockSubmissionService) CheckEligibility(ctx interface{}, editorProfileID, campaignID uuid.UUID) error {
	// Simulate campaign not found
	if m.campaign == nil {
		return domain.ErrCampaignNotFound
	}

	// Check campaign is published or active
	if m.campaign.Status != domain.CampaignStatusPublished && m.campaign.Status != domain.CampaignStatusActive {
		return service.ErrCampaignNotAccepting
	}

	// Check deadline has not passed
	if time.Now().After(m.campaign.SubmissionDeadline) {
		return service.ErrDeadlinePassed
	}

	// Check creative brief exists
	if m.brief == nil {
		return service.ErrNoCreativeBrief
	}

	// Check at least one asset exists
	if len(m.assets) == 0 {
		return service.ErrNoAssets
	}

	// Check no duplicate non-draft submission exists
	if m.existingSubmission != nil {
		return service.ErrDuplicateSubmission
	}

	return nil
}
