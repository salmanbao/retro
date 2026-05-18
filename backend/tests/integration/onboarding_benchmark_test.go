package integration

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	domain "viralforge/backend/src/domain/onboarding"
	onboardingRepo "viralforge/backend/src/repository/onboarding"
	onboardingSvc "viralforge/backend/src/service/onboarding"
)

func getBenchmarkDB() *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=onboarding_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil
	}
	return db
}

func BenchmarkGetProgressByProfileID(b *testing.B) {
	db := getBenchmarkDB()
	if db == nil {
		b.Skip("Skipping benchmark: database not available")
	}
	defer func() {
		db.Exec("DROP TABLE IF EXISTS step_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_templates")
	}()

	db.AutoMigrate(&domain.OnboardingTemplate{}, &domain.OnboardingStep{})
	db.AutoMigrate(&domain.OnboardingProgress{}, &domain.StepProgress{})

	templateRepo := onboardingRepo.NewTemplateRepo(db)
	progressRepo := onboardingRepo.NewProgressRepo(db)
	stepRepo := onboardingRepo.NewStepRepo(db)

	onboardingSvc.SeedTemplates(templateRepo)

	svc := onboardingSvc.NewService(templateRepo, progressRepo, stepRepo)

	profileID := uuid.New()
	svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetProgressByProfileID(profileID)
	}
}

func BenchmarkGetNextStep(b *testing.B) {
	db := getBenchmarkDB()
	if db == nil {
		b.Skip("Skipping benchmark: database not available")
	}
	defer func() {
		db.Exec("DROP TABLE IF EXISTS step_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_templates")
	}()

	db.AutoMigrate(&domain.OnboardingTemplate{}, &domain.OnboardingStep{})
	db.AutoMigrate(&domain.OnboardingProgress{}, &domain.StepProgress{})

	templateRepo := onboardingRepo.NewTemplateRepo(db)
	progressRepo := onboardingRepo.NewProgressRepo(db)
	stepRepo := onboardingRepo.NewStepRepo(db)

	onboardingSvc.SeedTemplates(templateRepo)

	svc := onboardingSvc.NewService(templateRepo, progressRepo, stepRepo)

	profileID := uuid.New()
	svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetNextStep(profileID)
	}
}