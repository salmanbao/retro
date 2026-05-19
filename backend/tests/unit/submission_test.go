package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

func TestSubmissionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.SubmissionStatus
		want   bool
	}{
		{"draft is valid", domain.SubmissionStatusDraft, true},
		{"submitted is valid", domain.SubmissionStatusSubmitted, true},
		{"under_review is valid", domain.SubmissionStatusUnderReview, true},
		{"shortlisted is valid", domain.SubmissionStatusShortlisted, true},
		{"approved is valid", domain.SubmissionStatusApproved, true},
		{"rejected is valid", domain.SubmissionStatusRejected, true},
		{"withdrawn is valid", domain.SubmissionStatusWithdrawn, true},
		{"invalid status", domain.SubmissionStatus("invalid"), false},
		{"empty status", domain.SubmissionStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubmissionStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     domain.SubmissionStatus
		to       domain.SubmissionStatus
		expected bool
	}{
		// Valid from draft
		{"draft to submitted", domain.SubmissionStatusDraft, domain.SubmissionStatusSubmitted, true},
		{"draft to withdrawn", domain.SubmissionStatusDraft, domain.SubmissionStatusWithdrawn, false},
		{"draft to approved", domain.SubmissionStatusDraft, domain.SubmissionStatusApproved, false},

		// Valid from submitted
		{"submitted to under_review", domain.SubmissionStatusSubmitted, domain.SubmissionStatusUnderReview, true},
		{"submitted to withdrawn", domain.SubmissionStatusSubmitted, domain.SubmissionStatusWithdrawn, true},
		{"submitted to rejected", domain.SubmissionStatusSubmitted, domain.SubmissionStatusRejected, false},

		// Valid from under_review
		{"under_review to shortlisted", domain.SubmissionStatusUnderReview, domain.SubmissionStatusShortlisted, true},
		{"under_review to rejected", domain.SubmissionStatusUnderReview, domain.SubmissionStatusRejected, true},
		{"under_review to withdrawn", domain.SubmissionStatusUnderReview, domain.SubmissionStatusWithdrawn, true},

		// Valid from shortlisted
		{"shortlisted to approved", domain.SubmissionStatusShortlisted, domain.SubmissionStatusApproved, true},
		{"shortlisted to rejected", domain.SubmissionStatusShortlisted, domain.SubmissionStatusRejected, false},

		// Terminal states block all transitions
		{"approved to any", domain.SubmissionStatusApproved, domain.SubmissionStatusDraft, false},
		{"rejected to any", domain.SubmissionStatusRejected, domain.SubmissionStatusDraft, false},
		{"withdrawn to any", domain.SubmissionStatusWithdrawn, domain.SubmissionStatusDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSubmission_CanEdit(t *testing.T) {
	tests := []struct {
		name    string
		status  domain.SubmissionStatus
		canEdit bool
	}{
		{"draft can edit", domain.SubmissionStatusDraft, true},
		{"submitted cannot edit", domain.SubmissionStatusSubmitted, false},
		{"under_review cannot edit", domain.SubmissionStatusUnderReview, false},
		{"shortlisted cannot edit", domain.SubmissionStatusShortlisted, false},
		{"approved cannot edit", domain.SubmissionStatusApproved, false},
		{"rejected cannot edit", domain.SubmissionStatusRejected, false},
		{"withdrawn cannot edit", domain.SubmissionStatusWithdrawn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Submission{Status: tt.status}
			assert.Equal(t, tt.canEdit, s.CanEdit())
		})
	}
}

func TestSubmission_CanSubmit(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.SubmissionStatus
		canSubmit bool
	}{
		{"draft can submit", domain.SubmissionStatusDraft, true},
		{"submitted cannot submit", domain.SubmissionStatusSubmitted, false},
		{"under_review cannot submit", domain.SubmissionStatusUnderReview, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Submission{Status: tt.status}
			assert.Equal(t, tt.canSubmit, s.CanSubmit())
		})
	}
}

func TestSubmission_CanWithdraw(t *testing.T) {
	tests := []struct {
		name        string
		status      domain.SubmissionStatus
		canWithdraw bool
	}{
		{"draft cannot withdraw", domain.SubmissionStatusDraft, false},
		{"submitted can withdraw", domain.SubmissionStatusSubmitted, true},
		{"under_review can withdraw", domain.SubmissionStatusUnderReview, true},
		{"shortlisted cannot withdraw", domain.SubmissionStatusShortlisted, false},
		{"approved cannot withdraw", domain.SubmissionStatusApproved, false},
		{"rejected cannot withdraw", domain.SubmissionStatusRejected, false},
		{"withdrawn cannot withdraw", domain.SubmissionStatusWithdrawn, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Submission{Status: tt.status}
			assert.Equal(t, tt.canWithdraw, s.CanWithdraw())
		})
	}
}

func TestSubmission_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sub     domain.Submission
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid submission",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 60,
				Status:          domain.SubmissionStatusDraft,
			},
			wantErr: false,
		},
		{
			name: "missing campaign_id",
			sub: domain.Submission{
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 60,
			},
			wantErr: true,
			errMsg:  "campaign_id is required",
		},
		{
			name: "missing editor_profile_id",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				Title:           "My Video",
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 60,
			},
			wantErr: true,
			errMsg:  "editor_profile_id is required",
		},
		{
			name: "missing title",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 60,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "title too long",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           string(make([]byte, 201)),
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 60,
			},
			wantErr: true,
			errMsg:  "title must be 200 characters or less",
		},
		{
			name: "missing video_url",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				DurationSeconds: 60,
			},
			wantErr: true,
			errMsg:  "video_url is required",
		},
		{
			name: "video_url too long",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				VideoURL:        string(make([]byte, 2001)),
				DurationSeconds: 60,
			},
			wantErr: true,
			errMsg:  "video_url must be 2000 characters or less",
		},
		{
			name: "zero duration",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 0,
			},
			wantErr: true,
			errMsg:  "duration_seconds must be positive",
		},
		{
			name: "negative duration",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: -1,
			},
			wantErr: true,
			errMsg:  "duration_seconds must be positive",
		},
		{
			name: "invalid status",
			sub: domain.Submission{
				CampaignID:      uuid.New(),
				EditorProfileID: uuid.New(),
				Title:           "My Video",
				VideoURL:        "https://example.com/video.mp4",
				DurationSeconds: 60,
				Status:          domain.SubmissionStatus("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid submission status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubmission_IsDeleted(t *testing.T) {
	t.Run("not deleted when deleted_at is nil", func(t *testing.T) {
		s := &domain.Submission{DeletedAt: nil}
		assert.False(t, s.IsDeleted())
	})

	t.Run("deleted when deleted_at is set", func(t *testing.T) {
		s := &domain.Submission{DeletedAt: &time.Time{}}
		assert.True(t, s.IsDeleted())
	})
}

func TestSubmissionSoftDeletion(t *testing.T) {
	t.Run("submission is not deleted when deleted_at is nil", func(t *testing.T) {
		s := &domain.Submission{DeletedAt: nil}
		assert.False(t, s.IsDeleted())
	})

	t.Run("submission is deleted when deleted_at is set", func(t *testing.T) {
		s := &domain.Submission{DeletedAt: &time.Time{}}
		assert.True(t, s.IsDeleted())
	})
}
