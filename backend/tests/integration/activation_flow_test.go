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

func setupActivationTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=onboarding_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping integration test: database not available")
	}
	return db
}

func TestActivationFlow_NotStartedToOnboarding(t *testing.T) {
	db := setupActivationTestDB(t)
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
	progress, err := svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)
	if err != nil {
		t.Fatalf("GetOrCreateProgress failed: %v", err)
	}

	if progress.ActivationStatus != domain.ActivationStatusNotStarted {
		t.Errorf("Initial status = %q, want not_started", progress.ActivationStatus)
	}

	// Complete first step - should transition to onboarding
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeEditor)
	firstStep := template.Steps[0]

	svc.UpdateStepStatus(progress.ID, firstStep.ID, domain.StepStatusInProgress)

	// Re-fetch and check activation status
	updatedProgress, _ := svc.GetProgressByProfileID(profileID)
	if updatedProgress.ActivationStatus != domain.ActivationStatusOnboarding {
		t.Errorf("After starting step, status = %q, want onboarding", updatedProgress.ActivationStatus)
	}
}

func TestActivationFlow_OnboardingToPendingReview(t *testing.T) {
	db := setupActivationTestDB(t)
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

	// Complete ALL required steps
	for _, step := range template.Steps {
		if step.Required {
			svc.UpdateStepStatus(progress.ID, step.ID, domain.StepStatusInProgress)
			svc.UpdateStepStatus(progress.ID, step.ID, domain.StepStatusCompleted)
		}
	}

	// Recalculate - should transition to pending_review
	recalculated, err := svc.RecalculateProgress(profileID)
	if err != nil {
		t.Fatalf("RecalculateProgress failed: %v", err)
	}

	if recalculated.ActivationStatus != domain.ActivationStatusPendingReview {
		t.Errorf("After all required complete, status = %q, want pending_review", recalculated.ActivationStatus)
	}
}

func TestActivationFlow_PendingReviewToActivated(t *testing.T) {
	db := setupActivationTestDB(t)
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
	activationSvc := onboardingSvc.NewActivationService(templateRepo, progressRepo, stepRepo)

	profileID := uuid.New()
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeEditor)
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeEditor)

	// Complete all required steps
	for _, step := range template.Steps {
		if step.Required {
			svc.UpdateStepStatus(progress.ID, step.ID, domain.StepStatusInProgress)
			svc.UpdateStepStatus(progress.ID, step.ID, domain.StepStatusCompleted)
		}
	}

	svc.RecalculateProgress(profileID)

	// Now activate via admin approval
	progress, _ = svc.GetProgressByProfileID(profileID)
	err := activationSvc.ActivateProfile(progress)
	if err != nil {
		t.Fatalf("ActivateProfile failed: %v", err)
	}

	// Re-fetch and check
	activatedProgress, _ := svc.GetProgressByProfileID(profileID)
	if activatedProgress.ActivationStatus != domain.ActivationStatusActivated {
		t.Errorf("After admin activation, status = %q, want activated", activatedProgress.ActivationStatus)
	}
}

func TestActivationFlow_RequiredStepSkippingBlocked(t *testing.T) {
	db := setupActivationTestDB(t)
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

	// Get a required step
	requiredStep := template.Steps[0] // First step is required

	// Start it first
	svc.UpdateStepStatus(progress.ID, requiredStep.ID, domain.StepStatusInProgress)

	// Try to skip it - should fail
	_, err := svc.UpdateStepStatus(progress.ID, requiredStep.ID, domain.StepStatusSkipped)
	if err != domain.ErrStepNotSkippable {
		t.Errorf("Expected ErrStepNotSkippable, got: %v", err)
	}
}

func TestActivationFlow_OptionalStepSkippingAllowed(t *testing.T) {
	db := setupActivationTestDB(t)
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
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeBrand)

	// Find an optional step (Brand template has step 4 as optional)
	var optionalStep domain.OnboardingStep
	for _, step := range template.Steps {
		if !step.Required {
			optionalStep = step
			break
		}
	}

	if optionalStep.ID == uuid.Nil {
		t.Skip("No optional step found in Brand template")
	}

	// Start then skip it - should succeed
	svc.UpdateStepStatus(progress.ID, optionalStep.ID, domain.StepStatusInProgress)
	updated, err := svc.UpdateStepStatus(progress.ID, optionalStep.ID, domain.StepStatusSkipped)
	if err != nil {
		t.Errorf("Skipping optional step failed: %v", err)
	}
	if updated.Status != domain.StepStatusSkipped {
		t.Errorf("Status = %q, want skipped", updated.Status)
	}
}