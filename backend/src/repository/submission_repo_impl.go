package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
)

// SubmissionDB implements SubmissionRepository using GORM.
type SubmissionDB struct {
	db *gorm.DB
}

// NewSubmissionDB creates a new SubmissionDB.
func NewSubmissionDB(db *gorm.DB) *SubmissionDB {
	return &SubmissionDB{db: db}
}

// Create inserts a new submission.
func (r *SubmissionDB) Create(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

// GetByID retrieves a submission by ID (excludes soft-deleted).
func (r *SubmissionDB) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&submission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}
	return &submission, nil
}

// GetByCampaignID retrieves all non-deleted submissions for a campaign.
func (r *SubmissionDB) GetByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*domain.Submission, error) {
	var submissions []*domain.Submission
	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND deleted_at IS NULL", campaignID).
		Order("created_at DESC").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

// GetByEditorProfileID retrieves all non-deleted submissions for an editor.
func (r *SubmissionDB) GetByEditorProfileID(ctx context.Context, editorProfileID uuid.UUID) ([]*domain.Submission, error) {
	var submissions []*domain.Submission
	err := r.db.WithContext(ctx).
		Where("editor_profile_id = ? AND deleted_at IS NULL", editorProfileID).
		Order("created_at DESC").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

// Update updates an existing submission.
func (r *SubmissionDB) Update(ctx context.Context, submission *domain.Submission) error {
	submission.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(submission).Error
}

// SoftDelete soft-deletes a submission by setting deleted_at.
func (r *SubmissionDB) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Submission{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// FindNonDraftByEditorAndCampaign finds a non-draft submission for an editor+campaign.
// Returns ErrSubmissionNotFound if none exists.
func (r *SubmissionDB) FindNonDraftByEditorAndCampaign(ctx context.Context, editorProfileID, campaignID uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).
		Where("editor_profile_id = ? AND campaign_id = ? AND status != ? AND deleted_at IS NULL",
			editorProfileID, campaignID, domain.SubmissionStatusDraft).
		First(&submission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}
	return &submission, nil
}

// ExistsByCampaignID checks if any non-deleted submission exists for a campaign.
func (r *SubmissionDB) ExistsByCampaignID(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Submission{}).
		Where("campaign_id = ? AND deleted_at IS NULL", campaignID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
