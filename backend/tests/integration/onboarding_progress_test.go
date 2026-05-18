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

func setupProgressTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=onboarding_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping integration test: database not available")
	}
	return db
}

func TestGetOrCreateProgress_CreatesNewProgress(t *testing.T) {
	db := setupProgressTestDB(t)
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

	// Seed templates
	onboardingSvc.SeedTemplates(templateRepo)

	svc := onboardingSvc.NewService(templateRepo, progressRepo, stepRepo)

	profileID := uuid.New()
	profileType := domain.ProfileTypeEditor

	// First call should create
	progress, err := svc.GetOrCreateProgress(profileID, profileType)
	if err != nil {
		t.Fatalf("GetOrCreateProgress failed: %v", err)
	}
	if progress.ProfileID != profileID {
		t.Errorf("ProfileID = %v, want %v", progress.ProfileID, profileID)
	}
	if progress.ActivationStatus != domain.ActivationStatusNotStarted {
		t.Errorf("ActivationStatus = %v, want not_started", progress.ActivationStatus)
	}

	// Second call should return existing
	progress2, err := svc.GetOrCreateProgress(profileID, profileType)
	if err != nil {
		t.Fatalf("GetOrCreateProgress failed: %v", err)
	}
	if progress2.ID != progress.ID {
		t.Errorf("Second call returned different ID")
	}
}

func TestUpdateStepStatus_ValidTransition(t *testing.T) {
	db := setupProgressTestDB(t)
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
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)

	// Get first step
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeEditor)
	firstStep := template.Steps[0]

	// Update step status
	updated, err := svc.UpdateStepStatus(progress.ID, firstStep.ID, domain.StepStatusInProgress)
	if err != nil {
		t.Fatalf("UpdateStepStatus failed: %v", err)
	}
	if updated.Status != domain.StepStatusInProgress {
		t.Errorf("Status = %v, want in_progress", updated.Status)
	}
}

func TestUpdateStepStatus_CannotSkipRequired(t *testing.T) {
	db := setupProgressTestDB(t)
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
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)

	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeEditor)
	requiredStep := template.Steps[0] // First step is required

	// Try to skip required step - should fail
	_, err := svc.UpdateStepStatus(progress.ID, requiredStep.ID, domain.StepStatusSkipped)
	if err != domain.ErrStepNotSkippable {
		t.Errorf("Expected ErrStepNotSkippable, got: %v", err)
	}
}

func TestRecalculateProgress_UpdatesActivationStatus(t *testing.T) {
	db := setupProgressTestDB(t)
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
	_, _ = svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)

	// Recalculate should update status to onboarding since steps exist
	recalculated, err := svc.RecalculateProgress(profileID)
	if err != nil {
		t.Fatalf("RecalculateProgress failed: %v", err)
	}
	if recalculated.ActivationStatus != domain.ActivationStatusOnboarding {
		t.Errorf("ActivationStatus = %v, want onboarding", recalculated.ActivationStatus)
	}
}