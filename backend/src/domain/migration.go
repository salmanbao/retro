package domain

import (
	"viralforge/backend/src/domain/onboarding"

	"gorm.io/gorm"
)

// MigrateProfileEnrichmentTables runs GORM auto-migrate for all profile enrichment entities.
func MigrateProfileEnrichmentTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&ProfileEnrichment{},
		&PortfolioItem{},
		&AudienceData{},
		&FollowerVerification{},
		&PayoutPreferences{},
		&KYCStatus{},
	)
}

// MigrateOnboardingTables runs GORM auto-migrate for all onboarding entities.
func MigrateOnboardingTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&onboarding.OnboardingTemplate{},
		&onboarding.OnboardingStep{},
		&onboarding.OnboardingProgress{},
		&onboarding.StepProgress{},
	)
}
