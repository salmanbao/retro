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

func setupAutoCompleteTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=onboarding_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping integration test: database not available")
	}
	return db
}

func TestAutoComplete_ProfileEnrichmentComplete(t *testing.T) {
	db := setupAutoCompleteTestDB(t)
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
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeInfluencer)
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeInfluencer)

	// Find the profile_completion step
	var profileStep domain.OnboardingStep
	for _, step := range template.Steps {
		if step.AutoCompleteKey == "profile_enrichment" {
			profileStep = step
			break
		}
	}

	// Simulate profile data with bio AND avatar (all-or-nothing)
	profileData := map[string]interface{}{
		"bio":     "This is my bio",
		"avatar":  "https://example.com/avatar.jpg",
	}

	// Get current step progress
	steps, _ := stepRepo.GetByOnboardingProgressID(progress.ID)

	// Apply auto-completion
	updatedSteps := onboardingSvc.ApplyAutoCompletion(profileData, steps, template.Steps)

	// Find the profile_completion step progress
	var profileStepProgress *domain.StepProgress
	for i := range updatedSteps {
		if updatedSteps[i].StepID == profileStep.ID {
			profileStepProgress = &updatedSteps[i]
			break
		}
	}

	if profileStepProgress == nil {
		t.Fatal("Profile step progress not found")
	}

	if profileStepProgress.Status != domain.StepStatusCompleted {
		t.Errorf("Profile step status = %q, want completed", profileStepProgress.Status)
	}
}

func TestAutoComplete_ProfileEnrichmentPartialNoAutoComplete(t *testing.T) {
	db := setupAutoCompleteTestDB(t)
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
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeInfluencer)
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeInfluencer)

	// Find the profile_completion step
	var profileStep domain.OnboardingStep
	for _, step := range template.Steps {
		if step.AutoCompleteKey == "profile_enrichment" {
			profileStep = step
			break
		}
	}

	// Simulate profile data with ONLY bio (not enough - all-or-nothing)
	profileData := map[string]interface{}{
		"bio": "This is my bio",
		// No avatar
	}

	steps, _ := stepRepo.GetByOnboardingProgressID(progress.ID)
	updatedSteps := onboardingSvc.ApplyAutoCompletion(profileData, steps, template.Steps)

	var profileStepProgress *domain.StepProgress
	for i := range updatedSteps {
		if updatedSteps[i].StepID == profileStep.ID {
			profileStepProgress = &updatedSteps[i]
			break
		}
	}

	if profileStepProgress == nil {
		t.Fatal("Profile step progress not found")
	}

	// Should NOT auto-complete because only bio is present, not both bio AND avatar
	if profileStepProgress.Status == domain.StepStatusCompleted {
		t.Errorf("Profile step should NOT be auto-completed with only bio data")
	}
}

func TestAutoComplete_KYCApproved(t *testing.T) {
	db := setupAutoCompleteTestDB(t)
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
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeInfluencer)
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeInfluencer)

	// Find the kyc step
	var kycStep domain.OnboardingStep
	for _, step := range template.Steps {
		if step.AutoCompleteKey == "kyc_status" {
			kycStep = step
			break
		}
	}

	// Simulate profile data with approved KYC
	profileData := map[string]interface{}{
		"kyc_status": "approved",
	}

	steps, _ := stepRepo.GetByOnboardingProgressID(progress.ID)
	updatedSteps := onboardingSvc.ApplyAutoCompletion(profileData, steps, template.Steps)

	var kycStepProgress *domain.StepProgress
	for i := range updatedSteps {
		if updatedSteps[i].StepID == kycStep.ID {
			kycStepProgress = &updatedSteps[i]
			break
		}
	}

	if kycStepProgress == nil {
		t.Fatal("KYC step progress not found")
	}

	if kycStepProgress.Status != domain.StepStatusCompleted {
		t.Errorf("KYC step status = %q, want completed", kycStepProgress.Status)
	}
}

func TestAutoComplete_SocialLinks(t *testing.T) {
	db := setupAutoCompleteTestDB(t)
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
	progress, _ := svc.GetOrCreateProgress(profileID, domain.ProfileTypeInfluencer)
	template, _ := svc.GetTemplateByProfileType(domain.ProfileTypeInfluencer)

	// Find the social links step
	var socialStep domain.OnboardingStep
	for _, step := range template.Steps {
		if step.AutoCompleteKey == "social_links" {
			socialStep = step
			break
		}
	}

	// Simulate profile data with social links
	profileData := map[string]interface{}{
		"social_links": map[string]interface{}{
			"twitter": "https://twitter.com/user",
			"instagram": "https://instagram.com/user",
		},
	}

	steps, _ := stepRepo.GetByOnboardingProgressID(progress.ID)
	updatedSteps := onboardingSvc.ApplyAutoCompletion(profileData, steps, template.Steps)

	var socialStepProgress *domain.StepProgress
	for i := range updatedSteps {
		if updatedSteps[i].StepID == socialStep.ID {
			socialStepProgress = &updatedSteps[i]
			break
		}
	}

	if socialStepProgress == nil {
		t.Fatal("Social links step progress not found")
	}

	if socialStepProgress.Status != domain.StepStatusCompleted {
		t.Errorf("Social links step status = %q, want completed", socialStepProgress.Status)
	}
}

func TestGetNextStep_ReturnsFirstIncomplete(t *testing.T) {
	db := setupAutoCompleteTestDB(t)
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

	// Get next step - should be first step since all are not_started
	nextStep, err := svc.GetNextStep(profileID)
	if err != nil {
		t.Fatalf("GetNextStep failed: %v", err)
	}
	if nextStep == nil {
		t.Fatal("GetNextStep returned nil, expected first step")
	}
	if nextStep.DisplayOrder != 1 {
		t.Errorf("First step DisplayOrder = %d, want 1", nextStep.DisplayOrder)
	}
}

func TestGetNextStep_ReturnsNullWhenAllComplete(t *testing.T) {
	db := setupAutoCompleteTestDB(t)
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

	// Complete all steps
	for _, step := range template.Steps {
		svc.UpdateStepStatus(progress.ID, step.ID, domain.StepStatusInProgress)
		svc.UpdateStepStatus(progress.ID, step.ID, domain.StepStatusCompleted)
	}

	// Get next step - should be nil since all are complete
	nextStep, err := svc.GetNextStep(profileID)
	if err != nil {
		t.Fatalf("GetNextStep failed: %v", err)
	}
	if nextStep != nil {
		t.Errorf("GetNextStep returned %v, want nil (all complete)", nextStep)
	}
}