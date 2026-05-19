package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AssetCategory string

const (
	AssetCategoryRawFootage      AssetCategory = "raw_footage"
	AssetCategoryProductImages   AssetCategory = "product_images"
	AssetCategoryLogos           AssetCategory = "logos"
	AssetCategoryBrandGuidelines AssetCategory = "brand_guidelines"
	AssetCategoryScripts         AssetCategory = "scripts"
	AssetCategoryExampleVideos   AssetCategory = "example_videos"
	AssetCategoryMusicReferences AssetCategory = "music_references"
	AssetCategoryLegalDocuments  AssetCategory = "legal_documents"
	AssetCategoryOther           AssetCategory = "other"
)

func (c AssetCategory) IsValid() bool {
	switch c {
	case AssetCategoryRawFootage, AssetCategoryProductImages, AssetCategoryLogos,
		AssetCategoryBrandGuidelines, AssetCategoryScripts, AssetCategoryExampleVideos,
		AssetCategoryMusicReferences, AssetCategoryLegalDocuments, AssetCategoryOther:
		return true
	}
	return false
}

type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusUploaded   ProcessingStatus = "uploaded"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusReady      ProcessingStatus = "ready"
	ProcessingStatusFailed     ProcessingStatus = "failed"
	ProcessingStatusArchived   ProcessingStatus = "archived"
)

func (s ProcessingStatus) IsValid() bool {
	switch s {
	case ProcessingStatusPending, ProcessingStatusUploaded, ProcessingStatusProcessing,
		ProcessingStatusReady, ProcessingStatusFailed, ProcessingStatusArchived:
		return true
	}
	return false
}

type VirusScanStatus string

const (
	VirusScanStatusNotScanned VirusScanStatus = "not_scanned"
	VirusScanStatusScanning   VirusScanStatus = "scanning"
	VirusScanStatusClean      VirusScanStatus = "clean"
	VirusScanStatusInfected   VirusScanStatus = "infected"
)

func (s VirusScanStatus) IsValid() bool {
	switch s {
	case VirusScanStatusNotScanned, VirusScanStatusScanning, VirusScanStatusClean, VirusScanStatusInfected:
		return true
	}
	return false
}

type AssetMetadata struct {
	ID                  uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CampaignID          uuid.UUID        `json:"campaign_id" gorm:"type:uuid;not null;index"`
	Category            AssetCategory    `json:"category" gorm:"type:text;not null;index"`
	OriginalFilename    string           `json:"original_filename" gorm:"type:text;not null"`
	DisplayName         string           `json:"display_name" gorm:"type:text;not null"`
	MimeType            string           `json:"mime_type" gorm:"type:text;not null"`
	FileSizeBytes       int64            `json:"file_size_bytes" gorm:"not null"`
	StorageKey          string           `json:"storage_key" gorm:"type:text;not null"`
	Checksum            string           `json:"checksum" gorm:"type:text;not null"`
	Version             int              `json:"version" gorm:"not null;default:1"`
	ProcessingStatus    ProcessingStatus `json:"processing_status" gorm:"type:text;not null;default:'pending';index"`
	VirusScanStatus     VirusScanStatus  `json:"virus_scan_status" gorm:"type:text;not null;default:'not_scanned'"`
	UploadedByProfileID uuid.UUID        `json:"uploaded_by_profile_id" gorm:"type:uuid;not null;index"`
	CreatedAt           time.Time        `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt           time.Time        `json:"updated_at" gorm:"not null;default:now()"`
	DeletedAt           *time.Time       `json:"deleted_at,omitempty" gorm:"index"`
}

func (AssetMetadata) TableName() string {
	return "asset_metadata"
}

var (
	ErrAssetCampaignRequired        = errors.New("campaign_id is required")
	ErrAssetCategoryInvalid         = errors.New("invalid asset category")
	ErrAssetFilenameRequired        = errors.New("original_filename is required")
	ErrAssetDisplayNameRequired     = errors.New("display_name is required")
	ErrAssetMimeTypeRequired        = errors.New("mime_type is required")
	ErrAssetFileSizeInvalid         = errors.New("file_size_bytes must be positive")
	ErrAssetStorageKeyRequired      = errors.New("storage_key is required")
	ErrAssetChecksumRequired        = errors.New("checksum is required")
	ErrAssetChecksumLength          = errors.New("checksum must be 64 characters (SHA-256)")
	ErrAssetProfileIDRequired       = errors.New("uploaded_by_profile_id is required")
	ErrAssetVersionImmutable        = errors.New("version cannot be modified")
	ErrAssetStatusTransitionInvalid = errors.New("invalid processing status transition")
)

func (a *AssetMetadata) Validate() error {
	if a.CampaignID == uuid.Nil {
		return ErrAssetCampaignRequired
	}
	if !a.Category.IsValid() {
		return ErrAssetCategoryInvalid
	}
	if a.OriginalFilename == "" {
		return ErrAssetFilenameRequired
	}
	if a.DisplayName == "" {
		return ErrAssetDisplayNameRequired
	}
	if a.MimeType == "" {
		return ErrAssetMimeTypeRequired
	}
	if a.FileSizeBytes <= 0 {
		return ErrAssetFileSizeInvalid
	}
	if a.StorageKey == "" {
		return ErrAssetStorageKeyRequired
	}
	if a.Checksum == "" {
		return ErrAssetChecksumRequired
	}
	if len(a.Checksum) != 64 {
		return ErrAssetChecksumLength
	}
	if a.UploadedByProfileID == uuid.Nil {
		return ErrAssetProfileIDRequired
	}
	if !a.ProcessingStatus.IsValid() {
		return ErrAssetStatusTransitionInvalid
	}
	return nil
}

func (a *AssetMetadata) CanUpdate() bool {
	return a.ProcessingStatus != ProcessingStatusArchived
}

func (a *AssetMetadata) IsDeleted() bool {
	return a.DeletedAt != nil
}
