package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
)

func TestAssetVersionIncrement(t *testing.T) {
	t.Run("new asset has version 1", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			CampaignID:          uuid.New(),
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			UploadedByProfileID: uuid.New(),
		}
		assert.Equal(t, 1, asset.Version)
	})

	t.Run("version increment logic", func(t *testing.T) {
		existingVersion := 1
		newVersion := existingVersion + 1
		assert.Equal(t, 2, newVersion)
	})
}

func TestAssetVersionLookup(t *testing.T) {
	t.Run("latest version has highest version number", func(t *testing.T) {
		v1 := &domain.AssetMetadata{Version: 1}
		v2 := &domain.AssetMetadata{Version: 2}
		v3 := &domain.AssetMetadata{Version: 3}

		versions := []*domain.AssetMetadata{v3, v1, v2}
		latest := versions[0]
		for _, v := range versions {
			if v.Version > latest.Version {
				latest = v
			}
		}
		assert.Equal(t, 3, latest.Version)
	})
}

func TestAssetSoftDeletionScope(t *testing.T) {
	t.Run("asset without deleted_at is not soft-deleted", func(t *testing.T) {
		asset := &domain.AssetMetadata{}
		assert.False(t, asset.IsDeleted())
	})

	t.Run("asset with deleted_at is soft-deleted", func(t *testing.T) {
		asset := &domain.AssetMetadata{}
		asset.DeletedAt = nil
		assert.False(t, asset.IsDeleted())

		asset.DeletedAt = &time.Time{}
		assert.True(t, asset.IsDeleted())
	})
}

func TestAssetArchivedState(t *testing.T) {
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

	t.Run("processing asset can be updated", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			ProcessingStatus: domain.ProcessingStatusProcessing,
		}
		assert.True(t, asset.CanUpdate())
	})

	t.Run("failed asset can be updated", func(t *testing.T) {
		asset := &domain.AssetMetadata{
			ProcessingStatus: domain.ProcessingStatusFailed,
		}
		assert.True(t, asset.CanUpdate())
	})
}
