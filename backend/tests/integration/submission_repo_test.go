package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

func setupSubmissionTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping integration test: database not available")
	}
	db.AutoMigrate(&domain.Submission{})
	return db
}

// submissionRepo wraps gorm.DB for testing
type submissionRepo struct {
	db *gorm.DB
}

func (r *submissionRepo) Create(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}
func (r *submissionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&submission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrSubmissionNotFound
		}
		return nil, err
	}
	return &submission, nil
}
func (r *submissionRepo) GetByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*domain.Submission, error) {
	var submissions []*domain.Submission
	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND deleted_at IS NULL", campaignID).
		Order("created_at DESC").Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}
func (r *submissionRepo) GetByEditorProfileID(ctx context.Context, editorProfileID uuid.UUID) ([]*domain.Submission, error) {
	var submissions []*domain.Submission
	err := r.db.WithContext(ctx).
		Where("editor_profile_id = ? AND deleted_at IS NULL", editorProfileID).
		Order("created_at DESC").Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}
func (r *submissionRepo) Update(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Save(submission).Error
}
func (r *submissionRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Submission{}).
		Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
func (r *submissionRepo) FindNonDraftByEditorAndCampaign(ctx context.Context, editorProfileID, campaignID uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).
		Where("editor_profile_id = ? AND campaign_id = ? AND status != ? AND deleted_at IS NULL",
			editorProfileID, campaignID, domain.SubmissionStatusDraft).First(&submission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrSubmissionNotFound
		}
		return nil, err
	}
	return &submission, nil
}
func (r *submissionRepo) ExistsByCampaignID(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Submission{}).
		Where("campaign_id = ? AND deleted_at IS NULL", campaignID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func TestSubmissionRepository_CRUD(t *testing.T) {
	db := setupSubmissionTestDB(t)
	repo := &submissionRepo{db: db}

	// Use unique IDs to avoid conflicts
	campaignID := uuid.New()
	editorID := uuid.New()
	submissionID := uuid.New()

	// Cleanup
	defer db.Where("id = ?", submissionID).Delete(&domain.Submission{})

	t.Run("create and get submission", func(t *testing.T) {
		submission := &domain.Submission{
			ID:              submissionID,
			CampaignID:      campaignID,
			EditorProfileID: editorID,
			Title:           "Integration Test",
			VideoURL:        "https://example.com/video.mp4",
			DurationSeconds: 60,
			Status:          domain.SubmissionStatusDraft,
		}

		err := repo.Create(context.Background(), submission)
		assert.NoError(t, err)

		retrieved, err := repo.GetByID(context.Background(), submissionID)
		assert.NoError(t, err)
		assert.Equal(t, submissionID, retrieved.ID)
		assert.Equal(t, "Integration Test", retrieved.Title)
	})

	t.Run("update submission", func(t *testing.T) {
		submission := &domain.Submission{
			ID:              submissionID,
			CampaignID:      campaignID,
			EditorProfileID: editorID,
			Title:           "Original Title",
			VideoURL:        "https://example.com/video.mp4",
			DurationSeconds: 60,
			Status:          domain.SubmissionStatusDraft,
		}
		repo.Create(context.Background(), submission)

		submission.Title = "Updated Title"
		now := time.Now()
		submission.Status = domain.SubmissionStatusSubmitted
		submission.SubmittedAt = &now

		err := repo.Update(context.Background(), submission)
		assert.NoError(t, err)

		retrieved, _ := repo.GetByID(context.Background(), submissionID)
		assert.Equal(t, "Updated Title", retrieved.Title)
		assert.Equal(t, domain.SubmissionStatusSubmitted, retrieved.Status)
	})

	t.Run("soft delete submission", func(t *testing.T) {
		newID := uuid.New()
		submission := &domain.Submission{
			ID:              newID,
			CampaignID:      campaignID,
			EditorProfileID: editorID,
			Title:           "To Delete",
			VideoURL:        "https://example.com/video.mp4",
			DurationSeconds: 60,
			Status:          domain.SubmissionStatusDraft,
		}
		repo.Create(context.Background(), submission)
		defer db.Where("id = ?", newID).Delete(&domain.Submission{})

		err := repo.SoftDelete(context.Background(), newID)
		assert.NoError(t, err)

		_, err = repo.GetByID(context.Background(), newID)
		assert.Equal(t, repository.ErrSubmissionNotFound, err)
	})

	t.Run("find non-draft by editor and campaign", func(t *testing.T) {
		// First create a submitted submission
		now := time.Now()
		submission := &domain.Submission{
			ID:              uuid.New(),
			CampaignID:      campaignID,
			EditorProfileID: editorID,
			Title:           "Submitted",
			VideoURL:        "https://example.com/video.mp4",
			DurationSeconds: 60,
			Status:          domain.SubmissionStatusSubmitted,
			SubmittedAt:     &now,
		}
		repo.Create(context.Background(), submission)

		found, err := repo.FindNonDraftByEditorAndCampaign(context.Background(), editorID, campaignID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, domain.SubmissionStatusSubmitted, found.Status)
	})

	t.Run("get by campaign ID", func(t *testing.T) {
		submissions, err := repo.GetByCampaignID(context.Background(), campaignID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(submissions), 1)
	})
}
