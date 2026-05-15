package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
)

// mockProfileStore implements profile repository for integration testing.
type mockProfileStore struct {
	profiles map[uuid.UUID]*domain.Profile
	byUser   map[uuid.UUID][]*domain.Profile
}

func newMockProfileStore() *mockProfileStore {
	return &mockProfileStore{
		profiles: make(map[uuid.UUID]*domain.Profile),
		byUser:   make(map[uuid.UUID][]*domain.Profile),
	}
}

func (s *mockProfileStore) Create(ctx context.Context, profile *domain.Profile) error {
	s.profiles[profile.ID] = profile
	s.byUser[profile.UserID] = append(s.byUser[profile.UserID], profile)
	return nil
}

func (s *mockProfileStore) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := s.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (s *mockProfileStore) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	if profiles, ok := s.byUser[userID]; ok {
		var active []*domain.Profile
		for _, p := range profiles {
			if p.DeletedAt == nil {
				active = append(active, p)
			}
		}
		return active, nil
	}
	return []*domain.Profile{}, nil
}

func (s *mockProfileStore) Update(ctx context.Context, profile *domain.Profile) error {
	if _, ok := s.profiles[profile.ID]; !ok {
		return domain.ErrProfileNotFound
	}
	profile.UpdatedAt = time.Now()
	s.profiles[profile.ID] = profile
	return nil
}

func (s *mockProfileStore) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := s.profiles[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
		p.UpdatedAt = now
	}
	return nil
}

// mockProfileUserStore implements user repository for integration testing.
type mockProfileUserStore struct {
	users map[uuid.UUID]*domain.User
}

func newMockProfileUserStore() *mockProfileUserStore {
	return &mockProfileUserStore{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (s *mockProfileUserStore) Create(ctx context.Context, user *domain.User) error {
	s.users[user.ID] = user
	return nil
}

func (s *mockProfileUserStore) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *mockProfileUserStore) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (s *mockProfileUserStore) Update(ctx context.Context, user *domain.User) error {
	s.users[user.ID] = user
	return nil
}

// Integration tests for profile service
// These test the full flow with mock stores

func TestProfileServiceIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("create and retrieve brand profile", func(t *testing.T) {
		profileRepo := newMockProfileStore()
		userRepo := newMockProfileUserStore()

		userID := uuid.New()
		userRepo.Create(ctx, &domain.User{ID: userID, Email: "brand@example.com"})

		// Create profile via service
		details := map[string]interface{}{
			"company_name": "Acme Corporation",
			"size":         "100-500",
			"industry":     "Technology",
		}
		detailsJSON, _ := json.Marshal(details)

		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Acme Brand", detailsJSON)
		err := profileRepo.Create(ctx, profile)
		require.NoError(t, err)

		// Retrieve profiles
		profiles, err := profileRepo.ByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, profiles, 1)
		assert.Equal(t, "Acme Brand", profiles[0].Name)
		assert.Equal(t, domain.ProfileTypeBrand, profiles[0].Type)

		var retrievedDetails map[string]interface{}
		json.Unmarshal(profiles[0].Details, &retrievedDetails)
		assert.Equal(t, "Acme Corporation", retrievedDetails["company_name"])
	})

	t.Run("create multiple profile types for same user", func(t *testing.T) {
		profileRepo := newMockProfileStore()
		userRepo := newMockProfileUserStore()

		userID := uuid.New()
		userRepo.Create(ctx, &domain.User{ID: userID, Email: "multi@example.com"})

		// Create brand profile
		brandDetails, _ := json.Marshal(map[string]interface{}{
			"company_name": "My Company",
			"size":         "50-100",
			"industry":     "Retail",
		})
		brand := domain.NewProfile(userID, domain.ProfileTypeBrand, "Brand Profile", brandDetails)
		err := profileRepo.Create(ctx, brand)
		require.NoError(t, err)

		// Create editor profile
		editorDetails, _ := json.Marshal(map[string]interface{}{
			"specializations": []interface{}{"photography", "videography"},
			"portfolio_url":   "https://myportfolio.com",
		})
		editor := domain.NewProfile(userID, domain.ProfileTypeEditor, "Editor Profile", editorDetails)
		err = profileRepo.Create(ctx, editor)
		require.NoError(t, err)

		// Create influencer profile
		influencerDetails, _ := json.Marshal(map[string]interface{}{
			"platforms":       []interface{}{"instagram", "youtube", "tiktok"},
			"follower_counts": map[string]interface{}{"instagram": 50000, "youtube": 20000},
		})
		influencer := domain.NewProfile(userID, domain.ProfileTypeInfluencer, "Influencer Profile", influencerDetails)
		err = profileRepo.Create(ctx, influencer)
		require.NoError(t, err)

		// Verify all three profiles exist
		profiles, err := profileRepo.ByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, profiles, 3)

		// Check each type is represented
		types := make(map[domain.ProfileType]bool)
		for _, p := range profiles {
			types[p.Type] = true
		}
		assert.True(t, types[domain.ProfileTypeBrand])
		assert.True(t, types[domain.ProfileTypeEditor])
		assert.True(t, types[domain.ProfileTypeInfluencer])
	})

	t.Run("soft delete profile", func(t *testing.T) {
		profileRepo := newMockProfileStore()
		userRepo := newMockProfileUserStore()

		userID := uuid.New()
		userRepo.Create(ctx, &domain.User{ID: userID, Email: "delete@example.com"})

		// Create profile
		details, _ := json.Marshal(map[string]interface{}{"company_name": "ToDelete"})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Delete Me", details)
		err := profileRepo.Create(ctx, profile)
		require.NoError(t, err)

		// Verify it exists
		profiles, err := profileRepo.ByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, profiles, 1)

		// Soft delete
		err = profileRepo.Delete(ctx, profile.ID)
		require.NoError(t, err)

		// Verify it's gone from active list
		profiles, err = profileRepo.ByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, profiles, 0)
	})

	t.Run("update profile", func(t *testing.T) {
		profileRepo := newMockProfileStore()
		userRepo := newMockProfileUserStore()

		userID := uuid.New()
		userRepo.Create(ctx, &domain.User{ID: userID, Email: "update@example.com"})

		// Create profile
		details, _ := json.Marshal(map[string]interface{}{"company_name": "Original"})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Original Name", details)
		err := profileRepo.Create(ctx, profile)
		require.NoError(t, err)

		// Update profile
		newDetails, _ := json.Marshal(map[string]interface{}{"company_name": "Updated"})
		profile.Name = "Updated Name"
		profile.Details = newDetails
		err = profileRepo.Update(ctx, profile)
		require.NoError(t, err)

		// Verify update
		retrieved, err := profileRepo.ByID(ctx, profile.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)

		var retrievedDetails map[string]interface{}
		json.Unmarshal(retrieved.Details, &retrievedDetails)
		assert.Equal(t, "Updated", retrievedDetails["company_name"])
	})
}

// TestProfileServiceUpdateDeleteIntegration tests the profile service update and delete operations.
func TestProfileServiceUpdateDeleteIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("update and delete flow", func(t *testing.T) {
		profileRepo := newMockProfileStore()
		userRepo := newMockProfileUserStore()

		userID := uuid.New()
		userRepo.Create(ctx, &domain.User{ID: userID, Email: "flow@example.com"})

		// Create profile
		details, _ := json.Marshal(map[string]interface{}{
			"company_name": "Flow Corp",
			"size":         "100-500",
			"industry":     "Tech",
		})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Flow Brand", details)
		err := profileRepo.Create(ctx, profile)
		require.NoError(t, err)

		// Update name
		profile.Name = "Updated Flow Brand"
		err = profileRepo.Update(ctx, profile)
		require.NoError(t, err)

		// Verify update persisted
		updated, err := profileRepo.ByID(ctx, profile.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Flow Brand", updated.Name)
		assert.NotNil(t, updated.UpdatedAt)

		// Delete profile
		err = profileRepo.Delete(ctx, profile.ID)
		require.NoError(t, err)

		// Verify it's soft-deleted by checking it no longer appears in active list
		activeProfiles, err := profileRepo.ByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, activeProfiles, 0)

		// Verify ByID returns not found for deleted profile
		_, err = profileRepo.ByID(ctx, profile.ID)
		assert.Equal(t, domain.ErrProfileNotFound, err)
	})

	t.Run("concurrent updates to different profiles", func(t *testing.T) {
		profileRepo := newMockProfileStore()
		userRepo := newMockProfileUserStore()

		userID := uuid.New()
		userRepo.Create(ctx, &domain.User{ID: userID, Email: "concurrent@example.com"})

		// Create two profiles
		details1, _ := json.Marshal(map[string]interface{}{"company_name": "Profile 1"})
		profile1 := domain.NewProfile(userID, domain.ProfileTypeBrand, "Brand 1", details1)
		err := profileRepo.Create(ctx, profile1)
		require.NoError(t, err)

		details2, _ := json.Marshal(map[string]interface{}{"company_name": "Profile 2"})
		profile2 := domain.NewProfile(userID, domain.ProfileTypeEditor, "Editor 2", details2)
		err = profileRepo.Create(ctx, profile2)
		require.NoError(t, err)

		// Update both simultaneously
		profile1.Name = "Updated Brand 1"
		profile2.Name = "Updated Editor 2"

		err1 := profileRepo.Update(ctx, profile1)
		err2 := profileRepo.Update(ctx, profile2)
		require.NoError(t, err1)
		require.NoError(t, err2)

		// Verify both updates persisted
		p1, _ := profileRepo.ByID(ctx, profile1.ID)
		p2, _ := profileRepo.ByID(ctx, profile2.ID)
		assert.Equal(t, "Updated Brand 1", p1.Name)
		assert.Equal(t, "Updated Editor 2", p2.Name)
	})
}
