package service

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

type AssetService struct {
	assetRepo    repository.AssetMetadataRepository
	campaignRepo repository.CampaignRepository
}

func NewAssetService(
	assetRepo repository.AssetMetadataRepository,
	campaignRepo repository.CampaignRepository,
) *AssetService {
	return &AssetService{
		assetRepo:    assetRepo,
		campaignRepo: campaignRepo,
	}
}

type RegisterAssetInput struct {
	CampaignID          uuid.UUID
	Category            domain.AssetCategory
	OriginalFilename    string
	DisplayName         string
	MimeType            string
	FileSizeBytes       int64
	StorageKey          string
	Checksum            string
	UploadedByProfileID uuid.UUID
}

func (s *AssetService) Register(ctx context.Context, input RegisterAssetInput) (*domain.AssetMetadata, error) {
	latest, err := s.assetRepo.ByCampaignAndFilename(ctx, input.CampaignID, input.OriginalFilename)
	version := 1
	if err == nil && latest != nil {
		version = latest.Version + 1
	}

	asset := &domain.AssetMetadata{
		CampaignID:          input.CampaignID,
		Category:            input.Category,
		OriginalFilename:    input.OriginalFilename,
		DisplayName:         input.DisplayName,
		MimeType:            input.MimeType,
		FileSizeBytes:       input.FileSizeBytes,
		StorageKey:          input.StorageKey,
		Checksum:            input.Checksum,
		Version:             version,
		ProcessingStatus:    domain.ProcessingStatusPending,
		VirusScanStatus:     domain.VirusScanStatusNotScanned,
		UploadedByProfileID: input.UploadedByProfileID,
	}

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	if err := s.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

type ListAssetsInput struct {
	CampaignID uuid.UUID
	Page       int
	PageSize   int
}

func (s *AssetService) ListByCampaign(ctx context.Context, input ListAssetsInput) ([]*domain.AssetMetadata, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}
	return s.assetRepo.ListByCampaign(ctx, input.CampaignID, input.Page, input.PageSize)
}

func (s *AssetService) GetByID(ctx context.Context, id uuid.UUID) (*domain.AssetMetadata, error) {
	return s.assetRepo.ByID(ctx, id)
}

type UpdateAssetInput struct {
	DisplayName      *string
	ProcessingStatus *domain.ProcessingStatus
	VirusScanStatus  *domain.VirusScanStatus
}

func (s *AssetService) Update(ctx context.Context, asset *domain.AssetMetadata, input UpdateAssetInput) error {
	if !asset.CanUpdate() {
		return ErrAssetArchived
	}

	if input.DisplayName != nil {
		asset.DisplayName = *input.DisplayName
	}
	if input.ProcessingStatus != nil {
		asset.ProcessingStatus = *input.ProcessingStatus
	}
	if input.VirusScanStatus != nil {
		asset.VirusScanStatus = *input.VirusScanStatus
	}

	return s.assetRepo.Update(ctx, asset)
}

func (s *AssetService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return s.assetRepo.SoftDelete(ctx, id)
}

func (s *AssetService) GetLatestVersion(ctx context.Context, campaignID uuid.UUID, filename string) (*domain.AssetMetadata, error) {
	return s.assetRepo.ByCampaignAndFilename(ctx, campaignID, filename)
}

func (s *AssetService) ListVersions(ctx context.Context, campaignID uuid.UUID, filename string) ([]*domain.AssetMetadata, error) {
	return s.assetRepo.ListVersions(ctx, campaignID, filename)
}
