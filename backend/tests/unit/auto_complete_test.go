package unit

import (
	"testing"

	"viralforge/backend/src/service/onboarding"
)

func TestProfileEnrichmentChecker_AllOrNothing(t *testing.T) {
	checker := &onboarding.ProfileEnrichmentChecker{}

	tests := []struct {
		name        string
		profileData map[string]interface{}
		expect      bool
	}{
		{"both bio and avatar", map[string]interface{}{"bio": "Hello", "avatar": "https://example.com/photo.jpg"}, true},
		{"only bio", map[string]interface{}{"bio": "Hello"}, false},
		{"only avatar", map[string]interface{}{"avatar": "https://example.com/photo.jpg"}, false},
		{"empty bio", map[string]interface{}{"bio": "", "avatar": "https://example.com/photo.jpg"}, false},
		{"empty avatar", map[string]interface{}{"bio": "Hello", "avatar": ""}, false},
		{"neither", map[string]interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.Check(tt.profileData)
			if result != tt.expect {
				t.Errorf("ProfileEnrichmentChecker.Check() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestPayoutPreferencesChecker(t *testing.T) {
	checker := &onboarding.PayoutPreferencesChecker{}

	tests := []struct {
		name        string
		profileData map[string]interface{}
		expect      bool
	}{
		{"with encrypted details", map[string]interface{}{"payout_preferences": map[string]interface{}{"encrypted_details": "xxx"}}, true},
		{"without encrypted details", map[string]interface{}{"payout_preferences": map[string]interface{}{}}, false},
		{"empty payout", map[string]interface{}{"payout_preferences": map[string]interface{}{"details": ""}}, false},
		{"no payout data", map[string]interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.Check(tt.profileData)
			if result != tt.expect {
				t.Errorf("PayoutPreferencesChecker.Check() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestKYCStatusChecker(t *testing.T) {
	checker := &onboarding.KYCStatusChecker{}

	tests := []struct {
		name        string
		profileData map[string]interface{}
		expect      bool
	}{
		{"approved", map[string]interface{}{"kyc_status": "approved"}, true},
		{"pending", map[string]interface{}{"kyc_status": "pending"}, false},
		{"rejected", map[string]interface{}{"kyc_status": "rejected"}, false},
		{"empty", map[string]interface{}{"kyc_status": ""}, false},
		{"no kyc data", map[string]interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.Check(tt.profileData)
			if result != tt.expect {
				t.Errorf("KYCStatusChecker.Check() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestSocialLinksChecker(t *testing.T) {
	checker := &onboarding.SocialLinksChecker{}

	tests := []struct {
		name        string
		profileData map[string]interface{}
		expect      bool
	}{
		{"with social links", map[string]interface{}{"social_links": map[string]interface{}{"twitter": "https://twitter.com/user"}}, true},
		{"empty social links", map[string]interface{}{"social_links": map[string]interface{}{}}, false},
		{"no social links data", map[string]interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.Check(tt.profileData)
			if result != tt.expect {
				t.Errorf("SocialLinksChecker.Check() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestDefaultAutoCompleteCheckers(t *testing.T) {
	checkers := onboarding.DefaultAutoCompleteCheckers()
	if len(checkers) != 4 {
		t.Errorf("Expected 4 checkers, got %d", len(checkers))
	}

	keys := make(map[string]bool)
	for _, c := range checkers {
		keys[c.Key()] = true
	}

	expectedKeys := []string{"profile_enrichment", "payout_preferences", "kyc_status", "social_links"}
	for _, k := range expectedKeys {
		if !keys[k] {
			t.Errorf("Missing checker for key: %s", k)
		}
	}
}

func TestAutoCompleteKey(t *testing.T) {
	checker := &onboarding.ProfileEnrichmentChecker{}
	if checker.Key() != "profile_enrichment" {
		t.Errorf("ProfileEnrichmentChecker.Key() = %q, want profile_enrichment", checker.Key())
	}
}
