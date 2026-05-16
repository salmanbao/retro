package domain

import (
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
