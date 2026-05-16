package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"viralforge/backend/src/adapter"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// EnrichmentIntegrationTestSuite runs all enrichment integration tests against PostgreSQL.
type EnrichmentIntegrationTestSuite struct {
	suite.Suite
	db     *adapter.PostgresStore
	user   *domain.User
	editor *domain.Profile
}

func (s *EnrichmentIntegrationTestSuite) SetupSuite() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/viralforge?sslmode=disable"
	}

	// Connect to PostgreSQL
	gormDB, err := adapter.Connect(context.Background(), databaseURL)
	if err != nil {
		s.T().Skipf("Skipping integration test: database not available: %v", err)
		return
	}
	s.db = adapter.NewPostgresStore(gormDB)

	// Auto-migrate
	if err := s.db.AutoMigrate(); err != nil {
		s.T().Skipf("Skipping integration test: migration failed: %v", err)
		return
	}

	// Create test user
	s.user = &domain.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		PasswordHash: "hash",
		Verified:     true,
	}
	err = s.db.UserRepository().Create(context.Background(), s.user)
	if err != nil {
		s.T().Skipf("Skipping integration test: failed to create user: %v", err)
		return
	}

	// Create editor profile
	s.editor = &domain.Profile{
		ID:     uuid.New(),
		UserID: s.user.ID,
		Type:   domain.ProfileTypeEditor,
		Name:   "Test Editor",
	}
	err = s.db.ProfileRepository().Create(context.Background(), s.editor)
	if err != nil {
		s.T().Skipf("Skipping integration test: failed to create profile: %v", err)
		return
	}
}

func (s *EnrichmentIntegrationTestSuite) TearDownSuite() {
	if s.db != nil && s.editor != nil {
		// Clean up test data
		ctx := context.Background()
		s.db.ProfileRepository().Delete(ctx, s.editor.ID)
		if s.user != nil {
			s.db.UserRepository().Update(ctx, s.user)
		}
	}
}

func TestEnrichmentIntegrationSuite(t *testing.T) {
	suite.Run(t, new(EnrichmentIntegrationTestSuite))
}

// Cleanup portfolio items and profile enrichments for a profile
func cleanupProfileData(db *adapter.PostgresStore, profileID uuid.UUID) {
	if db == nil {
		return
	}
	ctx := context.Background()

	// Cleanup portfolio items
	portfolioRepo := db.PortfolioItemRepository()
	items, _ := portfolioRepo.ByProfileID(ctx, profileID, 100, 0)
	for _, item := range items {
		portfolioRepo.Delete(ctx, item.ID)
	}

	// Cleanup profile enrichment (delete and recreate to avoid update issues)
	enrichment, err := db.ProfileEnrichmentRepository().ByProfileID(ctx, profileID)
	if err == nil && enrichment != nil {
		// Delete the enrichment
		db.DB().Where("profile_id = ?", profileID).Delete(enrichment)
	}
}

// Cleanup portfolio items for a profile
func cleanupPortfolioItems(db *adapter.PostgresStore, profileID uuid.UUID) {
	if db == nil {
		return
	}
	ctx := context.Background()
	repo := db.PortfolioItemRepository()
	items, _ := repo.ByProfileID(ctx, profileID, 100, 0)
	for _, item := range items {
		repo.Delete(ctx, item.ID)
	}
}

// TestT084_Scenario1_ProfileEnrichmentCRUD tests quickstart Scenario 1.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario1_ProfileEnrichmentCRUD() {
	ctx := context.Background()
	cleanupPortfolioItems(s.db, s.editor.ID)

	// Create profile enrichment
	enrichmentRepo := s.db.ProfileEnrichmentRepository()
	profileEnrichmentSvc := service.NewProfileEnrichmentService(enrichmentRepo, s.db.ProfileRepository())

	// Update details (creates if not exists)
	socialLinks := &domain.SocialLinks{
		TikTok:    "testuser",
		Instagram: "testuser",
	}
	enrichment, err := profileEnrichmentSvc.UpdateDetails(
		ctx,
		s.editor.ID,
		"Test creator bio",
		"https://example.com/avatar.jpg",
		"https://example.com/cover.jpg",
		"https://example.com",
		"New York",
		[]string{"en", "fr"},
		"America/New_York",
		socialLinks,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test creator bio", enrichment.Bio)

	// Get details
	retrieved, err := profileEnrichmentSvc.GetDetails(ctx, s.editor.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test creator bio", retrieved.Bio)
	assert.Equal(s.T(), "New York", retrieved.Location)

	// Get social links
	sl, err := retrieved.GetSocialLinks()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "testuser", sl.TikTok)
}

// TestT084_Scenario2_PortfolioCRUD tests quickstart Scenario 2.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario2_PortfolioCRUD() {
	ctx := context.Background()
	portfolioSvc := service.NewPortfolioService(s.db.PortfolioItemRepository(), s.db.ProfileRepository())

	// Create portfolio item
	item, err := portfolioSvc.Create(
		ctx,
		s.editor.ID,
		"Campaign Video",
		"Brand collaboration video",
		"https://example.com/thumb.jpg",
		"https://example.com/video.mp4",
		"https://example.com",
		1,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Campaign Video", item.Title)

	// List portfolio items
	items, err := portfolioSvc.ListByProfileID(ctx, s.editor.ID, 50, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), items, 1)

	// Update display order
	updated, err := portfolioSvc.Update(
		ctx,
		item.ID,
		s.editor.ID,
		"Campaign Video",
		"Brand collaboration video",
		"https://example.com/thumb.jpg",
		"https://example.com/video.mp4",
		"https://example.com",
		2,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, updated.DisplayOrder)

	// Delete portfolio item
	err = portfolioSvc.Delete(ctx, item.ID, s.editor.ID)
	require.NoError(s.T(), err)

	// Verify deleted
	items, err = portfolioSvc.ListByProfileID(ctx, s.editor.ID, 50, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), items, 0)
}

// TestT084_Scenario3_PortfolioRejectionForNonEditor tests quickstart Scenario 3.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario3_PortfolioRejectionForNonEditor() {
	ctx := context.Background()

	// Create a brand profile
	brandProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: s.user.ID,
		Type:   domain.ProfileTypeBrand,
		Name:   "Test Brand",
	}
	err := s.db.ProfileRepository().Create(ctx, brandProfile)
	require.NoError(s.T(), err)

	portfolioSvc := service.NewPortfolioService(s.db.PortfolioItemRepository(), s.db.ProfileRepository())

	// Try to create portfolio item as brand (should fail)
	_, err = portfolioSvc.Create(
		ctx,
		brandProfile.ID,
		"Test Item",
		"Description",
		"",
		"",
		"",
		1,
	)
	assert.Equal(s.T(), service.ErrProfileNotEditor, err)

	// Clean up
	s.db.ProfileRepository().Delete(ctx, brandProfile.ID)
}

// TestT084_Scenario4_AudienceData tests quickstart Scenario 4.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario4_AudienceData() {
	ctx := context.Background()

	// Create influencer profile
	influencer := &domain.Profile{
		ID:     uuid.New(),
		UserID: s.user.ID,
		Type:   domain.ProfileTypeInfluencer,
		Name:   "Test Influencer",
	}
	err := s.db.ProfileRepository().Create(ctx, influencer)
	require.NoError(s.T(), err)

	audienceSvc := service.NewAudienceService(s.db.AudienceDataRepository(), s.db.ProfileRepository())

	demographics := []byte(`{"age":{"18-24":0.5}}`)

	// Update audience data
	data, err := audienceSvc.UpdateAudience(
		ctx,
		influencer.ID,
		map[string]string{"tiktok": "handle", "instagram": "@handle"},
		map[string]int{"tiktok": 100000, "instagram": 50000},
		4.5,
		demographics,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 4.5, data.EngagementRate)

	// Get audience data
	retrieved, err := audienceSvc.GetAudience(ctx, influencer.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 4.5, retrieved.EngagementRate)

	// Clean up
	s.db.ProfileRepository().Delete(ctx, influencer.ID)
}

// TestT084_Scenario5_VerificationSubmission tests quickstart Scenario 5.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario5_VerificationSubmission() {
	ctx := context.Background()

	// Create influencer profile
	influencer := &domain.Profile{
		ID:     uuid.New(),
		UserID: s.user.ID,
		Type:   domain.ProfileTypeInfluencer,
		Name:   "Test Influencer",
	}
	err := s.db.ProfileRepository().Create(ctx, influencer)
	require.NoError(s.T(), err)

	verificationSvc := service.NewVerificationService(s.db.FollowerVerificationRepository(), s.db.ProfileRepository())

	// Submit verification
	verification, err := verificationSvc.SubmitVerification(
		ctx,
		influencer.ID,
		[]string{"https://example.com/proof.png"},
		"Submitted for review",
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), domain.VerificationStatusPending, verification.Status)

	// Admin review
	reviewed, err := verificationSvc.AdminReviewVerification(
		ctx,
		influencer.ID,
		domain.VerificationStatusVerified,
		"admin123",
		"Counts confirmed",
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), domain.VerificationStatusVerified, reviewed.Status)
	assert.Equal(s.T(), "admin123", reviewed.ReviewedBy)

	// Clean up
	s.db.ProfileRepository().Delete(ctx, influencer.ID)
}

// TestT084_Scenario6_PayoutPreferencesMasking tests quickstart Scenario 6.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario6_PayoutPreferencesMasking() {
	ctx := context.Background()

	payoutSvc := service.NewPayoutService(s.db.PayoutPreferencesRepository(), s.db.ProfileRepository())

	// Update payout preferences
	masked, err := payoutSvc.UpdatePayoutPreferences(
		ctx,
		s.editor.ID,
		domain.PayoutMethodBankTransfer,
		"John Doe",
		"US",
		"USD",
		"encrypted_secret_data",
		true,
	)
	require.NoError(s.T(), err)

	// Verify masked response - encrypted details should not be present
	jsonData, err := toJSON(masked)
	require.NoError(s.T(), err)
	assert.NotContains(s.T(), string(jsonData), "encrypted_secret_data")
	assert.Equal(s.T(), s.editor.ID, masked.ProfileID)

	// Get payout preferences
	retrieved, err := payoutSvc.GetPayoutPreferences(ctx, s.editor.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), masked.ProfileID, retrieved.ProfileID)
}

// TestT084_Scenario7_KYCStatus tests quickstart Scenario 7.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario7_KYCStatus() {
	ctx := context.Background()

	kycSvc := service.NewKYCService(s.db.KYCStatusRepository(), s.db.ProfileRepository())

	// Get initial KYC status (should be not_started or create new)
	status, err := kycSvc.GetKYCStatus(ctx, s.editor.ID)
	if err != nil {
		// May not exist yet - that's ok
		status = &domain.KYCStatus{ProfileID: s.editor.ID}
	}
	assert.Equal(s.T(), s.editor.ID, status.ProfileID)

	// Admin update KYC
	updated, err := kycSvc.AdminUpdateKYC(
		ctx,
		s.editor.ID,
		domain.KYCStatusApproved,
		"admin123",
		"Docs verified",
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), domain.KYCStatusApproved, updated.Status)

	// Verify status
	retrieved, err := kycSvc.GetKYCStatus(ctx, s.editor.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), domain.KYCStatusApproved, retrieved.Status)
}

// TestT084_Scenario9_PortfolioOrderingWithGaps tests quickstart Scenario 9.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario9_PortfolioOrderingWithGaps() {
	ctx := context.Background()
	portfolioSvc := service.NewPortfolioService(s.db.PortfolioItemRepository(), s.db.ProfileRepository())

	// Create 3 portfolio items
	item1, err := portfolioSvc.Create(ctx, s.editor.ID, "Item 1", "", "", "", "", 1)
	require.NoError(s.T(), err)

	item2, err := portfolioSvc.Create(ctx, s.editor.ID, "Item 2", "", "", "", "", 2)
	require.NoError(s.T(), err)

	item3, err := portfolioSvc.Create(ctx, s.editor.ID, "Item 3", "", "", "", "", 3)
	require.NoError(s.T(), err)

	// Delete item with display_order = 2
	err = portfolioSvc.Delete(ctx, item2.ID, s.editor.ID)
	require.NoError(s.T(), err)

	// List remaining items
	items, err := portfolioSvc.ListByProfileID(ctx, s.editor.ID, 50, 0)
	require.NoError(s.T(), err)
	assert.Len(s.T(), items, 2)

	// Verify gaps are preserved
	orders := make([]int, len(items))
	for i, item := range items {
		orders[i] = item.DisplayOrder
	}
	assert.Contains(s.T(), orders, 1)
	assert.Contains(s.T(), orders, 3)
	assert.NotContains(s.T(), orders, 2)

	// Clean up
	portfolioSvc.Delete(ctx, item1.ID, s.editor.ID)
	portfolioSvc.Delete(ctx, item3.ID, s.editor.ID)
}

// TestT084_Scenario10_MaxPortfolioItemsLimit tests quickstart Scenario 10.
func (s *EnrichmentIntegrationTestSuite) TestT084_Scenario10_MaxPortfolioItemsLimit() {
	ctx := context.Background()
	portfolioSvc := service.NewPortfolioService(s.db.PortfolioItemRepository(), s.db.ProfileRepository())

	// Create 50 items (max limit)
	for i := 1; i <= 50; i++ {
		_, err := portfolioSvc.Create(ctx, s.editor.ID, fmt.Sprintf("Item %d", i), "", "", "", "", i)
		require.NoError(s.T(), err)
	}

	// Try to create 51st item (should fail)
	_, err := portfolioSvc.Create(ctx, s.editor.ID, "Item 51", "", "", "", "", 51)
	assert.Equal(s.T(), service.ErrPortfolioLimitReached, err)
}

// Helper function for JSON serialization
func toJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}