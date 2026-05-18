package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// TestValidateLanguageCodes tests ISO 639-1 language code validation
func TestValidateLanguageCodes(t *testing.T) {
	tests := []struct {
		name    string
		codes   []string
		wantErr bool
	}{
		{"valid single code", []string{"en"}, false},
		{"valid multiple codes", []string{"en", "es", "fr"}, false},
		{"valid codes with duplicates", []string{"en", "en", "es"}, false},
		{"empty list", []string{}, false},
		{"invalid too short", []string{"e"}, true},
		{"invalid too long", []string{"eng"}, true},
		{"invalid uppercase", []string{"EN"}, true},
		{"invalid with numbers", []string{"e1"}, true},
		{"invalid with special chars", []string{"e!"}, true},
		{"invalid mixed valid and invalid", []string{"en", "e"}, true},
		{"invalid empty string in list", []string{"en", ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateLanguageCodes(tt.codes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguageCodes(%v) error = %v, wantErr %v", tt.codes, err, tt.wantErr)
			}
		})
	}
}

// TestValidateTimezone tests IANA timezone identifier validation
func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{"valid America/New_York", "America/New_York", false},
		{"valid Europe/London", "Europe/London", false},
		{"valid Asia/Tokyo", "Asia/Tokyo", false},
		{"valid Etc/GMT+5", "Etc/GMT+5", false},
		{"valid with underscore", "America/Indiana/Petersburg", false},
		{"invalid UTC (needs slash)", "UTC", true}, // UTC doesn't have slash per IANA
		{"invalid no slash", "Eastern", true},
		{"invalid too short", "AM", true},
		{"invalid special chars", "America/New*York", true},
		{"invalid spaces", "America/New York", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateTimezone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimezone(%q) error = %v, wantErr %v", tt.tz, err, tt.wantErr)
			}
		})
	}
}

// TestValidateCountryCode tests ISO 3166-1 alpha-2 country code validation
func TestValidateCountryCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid US", "US", false},
		{"valid GB", "GB", false},
		{"valid JP", "JP", false},
		{"valid DE", "DE", false},
		{"invalid lowercase", "us", true},
		{"invalid too short", "U", true},
		{"invalid too long", "USA", true},
		{"invalid with numbers", "U1", true},
		{"invalid with special chars", "U!", true},
		{"invalid empty string", "", true},
		{"invalid mixed case", "Us", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateCountryCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCountryCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

// TestValidateCurrencyCode tests ISO 4217 currency code validation
func TestValidateCurrencyCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid USD", "USD", false},
		{"valid EUR", "EUR", false},
		{"valid GBP", "GBP", false},
		{"valid JPY", "JPY", false},
		{"invalid lowercase", "usd", true},
		{"invalid too short", "US", true},
		{"invalid too long", "USDD", true},
		{"invalid with numbers", "US1", true},
		{"invalid with special chars", "US!", true},
		{"invalid empty string", "", true},
		{"invalid mixed case", "Usd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateCurrencyCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrencyCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

// TestSocialLinksJSON tests SocialLinks serialization and deserialization
func TestSocialLinksJSON(t *testing.T) {
	tests := []struct {
		name    string
		links   *domain.SocialLinks
		wantErr bool
	}{
		{
			name:    "valid with all fields",
			links:   &domain.SocialLinks{TikTok: "user1", Instagram: "user1", YouTube: "channel1", XTwitter: "user1", LinkedIn: "user1", Website: "https://example.com"},
			wantErr: false,
		},
		{
			name:    "valid with minimal fields",
			links:   &domain.SocialLinks{TikTok: "user1"},
			wantErr: false,
		},
		{
			name:    "valid empty",
			links:   &domain.SocialLinks{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.links.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("SocialLinks.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				parsed, err := domain.SocialLinksFromJSON(data)
				if err != nil {
					t.Errorf("SocialLinksFromJSON() error = %v", err)
					return
				}
				if parsed.TikTok != tt.links.TikTok {
					t.Errorf("TikTok mismatch: got %v, want %v", parsed.TikTok, tt.links.TikTok)
				}
			}
		})
	}
}

// TestSocialLinksFromJSON tests deserialization edge cases
func TestSocialLinksFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"valid full", []byte(`{"tiktok":"user","instagram":"user","youtube":"channel"}`), false},
		{"valid empty object", []byte(`{}`), false},
		{"valid null", []byte(`null`), false},
		{"invalid json", []byte(`{invalid`), true},
		{"invalid type", []byte(`{"tiktok":123}`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.SocialLinksFromJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SocialLinksFromJSON(%q) error = %v, wantErr %v", string(tt.data), err, tt.wantErr)
			}
		})
	}
}

// TestProfileEnrichmentDomain tests ProfileEnrichment entity methods
func TestProfileEnrichmentDomain(t *testing.T) {
	profileID := uuid.New()
	now := time.Now()

	enrichment := &domain.ProfileEnrichment{
		ID:         uuid.New(),
		ProfileID:  profileID,
		Bio:        "Test bio",
		AvatarURL:  "https://example.com/avatar.jpg",
		CoverURL:   "https://example.com/cover.jpg",
		WebsiteURL: "https://example.com",
		Location:   "New York, NY",
		Languages:  []string{"en", "es"},
		Timezone:   "America/New_York",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Test GetSocialLinks method
	sl := &domain.SocialLinks{TikTok: "testuser", Instagram: "testuser"}
	err := enrichment.SetSocialLinks(sl)
	if err != nil {
		t.Errorf("SetSocialLinks() error = %v", err)
	}

	parsedSL, err := enrichment.GetSocialLinks()
	if err != nil {
		t.Errorf("GetSocialLinks() error = %v", err)
	}
	if parsedSL.TikTok != "testuser" {
		t.Errorf("GetSocialLinks().TikTok = %v, want %v", parsedSL.TikTok, "testuser")
	}

	// Test Update method
	err = enrichment.Update("New bio", "https://example.com/new-avatar.jpg", "https://example.com/new-cover.jpg", "https://example.com/new-site", "Los Angeles, CA", []string{"en"}, "America/Los_Angeles", sl)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if enrichment.Bio != "New bio" {
		t.Errorf("Bio not updated: got %v, want %v", enrichment.Bio, "New bio")
	}
	if enrichment.Timezone != "America/Los_Angeles" {
		t.Errorf("Timezone not updated: got %v, want %v", enrichment.Timezone, "America/Los_Angeles")
	}
}

// TestProfileEnrichmentValidation tests profile enrichment validation
func TestProfileEnrichmentValidation(t *testing.T) {
	profileID := uuid.New()
	now := time.Now()

	t.Run("valid enrichment", func(t *testing.T) {
		enrichment := &domain.ProfileEnrichment{
			ID:        uuid.New(),
			ProfileID: profileID,
			Languages: []string{"en", "es"},
			Timezone:  "America/New_York",
			CreatedAt: now,
			UpdatedAt: now,
		}
		// Validation happens via ValidateLanguageCodes and ValidateTimezone
		if err := domain.ValidateLanguageCodes(enrichment.Languages); err != nil {
			t.Errorf("ValidateLanguageCodes() unexpected error = %v", err)
		}
		if err := domain.ValidateTimezone(enrichment.Timezone); err != nil {
			t.Errorf("ValidateTimezone() unexpected error = %v", err)
		}
	})

	t.Run("invalid language code", func(t *testing.T) {
		enrichment := &domain.ProfileEnrichment{
			ID:        uuid.New(),
			ProfileID: profileID,
			Languages: []string{"EN"}, // uppercase not allowed
			Timezone:  "America/New_York",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := domain.ValidateLanguageCodes(enrichment.Languages); err == nil {
			t.Errorf("ValidateLanguageCodes() expected error for invalid language code")
		}
	})

	t.Run("invalid timezone", func(t *testing.T) {
		enrichment := &domain.ProfileEnrichment{
			ID:        uuid.New(),
			ProfileID: profileID,
			Timezone:  "Invalid", // missing slash
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := domain.ValidateTimezone(enrichment.Timezone); err == nil {
			t.Errorf("ValidateTimezone() expected error for invalid timezone")
		}
	})
}

// TestPortfolioItemDomain tests PortfolioItem entity methods
func TestPortfolioItemDomain(t *testing.T) {
	profileID := uuid.New()
	now := time.Now()

	t.Run("create and validate", func(t *testing.T) {
		item := &domain.PortfolioItem{
			ID:           uuid.New(),
			ProfileID:    profileID,
			Title:        "Test Portfolio Item",
			Description:  "A test portfolio item",
			ThumbnailURL: "https://example.com/thumb.jpg",
			VideoURL:     "https://example.com/video.mp4",
			ExternalLink: "https://example.com",
			DisplayOrder: 1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if item.ProfileID != profileID {
			t.Errorf("ProfileID mismatch")
		}
	})

	t.Run("soft delete", func(t *testing.T) {
		item := &domain.PortfolioItem{
			ID:        uuid.New(),
			ProfileID: profileID,
			Title:     "To be deleted",
			CreatedAt: now,
			UpdatedAt: now,
		}

		item.SoftDelete()
		if !item.IsDeleted() {
			t.Errorf("IsDeleted() = false, want true after SoftDelete()")
		}
		if item.DeletedAt == nil {
			t.Errorf("DeletedAt not set after SoftDelete()")
		}
	})

	t.Run("update portfolio item", func(t *testing.T) {
		item := &domain.PortfolioItem{
			ID:           uuid.New(),
			ProfileID:    profileID,
			DisplayOrder: 1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		item.Update("New Title", "New Description", "https://new.com/thumb.jpg", "https://new.com/video.mp4", "https://new.com", 5)
		if item.Title != "New Title" {
			t.Errorf("Title not updated")
		}
		if item.DisplayOrder != 5 {
			t.Errorf("DisplayOrder = %d, want 5", item.DisplayOrder)
		}
	})
}

// TestAudienceDataDomain tests AudienceData entity methods
func TestAudienceDataDomain(t *testing.T) {
	profileID := uuid.New()

	t.Run("create with valid data", func(t *testing.T) {
		data := &domain.AudienceData{
			ProfileID:      profileID,
			EngagementRate: 5.5,
		}
		if data.ProfileID != profileID {
			t.Errorf("ProfileID mismatch")
		}
	})

	t.Run("set and get platform handles", func(t *testing.T) {
		data := &domain.AudienceData{
			ProfileID:      profileID,
			EngagementRate: 3.0,
		}

		err := data.SetPlatformHandles(map[string]string{"tiktok": "user1", "instagram": "user1"})
		if err != nil {
			t.Errorf("SetPlatformHandles() error = %v", err)
		}

		handles, err := data.GetPlatformHandles()
		if err != nil {
			t.Errorf("GetPlatformHandles() error = %v", err)
		}
		if handles["tiktok"] != "user1" {
			t.Errorf("tiktok handle = %v, want %v", handles["tiktok"], "user1")
		}
	})

	t.Run("update audience data", func(t *testing.T) {
		data := &domain.AudienceData{
			ProfileID:      profileID,
			EngagementRate: 3.0,
		}

		err := data.Update(map[string]string{"tiktok": "newuser"}, map[string]int{"tiktok": 200000}, 4.5, nil)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
		if data.EngagementRate != 4.5 {
			t.Errorf("EngagementRate = %f, want 4.5", data.EngagementRate)
		}
	})
}

// TestFollowerVerificationDomain tests FollowerVerification entity methods
func TestFollowerVerificationDomain(t *testing.T) {
	profileID := uuid.New()

	t.Run("create verification", func(t *testing.T) {
		verification := &domain.FollowerVerification{
			ProfileID: profileID,
			Status:    domain.VerificationStatusUnverified,
		}

		if verification.Status != domain.VerificationStatusUnverified {
			t.Errorf("Status = %v, want %v", verification.Status, domain.VerificationStatusUnverified)
		}
	})

	t.Run("submit evidence", func(t *testing.T) {
		verification := &domain.FollowerVerification{
			ProfileID: profileID,
			Status:    domain.VerificationStatusUnverified,
		}

		err := verification.SubmitEvidence([]string{"https://example.com/proof1.jpg", "https://example.com/proof2.jpg"}, "Submitted for review")
		if err != nil {
			t.Errorf("SubmitEvidence() error = %v", err)
		}
		if verification.Status != domain.VerificationStatusPending {
			t.Errorf("Status = %v, want %v", verification.Status, domain.VerificationStatusPending)
		}

		evidence, err := verification.GetEvidenceURLs()
		if err != nil {
			t.Errorf("GetEvidenceURLs() error = %v", err)
		}
		if len(evidence) != 2 {
			t.Errorf("EvidenceURLs count = %d, want 2", len(evidence))
		}
	})

	t.Run("admin review - verify", func(t *testing.T) {
		verification := &domain.FollowerVerification{
			ProfileID: profileID,
			Status:    domain.VerificationStatusPending,
		}

		verification.Review(domain.VerificationStatusVerified, "admin123", "Approved by admin")
		if verification.Status != domain.VerificationStatusVerified {
			t.Errorf("Status = %v, want %v", verification.Status, domain.VerificationStatusVerified)
		}
		if verification.ReviewedBy != "admin123" {
			t.Errorf("ReviewedBy = %v, want %v", verification.ReviewedBy, "admin123")
		}
	})
}

// TestKYCStatusDomain tests KYCStatus entity methods
func TestKYCStatusDomain(t *testing.T) {
	profileID := uuid.New()

	t.Run("create KYC status", func(t *testing.T) {
		kyc := &domain.KYCStatus{
			ProfileID: profileID,
			Status:    domain.KYCStatusNotStarted,
		}

		if kyc.Status != domain.KYCStatusNotStarted {
			t.Errorf("Status = %v, want %v", kyc.Status, domain.KYCStatusNotStarted)
		}
	})

	t.Run("admin update KYC status", func(t *testing.T) {
		kyc := &domain.KYCStatus{
			ProfileID: profileID,
			Status:    domain.KYCStatusPending,
		}

		kyc.UpdateStatus(domain.KYCStatusApproved, "admin123", "Approved after review")
		if kyc.Status != domain.KYCStatusApproved {
			t.Errorf("Status = %v, want %v", kyc.Status, domain.KYCStatusApproved)
		}
		if kyc.ReviewedBy != "admin123" {
			t.Errorf("ReviewedBy = %v, want %v", kyc.ReviewedBy, "admin123")
		}
	})
}

// TestPayoutPreferencesDomain tests PayoutPreferences entity methods
func TestPayoutPreferencesDomain(t *testing.T) {
	profileID := uuid.New()

	t.Run("create payout preferences", func(t *testing.T) {
		prefs := &domain.PayoutPreferences{
			ProfileID: profileID,
		}

		if prefs.ProfileID != profileID {
			t.Errorf("ProfileID mismatch")
		}
	})

	t.Run("update payout preferences", func(t *testing.T) {
		prefs := &domain.PayoutPreferences{
			ProfileID:       profileID,
			PreferredMethod: domain.PayoutMethodBankTransfer,
		}

		prefs.Update(domain.PayoutMethodPayPal, "John Doe", "US", "USD", "encrypted_data", true)
		if prefs.PreferredMethod != domain.PayoutMethodPayPal {
			t.Errorf("PreferredMethod = %v, want %v", prefs.PreferredMethod, domain.PayoutMethodPayPal)
		}
		if prefs.BeneficiaryName != "John Doe" {
			t.Errorf("BeneficiaryName = %v, want %v", prefs.BeneficiaryName, "John Doe")
		}
		if prefs.PayoutReady != true {
			t.Errorf("PayoutReady = %v, want true", prefs.PayoutReady)
		}
	})

	t.Run("get masked returns no encrypted details", func(t *testing.T) {
		prefs := &domain.PayoutPreferences{
			ProfileID:        profileID,
			PreferredMethod:  domain.PayoutMethodBankTransfer,
			EncryptedDetails: "secret_encrypted_data",
		}

		masked := prefs.GetMasked()
		// The masked version should not contain the encrypted details
		// (GetMasked returns PayoutPreferencesMasked which doesn't have EncryptedDetails field)
		if masked.ProfileID != profileID {
			t.Errorf("ProfileID mismatch in masked")
		}
	})
}
