package integration

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	domain "viralforge/backend/src/domain/onboarding"
	onboardingRepo "viralforge/backend/src/repository/onboarding"
	onboardingSvc "viralforge/backend/src/service/onboarding"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=onboarding_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping integration test: database not available")
	}
	return db
}

func TestSeedTemplates_CreatesAllProfileTypes(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		db.Exec("DROP TABLE IF EXISTS step_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_steps")
		db.Exec("DROP TABLE IF EXISTS onboarding_templates")
	}()

	// Migrate
	db.AutoMigrate(&domain.OnboardingTemplate{}, &domain.OnboardingStep{})
	db.AutoMigrate(&domain.OnboardingProgress{}, &domain.StepProgress{})

	repo := onboardingRepo.NewTemplateRepo(db)
	seedErr := onboardingSvc.SeedTemplates(repo)
	if seedErr != nil {
		t.Fatalf("SeedTemplates failed: %v", seedErr)
	}

	// Verify Brand template
	brandTemplate, err := repo.GetByProfileType(domain.ProfileTypeBrand)
	if err != nil {
		t.Fatalf("Failed to get Brand template: %v", err)
	}
	if brandTemplate.ProfileType != domain.ProfileTypeBrand {
		t.Errorf("Brand template profile type = %v, want %v", brandTemplate.ProfileType, domain.ProfileTypeBrand)
	}
	if len(brandTemplate.Steps) != 4 {
		t.Errorf("Brand template step count = %v, want 4", len(brandTemplate.Steps))
	}

	// Verify Editor template
	editorTemplate, err := repo.GetByProfileType(domain.ProfileTypeEditor)
	if err != nil {
		t.Fatalf("Failed to get Editor template: %v", err)
	}
	if editorTemplate.ProfileType != domain.ProfileTypeEditor {
		t.Errorf("Editor template profile type = %v, want %v", editorTemplate.ProfileType, domain.ProfileTypeEditor)
	}
	if len(editorTemplate.Steps) != 4 {
		t.Errorf("Editor template step count = %v, want 4", len(editorTemplate.Steps))
	}

	// Verify Influencer template
	influencerTemplate, err := repo.GetByProfileType(domain.ProfileTypeInfluencer)
	if err != nil {
		t.Fatalf("Failed to get Influencer template: %v", err)
	}
	if influencerTemplate.ProfileType != domain.ProfileTypeInfluencer {
		t.Errorf("Influencer template profile type = %v, want %v", influencerTemplate.ProfileType, domain.ProfileTypeInfluencer)
	}
	if len(influencerTemplate.Steps) != 5 {
		t.Errorf("Influencer template step count = %v, want 5", len(influencerTemplate.Steps))
	}
}

func TestGetTemplateByProfileType_EditorTemplate_Has4Steps(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		db.Exec("DROP TABLE IF EXISTS step_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_templates")
	}()

	db.AutoMigrate(&domain.OnboardingTemplate{}, &domain.OnboardingStep{})

	repo := onboardingRepo.NewTemplateRepo(db)
	onboardingSvc.SeedTemplates(repo)

	template, err := repo.GetByProfileType(domain.ProfileTypeEditor)
	if err != nil {
		t.Fatalf("GetByProfileType failed: %v", err)
	}

	expectedSteps := []struct {
		title  string
		required bool
	}{
		{"Complete public profile", true},
		{"Upload portfolio items", true},
		{"Add payout preferences", true},
		{"Complete KYC", true},
	}

	if len(template.Steps) != len(expectedSteps) {
		t.Fatalf("Editor template step count = %v, want %v", len(template.Steps), len(expectedSteps))
	}

	for i, expected := range expectedSteps {
		step := template.Steps[i]
		if step.Title != expected.title {
			t.Errorf("Step[%d].Title = %v, want %v", i, step.Title, expected.title)
		}
		if step.Required != expected.required {
			t.Errorf("Step[%d].Required = %v, want %v", i, step.Required, expected.required)
		}
	}
}

func TestGetTemplateByID(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		db.Exec("DROP TABLE IF EXISTS step_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_templates")
	}()

	db.AutoMigrate(&domain.OnboardingTemplate{}, &domain.OnboardingStep{})

	repo := onboardingRepo.NewTemplateRepo(db)
	onboardingSvc.SeedTemplates(repo)

	// Get the brand template by profile type first
	brandTemplate, _ := repo.GetByProfileType(domain.ProfileTypeBrand)

	// Then retrieve by ID
	retrieved, err := repo.GetByID(brandTemplate.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.ID != brandTemplate.ID {
		t.Errorf("GetByID returned wrong template")
	}
}

func TestListTemplates_ReturnsAll3(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		db.Exec("DROP TABLE IF EXISTS step_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_progresses")
		db.Exec("DROP TABLE IF EXISTS onboarding_templates")
	}()

	db.AutoMigrate(&domain.OnboardingTemplate{}, &domain.OnboardingStep{})

	repo := onboardingRepo.NewTemplateRepo(db)
	onboardingSvc.SeedTemplates(repo)

	templates, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(templates) != 3 {
		t.Errorf("List returned %v templates, want 3", len(templates))
	}
}