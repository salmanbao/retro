package domain

import (
	"time"

	"github.com/google/uuid"
)

// KYCStatusValue represents the KYC verification status.
type KYCStatusValue string

const (
	KYCStatusNotStarted KYCStatusValue = "not_started"
	KYCStatusPending    KYCStatusValue = "pending"
	KYCStatusApproved   KYCStatusValue = "approved"
	KYCStatusRejected   KYCStatusValue = "rejected"
	KYCStatusSuspended  KYCStatusValue = "suspended"
)

// KYCStatus represents a profile's KYC verification state.
type KYCStatus struct {
	ProfileID   uuid.UUID      `gorm:"type:uuid;primary_key;not null" json:"profile_id"`
	Status      KYCStatusValue `gorm:"type:varchar(50);not null;default:'not_started'" json:"status"`
	ReviewNotes string         `gorm:"type:text" json:"review_notes,omitempty"`
	ReviewedAt  *time.Time     `gorm:"type:timestamptz" json:"reviewed_at,omitempty"`
	ReviewedBy  string         `gorm:"type:varchar(255)" json:"reviewed_by,omitempty"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for KYCStatus.
func (KYCStatus) TableName() string { return "kyc_status" }

// UpdateStatus updates KYC status after admin review.
func (k *KYCStatus) UpdateStatus(status KYCStatusValue, reviewedBy string, notes string) {
	k.Status = status
	k.ReviewedBy = reviewedBy
	k.ReviewNotes = notes
	now := time.Now()
	k.ReviewedAt = &now
	k.UpdatedAt = now
}
