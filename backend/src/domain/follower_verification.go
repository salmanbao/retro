package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// VerificationStatus represents the state of follower verification.
type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusPending    VerificationStatus = "pending"
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusRejected   VerificationStatus = "rejected"
)

// FollowerVerification tracks an Influencer's follower verification status and evidence.
type FollowerVerification struct {
	ProfileID         uuid.UUID          `gorm:"type:uuid;primary_key;not null" json:"profile_id"`
	Status            VerificationStatus `gorm:"type:varchar(50);not null;default:'unverified'" json:"status"`
	EvidenceURLs      json.RawMessage    `gorm:"type:jsonb" json:"evidence_urls,omitempty"`
	VerificationNotes string             `gorm:"type:text" json:"verification_notes,omitempty"`
	ReviewedAt        *time.Time         `gorm:"type:timestamptz" json:"reviewed_at,omitempty"`
	ReviewedBy        string             `gorm:"type:varchar(255)" json:"reviewed_by,omitempty"`
	CreatedAt         time.Time          `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time          `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for FollowerVerification.
func (FollowerVerification) TableName() string { return "follower_verifications" }

// GetEvidenceURLs parses evidence URLs from JSON.
func (f *FollowerVerification) GetEvidenceURLs() ([]string, error) {
	if f.EvidenceURLs == nil || len(f.EvidenceURLs) == 0 {
		return []string{}, nil
	}
	var result []string
	err := json.Unmarshal(f.EvidenceURLs, &result)
	return result, err
}

// SetEvidenceURLs serializes evidence URLs with size validation (max 50KB total).
func (f *FollowerVerification) SetEvidenceURLs(urls []string) error {
	data, err := json.Marshal(urls)
	if err != nil {
		return err
	}
	if len(data) > 50*1024 {
		return ErrEvidenceURLsTooLarge
	}
	f.EvidenceURLs = data
	f.UpdatedAt = time.Now()
	return nil
}

// SubmitEvidence updates status to pending and sets evidence.
func (f *FollowerVerification) SubmitEvidence(urls []string, notes string) error {
	if err := f.SetEvidenceURLs(urls); err != nil {
		return err
	}
	f.Status = VerificationStatusPending
	f.VerificationNotes = notes
	f.UpdatedAt = time.Now()
	return nil
}

// Review updates verification status after admin review.
func (f *FollowerVerification) Review(status VerificationStatus, reviewedBy string, notes string) {
	f.Status = status
	f.ReviewedBy = reviewedBy
	f.VerificationNotes = notes
	now := time.Now()
	f.ReviewedAt = &now
	f.UpdatedAt = now
}
