package onboarding

import (
	domain "viralforge/backend/src/domain/onboarding"
)

// AutoCompleteChecker checks if a step should be auto-completed based on profile data
type AutoCompleteChecker interface {
	Check(profileData map[string]interface{}) bool
	Key() string
}

// ProfileEnrichmentChecker checks if profile enrichment is complete (bio AND avatar)
type ProfileEnrichmentChecker struct{}

func (c *ProfileEnrichmentChecker) Key() string { return "profile_enrichment" }

func (c *ProfileEnrichmentChecker) Check(profileData map[string]interface{}) bool {
	// All-or-nothing: must have BOTH bio AND avatar
	bio, hasBio := profileData["bio"].(string)
	avatar, hasAvatar := profileData["avatar"].(string)
	return hasBio && len(bio) > 0 && hasAvatar && len(avatar) > 0
}

// PayoutPreferencesChecker checks if payout preferences are configured
type PayoutPreferencesChecker struct{}

func (c *PayoutPreferencesChecker) Key() string { return "payout_preferences" }

func (c *PayoutPreferencesChecker) Check(profileData map[string]interface{}) bool {
	// Must have payout preferences with encrypted details set
	payout, ok := profileData["payout_preferences"].(map[string]interface{})
	if !ok {
		return false
	}
	// Check if encrypted details are configured (not empty)
	_, hasEncrypted := payout["encrypted_details"]
	return hasEncrypted
}

// KYCStatusChecker checks if KYC is approved
type KYCStatusChecker struct{}

func (c *KYCStatusChecker) Key() string { return "kyc_status" }

func (c *KYCStatusChecker) Check(profileData map[string]interface{}) bool {
	// Must have KYC status that is approved
	kyc, ok := profileData["kyc_status"].(string)
	return ok && kyc == "approved"
}

// SocialLinksChecker checks if social links are configured
type SocialLinksChecker struct{}

func (c *SocialLinksChecker) Key() string { return "social_links" }

func (c *SocialLinksChecker) Check(profileData map[string]interface{}) bool {
	// Must have at least one social link configured
	socials, ok := profileData["social_links"].(map[string]interface{})
	if !ok {
		return false
	}
	return len(socials) > 0
}

// DefaultAutoCompleteCheckers returns the default set of auto-complete checkers
func DefaultAutoCompleteCheckers() []AutoCompleteChecker {
	return []AutoCompleteChecker{
		&ProfileEnrichmentChecker{},
		&PayoutPreferencesChecker{},
		&KYCStatusChecker{},
		&SocialLinksChecker{},
	}
}

// ApplyAutoCompletion applies auto-completion to all relevant steps
func ApplyAutoCompletion(profileData map[string]interface{}, steps []domain.StepProgress, templateSteps []domain.OnboardingStep) []domain.StepProgress {
	checkers := DefaultAutoCompleteCheckers()
	checkerMap := make(map[string]AutoCompleteChecker)
	for _, c := range checkers {
		checkerMap[c.Key()] = c
	}

	for i := range steps {
		step := &steps[i]
		// Skip already completed or skipped steps
		if step.Status == domain.StepStatusCompleted || step.Status == domain.StepStatusSkipped {
			continue
		}

		// Find the template step to get its auto_complete_key
		var autoCompleteKey string
		for _, ts := range templateSteps {
			if ts.ID == step.StepID {
				autoCompleteKey = ts.AutoCompleteKey
				break
			}
		}

		if autoCompleteKey == "" {
			continue
		}

		// Get the appropriate checker
		checker, exists := checkerMap[autoCompleteKey]
		if !exists {
			continue
		}

		// Check if auto-completion criteria are met
		if checker.Check(profileData) {
			step.Status = domain.StepStatusCompleted
		}
	}

	return steps
}
