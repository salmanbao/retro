package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type SubmissionStatus string

const (
	SubmissionStatusDraft       SubmissionStatus = "draft"
	SubmissionStatusSubmitted   SubmissionStatus = "submitted"
	SubmissionStatusUnderReview SubmissionStatus = "under_review"
	SubmissionStatusShortlisted SubmissionStatus = "shortlisted"
	SubmissionStatusApproved    SubmissionStatus = "approved"
	SubmissionStatusRejected    SubmissionStatus = "rejected"
	SubmissionStatusWithdrawn   SubmissionStatus = "withdrawn"
)

func (s SubmissionStatus) IsValid() bool {
	switch s {
	case SubmissionStatusDraft, SubmissionStatusSubmitted, SubmissionStatusUnderReview,
		SubmissionStatusShortlisted, SubmissionStatusApproved, SubmissionStatusRejected,
		SubmissionStatusWithdrawn:
		return true
	}
	return false
}

func (s SubmissionStatus) CanTransitionTo(target SubmissionStatus) bool {
	switch s {
	case SubmissionStatusDraft:
		return target == SubmissionStatusSubmitted
	case SubmissionStatusSubmitted:
		return target == SubmissionStatusUnderReview || target == SubmissionStatusWithdrawn
	case SubmissionStatusUnderReview:
		return target == SubmissionStatusShortlisted || target == SubmissionStatusRejected || target == SubmissionStatusWithdrawn
	case SubmissionStatusShortlisted:
		return target == SubmissionStatusApproved
	// Terminal states: approved, rejected, withdrawn
	case SubmissionStatusApproved, SubmissionStatusRejected, SubmissionStatusWithdrawn:
		return false
	}
	return false
}

var (
	ErrSubmissionNotFound   = errors.New("submission not found")
	ErrNotEligible          = errors.New("editor not eligible to create submission")
	ErrCannotEditNonDraft   = errors.New("cannot edit non-draft submission")
	ErrCannotSubmitNonDraft = errors.New("cannot submit non-draft submission")
	ErrCannotWithdraw       = errors.New("cannot withdraw submission in current state")
	ErrDeadlinePassed       = errors.New("submission deadline has passed")
	ErrDuplicateSubmission  = errors.New("editor already has a non-draft submission for this campaign")
	ErrCampaignNotEditable  = errors.New("campaign is not in an editable state")
	ErrNotOwner             = errors.New("editor does not own this submission")
	ErrNotCampaignOwner     = errors.New("editor is not the campaign owner")
	ErrProfileNotEditor     = errors.New("profile is not an editor")
)

type Submission struct {
	ID              uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CampaignID      uuid.UUID        `json:"campaign_id" gorm:"type:uuid;not null;index"`
	EditorProfileID uuid.UUID        `json:"editor_profile_id" gorm:"type:uuid;not null;index"`
	Title           string           `json:"title" gorm:"type:varchar(200);not null"`
	Description     string           `json:"description" gorm:"type:text"`
	VideoURL        string           `json:"video_url" gorm:"type:varchar(2000);not null"`
	ThumbnailURL    string           `json:"thumbnail_url" gorm:"type:varchar(2000)"`
	DurationSeconds int              `json:"duration_seconds" gorm:"not null"`
	Notes           string           `json:"notes" gorm:"type:text"`
	Tags            JSONBArray       `json:"tags" gorm:"type:jsonb;default:'[]'::jsonb"`
	Status          SubmissionStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft'"`
	CreatedAt       time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
	SubmittedAt     *time.Time       `json:"submitted_at,omitempty"`
	ReviewedAt      *time.Time       `json:"reviewed_at,omitempty"`
	WithdrawnAt     *time.Time       `json:"withdrawn_at,omitempty"`
	DeletedAt       *time.Time       `json:"deleted_at,omitempty" gorm:"index"`
}

func (s *Submission) CanEdit() bool {
	return s.Status == SubmissionStatusDraft
}

func (s *Submission) CanSubmit() bool {
	return s.Status == SubmissionStatusDraft
}

func (s *Submission) CanWithdraw() bool {
	return s.Status == SubmissionStatusSubmitted || s.Status == SubmissionStatusUnderReview
}

func (s *Submission) IsDeleted() bool {
	return s.DeletedAt != nil
}

func (s *Submission) Validate() error {
	if s.CampaignID == uuid.Nil {
		return errors.New("campaign_id is required")
	}
	if s.EditorProfileID == uuid.Nil {
		return errors.New("editor_profile_id is required")
	}
	if s.Title == "" {
		return errors.New("title is required")
	}
	if len(s.Title) > 200 {
		return errors.New("title must be 200 characters or less")
	}
	if len(s.Description) > 5000 {
		return errors.New("description must be 5000 characters or less")
	}
	if s.VideoURL == "" {
		return errors.New("video_url is required")
	}
	if len(s.VideoURL) > 2000 {
		return errors.New("video_url must be 2000 characters or less")
	}
	if s.ThumbnailURL != "" && len(s.ThumbnailURL) > 2000 {
		return errors.New("thumbnail_url must be 2000 characters or less")
	}
	if s.DurationSeconds <= 0 {
		return errors.New("duration_seconds must be positive")
	}
	if s.Notes != "" && len(s.Notes) > 2000 {
		return errors.New("notes must be 2000 characters or less")
	}
	if !s.Status.IsValid() {
		return errors.New("invalid submission status")
	}
	return nil
}
