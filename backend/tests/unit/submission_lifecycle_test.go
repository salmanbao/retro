package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

func TestLifecycle_SubmitDraftSubmission(t *testing.T) {
	editorID := uuid.New()
	submission := &domain.Submission{
		ID:              uuid.New(),
		CampaignID:      uuid.New(),
		EditorProfileID: editorID,
		Title:           "Test Submission",
		VideoURL:        "https://example.com/video.mp4",
		DurationSeconds: 60,
		Status:          domain.SubmissionStatusDraft,
	}

	assert.True(t, submission.CanSubmit())
	assert.True(t, submission.Status.CanTransitionTo(domain.SubmissionStatusSubmitted))

	// Simulate submit
	now := time.Now()
	submission.Status = domain.SubmissionStatusSubmitted
	submission.SubmittedAt = &now

	assert.Equal(t, domain.SubmissionStatusSubmitted, submission.Status)
	assert.NotNil(t, submission.SubmittedAt)
}

func TestLifecycle_WithdrawSubmittedSubmission(t *testing.T) {
	editorID := uuid.New()
	now := time.Now()
	submission := &domain.Submission{
		ID:              uuid.New(),
		CampaignID:      uuid.New(),
		EditorProfileID: editorID,
		Title:           "Test Submission",
		VideoURL:        "https://example.com/video.mp4",
		DurationSeconds: 60,
		Status:          domain.SubmissionStatusSubmitted,
		SubmittedAt:     &now,
	}

	assert.True(t, submission.CanWithdraw())
	assert.True(t, submission.Status.CanTransitionTo(domain.SubmissionStatusWithdrawn))

	// Simulate withdraw
	submission.Status = domain.SubmissionStatusWithdrawn
	submission.WithdrawnAt = &now

	assert.Equal(t, domain.SubmissionStatusWithdrawn, submission.Status)
	assert.NotNil(t, submission.WithdrawnAt)
}

func TestLifecycle_WithdrawUnderReviewSubmission(t *testing.T) {
	editorID := uuid.New()
	now := time.Now()
	submission := &domain.Submission{
		ID:              uuid.New(),
		CampaignID:      uuid.New(),
		EditorProfileID: editorID,
		Title:           "Test Submission",
		VideoURL:        "https://example.com/video.mp4",
		DurationSeconds: 60,
		Status:          domain.SubmissionStatusUnderReview,
		SubmittedAt:     &now,
	}

	assert.True(t, submission.CanWithdraw())
	assert.True(t, submission.Status.CanTransitionTo(domain.SubmissionStatusWithdrawn))
}

func TestLifecycle_CannotWithdrawShortlisted(t *testing.T) {
	editorID := uuid.New()
	now := time.Now()
	submission := &domain.Submission{
		ID:              uuid.New(),
		CampaignID:      uuid.New(),
		EditorProfileID: editorID,
		Title:           "Test Submission",
		VideoURL:        "https://example.com/video.mp4",
		DurationSeconds: 60,
		Status:          domain.SubmissionStatusShortlisted,
		SubmittedAt:     &now,
	}

	assert.False(t, submission.CanWithdraw())
}

func TestLifecycle_CannotWithdrawApproved(t *testing.T) {
	editorID := uuid.New()
	now := time.Now()
	submission := &domain.Submission{
		ID:              uuid.New(),
		CampaignID:      uuid.New(),
		EditorProfileID: editorID,
		Title:           "Test Submission",
		VideoURL:        "https://example.com/video.mp4",
		DurationSeconds: 60,
		Status:          domain.SubmissionStatusApproved,
		SubmittedAt:     &now,
	}

	assert.False(t, submission.CanWithdraw())
}

func TestLifecycle_CannotWithdrawRejected(t *testing.T) {
	editorID := uuid.New()
	now := time.Now()
	submission := &domain.Submission{
		ID:              uuid.New(),
		CampaignID:      uuid.New(),
		EditorProfileID: editorID,
		Title:           "Test Submission",
		VideoURL:        "https://example.com/video.mp4",
		DurationSeconds: 60,
		Status:          domain.SubmissionStatusRejected,
		SubmittedAt:     &now,
	}

	assert.False(t, submission.CanWithdraw())
}

func TestLifecycle_CannotSubmitNonDraft(t *testing.T) {
	testCases := []struct {
		status domain.SubmissionStatus
	}{
		{domain.SubmissionStatusSubmitted},
		{domain.SubmissionStatusUnderReview},
		{domain.SubmissionStatusShortlisted},
		{domain.SubmissionStatusApproved},
		{domain.SubmissionStatusRejected},
		{domain.SubmissionStatusWithdrawn},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			submission := &domain.Submission{Status: tc.status}
			assert.False(t, submission.CanSubmit())
		})
	}
}

func TestLifecycle_TerminalStatesBlockAllTransitions(t *testing.T) {
	terminalStates := []domain.SubmissionStatus{
		domain.SubmissionStatusApproved,
		domain.SubmissionStatusRejected,
		domain.SubmissionStatusWithdrawn,
	}

	for _, status := range terminalStates {
		submission := &domain.Submission{Status: status}
		assert.False(t, submission.CanEdit())
		assert.False(t, submission.CanSubmit())
		assert.False(t, submission.CanWithdraw())
	}
}
