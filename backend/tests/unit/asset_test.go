package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

func TestAssetMetadataValidation(t *testing.T) {
	t.Run("valid asset passes validation", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.New(),
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			UploadedByProfileID: uuid.New(),
			ProcessingStatus:    domain.ProcessingStatusPending,
		}
		err := asset.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing campaign ID fails validation", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.Nil,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
			UploadedByProfileID: uuid.New(),
		}
		err := asset.Validate()
		assert.ErrorIs(t, err, domain.ErrAssetCampaignRequired)
	})

	t.Run("invalid category fails validation", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.New(),
			Category:            "invalid_category",
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
			UploadedByProfileID: uuid.New(),
		}
		err := asset.Validate()
		assert.ErrorIs(t, err, domain.ErrAssetCategoryInvalid)
	})

	t.Run("missing filename fails validation", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.New(),
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
			UploadedByProfileID: uuid.New(),
		}
		err := asset.Validate()
		assert.ErrorIs(t, err, domain.ErrAssetFilenameRequired)
	})

	t.Run("invalid checksum length fails validation", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.New(),
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "short",
			UploadedByProfileID: uuid.New(),
		}
		err := asset.Validate()
		assert.ErrorIs(t, err, domain.ErrAssetChecksumLength)
	})

	t.Run("zero file size fails validation", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.New(),
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       0,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
			UploadedByProfileID: uuid.New(),
		}
		err := asset.Validate()
		assert.ErrorIs(t, err, domain.ErrAssetFileSizeInvalid)
	})
}

func TestAssetCategories(t *testing.T) {
	t.Run("all valid categories return true", func(t *testing.T) {
		categories := []domain.AssetCategory{
			domain.AssetCategoryRawFootage,
			domain.AssetCategoryProductImages,
			domain.AssetCategoryLogos,
			domain.AssetCategoryBrandGuidelines,
			domain.AssetCategoryScripts,
			domain.AssetCategoryExampleVideos,
			domain.AssetCategoryMusicReferences,
			domain.AssetCategoryLegalDocuments,
			domain.AssetCategoryOther,
		}
		for _, cat := range categories {
			assert.True(t, cat.IsValid(), "expected %s to be valid", cat)
		}
	})

	t.Run("invalid category returns false", func(t *testing.T) {
		cat := domain.AssetCategory("invalid")
		assert.False(t, cat.IsValid())
	})
}

func TestProcessingStatuses(t *testing.T) {
	t.Run("all valid statuses return true", func(t *testing.T) {
		statuses := []domain.ProcessingStatus{
			domain.ProcessingStatusPending,
			domain.ProcessingStatusUploaded,
			domain.ProcessingStatusProcessing,
			domain.ProcessingStatusReady,
			domain.ProcessingStatusFailed,
			domain.ProcessingStatusArchived,
		}
		for _, status := range statuses {
			assert.True(t, status.IsValid(), "expected %s to be valid", status)
		}
	})

	t.Run("invalid status returns false", func(t *testing.T) {
		status := domain.ProcessingStatus("invalid")
		assert.False(t, status.IsValid())
	})
}

func TestVirusScanStatuses(t *testing.T) {
	t.Run("all valid statuses return true", func(t *testing.T) {
		statuses := []domain.VirusScanStatus{
			domain.VirusScanStatusNotScanned,
			domain.VirusScanStatusScanning,
			domain.VirusScanStatusClean,
			domain.VirusScanStatusInfected,
		}
		for _, status := range statuses {
			assert.True(t, status.IsValid(), "expected %s to be valid", status)
		}
	})
}

func TestAssetSoftDeletion(t *testing.T) {
	t.Run("asset is not deleted when deleted_at is nil", func(t *testing.T) {
		asset := &domain.AssetMetadata{}
		assert.False(t, asset.IsDeleted())
	})

	t.Run("asset is deleted when deleted_at is set", func(t *testing.T) {
		asset := &domain.AssetMetadata{}
		asset.DeletedAt = &time.Time{}
		assert.True(t, asset.IsDeleted())
	})
}

func TestAssetCanUpdate(t *testing.T) {
	t.Run("archived asset cannot be updated", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			ProcessingStatus: domain.ProcessingStatusArchived,
		}
		assert.False(t, asset.CanUpdate())
	})

	t.Run("ready asset can be updated", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			ProcessingStatus: domain.ProcessingStatusReady,
		}
		assert.True(t, asset.CanUpdate())
	})

	t.Run("pending asset can be updated", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			ProcessingStatus: domain.ProcessingStatusPending,
		}
		assert.True(t, asset.CanUpdate())
	})
}
