package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

type mockAssetRepo struct {
	assets []*domain.AssetMetadata
}

func (m *mockAssetRepo) Create(ctx interface{}, asset *domain.AssetMetadata) error {
	asset.ID = uuid.New()
	asset.CreatedAt = now()
	asset.UpdatedAt = now()
	m.assets = append(m.assets, asset)
	return nil
}

func (m *mockAssetRepo) ByID(ctx interface{}, id uuid.UUID) (*domain.AssetMetadata, error) {
	for _, a := range m.assets {
		if a.ID == id && a.DeletedAt == nil {
			return a, nil
		}
	}
	return nil, repository.ErrAssetNotFound
}

func (m *mockAssetRepo) ListByCampaign(ctx interface{}, campaignID uuid.UUID, page, pageSize int) ([]*domain.AssetMetadata, int64, error) {
	var result []*domain.AssetMetadata
	for _, a := range m.assets {
		if a.CampaignID == campaignID && a.DeletedAt == nil {
			result = append(result, a)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockAssetRepo) Update(ctx interface{}, asset *domain.AssetMetadata) error {
	for i, a := range m.assets {
		if a.ID == asset.ID {
			asset.UpdatedAt = now()
			m.assets[i] = asset
			return nil
		}
	}
	return repository.ErrAssetNotFound
}

func (m *mockAssetRepo) SoftDelete(ctx interface{}, id uuid.UUID) error {
	for i, a := range m.assets {
		if a.ID == id {
			now := now()
			m.assets[i].DeletedAt = &now
			return nil
		}
	}
	return repository.ErrAssetNotFound
}

func (m *mockAssetRepo) ByCampaignAndFilename(ctx interface{}, campaignID uuid.UUID, filename string) (*domain.AssetMetadata, error) {
	var latest *domain.AssetMetadata
	for _, a := range m.assets {
		if a.CampaignID == campaignID && a.OriginalFilename == filename && a.DeletedAt == nil {
			if latest == nil || a.Version > latest.Version {
				latest = a
			}
		}
	}
	if latest == nil {
		return nil, repository.ErrAssetNotFound
	}
	return latest, nil
}

func (m *mockAssetRepo) ListVersions(ctx interface{}, campaignID uuid.UUID, filename string) ([]*domain.AssetMetadata, error) {
	var result []*domain.AssetMetadata
	for _, a := range m.assets {
		if a.CampaignID == campaignID && a.OriginalFilename == filename {
			result = append(result, a)
		}
	}
	return result, nil
}

func TestAssetCRUD(t *testing.T) {
	repo := &mockAssetRepo{assets: make([]*domain.AssetMetadata, 0)}

	t.Run("create asset", func(t *testing.T) {
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
			ProcessingStatus:    domain.ProcessingStatusPending,
			VirusScanStatus:     domain.VirusScanStatusNotScanned,
			UploadedByProfileID: uuid.New(),
		}

		err := repo.Create(nil, asset)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, asset.ID)
	})

	t.Run("read asset by ID", func(t *testing.T) {
		campaignID := uuid.New()
		asset := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			ProcessingStatus:    domain.ProcessingStatusPending,
			UploadedByProfileID: uuid.New(),
		}
		repo.Create(nil, asset)

		found, err := repo.ByID(nil, asset.ID)
		assert.NoError(t, err)
		assert.Equal(t, asset.ID, found.ID)
	})

	t.Run("list assets by campaign", func(t *testing.T) {
		campaignID := uuid.New()

		asset1 := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video1.mp4",
			DisplayName:         "Test Video 1",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video1.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			ProcessingStatus:    domain.ProcessingStatusPending,
			UploadedByProfileID: uuid.New(),
		}
		asset2 := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryProductImages,
			OriginalFilename:    "image1.png",
			DisplayName:         "Test Image",
			MimeType:            "image/png",
			FileSizeBytes:       2048,
			StorageKey:          "campaigns/abc/image1.png",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			ProcessingStatus:    domain.ProcessingStatusReady,
			UploadedByProfileID: uuid.New(),
		}

		repo.Create(nil, asset1)
		repo.Create(nil, asset2)

		assets, total, err := repo.ListByCampaign(nil, campaignID, 1, 20)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, assets, 2)
	})
}

func TestAssetVersioning(t *testing.T) {
	repo := &mockAssetRepo{assets: make([]*domain.AssetMetadata, 0)}

	t.Run("version increment on same filename", func(t *testing.T) {
		campaignID := uuid.New()

		v1 := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Video v1",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video_v1.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			ProcessingStatus:    domain.ProcessingStatusPending,
			UploadedByProfileID: uuid.New(),
		}
		repo.Create(nil, v1)

		latest, _ := repo.ByCampaignAndFilename(nil, campaignID, "video.mp4")
		assert.Equal(t, 1, latest.Version)

		v2 := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Video v2",
			MimeType:            "video/mp4",
			FileSizeBytes:       2048,
			StorageKey:          "campaigns/abc/video_v2.mp4",
			Checksum:            "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Version:             2,
			ProcessingStatus:    domain.ProcessingStatusPending,
			UploadedByProfileID: uuid.New(),
		}
		repo.Create(nil, v2)

		latest, _ = repo.ByCampaignAndFilename(nil, campaignID, "video.mp4")
		assert.Equal(t, 2, latest.Version)
	})

	t.Run("list all versions", func(t *testing.T) {
		campaignID := uuid.New()
		filename := "video.mp4"

		v1 := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    filename,
			DisplayName:         "Video v1",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video_v1.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			ProcessingStatus:    domain.ProcessingStatusReady,
			UploadedByProfileID: uuid.New(),
		}
		v2 := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    filename,
			DisplayName:         "Video v2",
			MimeType:            "video/mp4",
			FileSizeBytes:       2048,
			StorageKey:          "campaigns/abc/video_v2.mp4",
			Checksum:            "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Version:             2,
			ProcessingStatus:    domain.ProcessingStatusReady,
			UploadedByProfileID: uuid.New(),
		}

		repo.Create(nil, v1)
		repo.Create(nil, v2)

		versions, err := repo.ListVersions(nil, campaignID, filename)
		assert.NoError(t, err)
		assert.Len(t, versions, 2)
	})
}

func TestAssetSoftDeletion(t *testing.T) {
	repo := &mockAssetRepo{assets: make([]*domain.AssetMetadata, 0)}

	t.Run("soft delete excludes from list", func(t *testing.T) {
		campaignID := uuid.New()

		asset := &domain.AssetMetadata{
			CampaignID:          campaignID,
			Category:            domain.AssetCategoryRawFootage,
			OriginalFilename:    "video.mp4",
			DisplayName:         "Test Video",
			MimeType:            "video/mp4",
			FileSizeBytes:       1024,
			StorageKey:          "campaigns/abc/video.mp4",
			Checksum:            "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Version:             1,
			ProcessingStatus:    domain.ProcessingStatusPending,
			UploadedByProfileID: uuid.New(),
		}
		repo.Create(nil, asset)

		assets, _, _ := repo.ListByCampaign(nil, campaignID, 1, 20)
		assert.Len(t, assets, 1)

		repo.SoftDelete(nil, asset.ID)

		assets, _, _ = repo.ListByCampaign(nil, campaignID, 1, 20)
		assert.Len(t, assets, 0)
	})

	t.Run("soft deleted asset not found by ID", func(t *testing.T) {
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
			ProcessingStatus:    domain.ProcessingStatusPending,
			UploadedByProfileID: uuid.New(),
		}
		repo.Create(nil, asset)
		repo.SoftDelete(nil, asset.ID)

		_, err := repo.ByID(nil, asset.ID)
		assert.ErrorIs(t, err, repository.ErrAssetNotFound)
	})
}

func now() time.Time {
	return time.Now()
}
