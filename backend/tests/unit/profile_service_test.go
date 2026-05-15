package unit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// mockProfileRepo implements repository.ProfileRepository for testing.
type mockProfileRepo struct {
	profiles map[uuid.UUID]*domain.Profile
	byUser   map[uuid.UUID][]*domain.Profile
}

func newMockProfileRepo() *mockProfileRepo {
	return &mockProfileRepo{
		profiles: make(map[uuid.UUID]*domain.Profile),
		byUser:   make(map[uuid.UUID][]*domain.Profile),
	}
}

func (r *mockProfileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	r.byUser[profile.UserID] = append(r.byUser[profile.UserID], profile)
	return nil
}

func (r *mockProfileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := r.profiles[id]; ok {
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (r *mockProfileRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	if profiles, ok := r.byUser[userID]; ok {
		return profiles, nil
	}
	return nil, nil
}

func (r *mockProfileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	if _, ok := r.profiles[profile.ID]; !ok {
		return domain.ErrProfileNotFound
	}
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := r.profiles[id]; ok {
		p.SoftDelete()
	}
	return nil
}

// mockProfileUserRepo implements repository.UserRepository for testing.
type mockProfileUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockProfileUserRepo() *mockProfileUserRepo {
	return &mockProfileUserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *mockProfileUserRepo) Create(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockProfileUserRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockProfileUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *mockProfileUserRepo) Update(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

// TestCreateProfile tests the CreateProfile method.
func TestCreateProfile(t *testing.T) {
	ctx := context.Background()
	profileRepo := newMockProfileRepo()
	userRepo := newMockProfileUserRepo()
	svc := service.NewProfileService(profileRepo, userRepo)

	t.Run("successful brand profile creation", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID, Email: "brand@example.com"}

		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeBrand,
			Name: "My Brand",
			Details: map[string]interface{}{
				"company_name": "Acme Corp",
				"size":         "100-500",
				"industry":     "Technology",
			},
		}

		profile, err := svc.CreateProfile(ctx, userID, req)
		require.NoError(t, err)
		assert.Equal(t, domain.ProfileTypeBrand, profile.Type)
		assert.Equal(t, "My Brand", profile.Name)
		assert.NotEmpty(t, profile.ID)
	})

	t.Run("successful editor profile creation", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID, Email: "editor@example.com"}

		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeEditor,
			Name: "My Editor Profile",
			Details: map[string]interface{}{
				"specializations": []interface{}{"video", "photo"},
				"portfolio_url":   "https://portfolio.example.com",
			},
		}

		profile, err := svc.CreateProfile(ctx, userID, req)
		require.NoError(t, err)
		assert.Equal(t, domain.ProfileTypeEditor, profile.Type)
	})

	t.Run("successful influencer profile creation", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID, Email: "influencer@example.com"}

		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeInfluencer,
			Name: "My Influencer Profile",
			Details: map[string]interface{}{
				"platforms":       []interface{}{"instagram", "youtube"},
				"follower_counts": map[string]interface{}{"instagram": 10000, "youtube": 5000},
			},
		}

		profile, err := svc.CreateProfile(ctx, userID, req)
		require.NoError(t, err)
		assert.Equal(t, domain.ProfileTypeInfluencer, profile.Type)
	})

	t.Run("invalid profile type", func(t *testing.T) {
		userID := uuid.New()
		req := &service.CreateProfileRequest{
			Type:    "invalid",
			Name:    "Test",
			Details: map[string]interface{}{},
		}

		_, err := svc.CreateProfile(ctx, userID, req)
		assert.ErrorIs(t, err, domain.ErrInvalidProfileType)
	})

	t.Run("brand missing company_name", func(t *testing.T) {
		userID := uuid.New()
		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeBrand,
			Name: "Test Brand",
			Details: map[string]interface{}{
				"size":     "100-500",
				"industry": "Technology",
			},
		}

		_, err := svc.CreateProfile(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "company_name")
	})

	t.Run("brand missing size", func(t *testing.T) {
		userID := uuid.New()
		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeBrand,
			Name: "Test Brand",
			Details: map[string]interface{}{
				"company_name": "Acme",
				"industry":     "Technology",
			},
		}

		_, err := svc.CreateProfile(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size")
	})

	t.Run("editor missing specializations", func(t *testing.T) {
		userID := uuid.New()
		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeEditor,
			Name: "Test Editor",
			Details: map[string]interface{}{
				"portfolio_url": "https://portfolio.example.com",
			},
		}

		_, err := svc.CreateProfile(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "specializations")
	})

	t.Run("editor empty specializations array", func(t *testing.T) {
		userID := uuid.New()
		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeEditor,
			Name: "Test Editor",
			Details: map[string]interface{}{
				"specializations": []interface{}{},
				"portfolio_url":   "https://portfolio.example.com",
			},
		}

		_, err := svc.CreateProfile(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-empty")
	})

	t.Run("influencer missing platforms", func(t *testing.T) {
		userID := uuid.New()
		req := &service.CreateProfileRequest{
			Type: domain.ProfileTypeInfluencer,
			Name: "Test Influencer",
			Details: map[string]interface{}{
				"follower_counts": map[string]interface{}{},
			},
		}

		_, err := svc.CreateProfile(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "platforms")
	})

	t.Run("duplicate profile type prevented", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID, Email: "dup@example.com"}

		// Create first brand profile
		req1 := &service.CreateProfileRequest{
			Type: domain.ProfileTypeBrand,
			Name: "First Brand",
			Details: map[string]interface{}{
				"company_name": "Acme",
				"size":         "100",
				"industry":     "Tech",
			},
		}
		_, err := svc.CreateProfile(ctx, userID, req1)
		require.NoError(t, err)

		// Try to create second brand profile
		req2 := &service.CreateProfileRequest{
			Type: domain.ProfileTypeBrand,
			Name: "Second Brand",
			Details: map[string]interface{}{
				"company_name": "Other",
				"size":         "50",
				"industry":     "Retail",
			},
		}
		_, err = svc.CreateProfile(ctx, userID, req2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("can create multiple different profile types", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID, Email: "multi@example.com"}

		// Create brand
		req1 := &service.CreateProfileRequest{
			Type: domain.ProfileTypeBrand,
			Name: "Brand Profile",
			Details: map[string]interface{}{
				"company_name": "Acme",
				"size":         "100",
				"industry":     "Tech",
			},
		}
		_, err := svc.CreateProfile(ctx, userID, req1)
		require.NoError(t, err)

		// Create editor
		req2 := &service.CreateProfileRequest{
			Type: domain.ProfileTypeEditor,
			Name: "Editor Profile",
			Details: map[string]interface{}{
				"specializations": []interface{}{"video"},
				"portfolio_url":   "https://portfolio.example.com",
			},
		}
		_, err = svc.CreateProfile(ctx, userID, req2)
		require.NoError(t, err)

		// Create influencer
		req3 := &service.CreateProfileRequest{
			Type: domain.ProfileTypeInfluencer,
			Name: "Influencer Profile",
			Details: map[string]interface{}{
				"platforms":       []interface{}{"instagram"},
				"follower_counts": map[string]interface{}{"instagram": 1000},
			},
		}
		_, err = svc.CreateProfile(ctx, userID, req3)
		require.NoError(t, err)
	})
}

// TestListProfiles tests the ListProfiles method.
func TestListProfiles(t *testing.T) {
	ctx := context.Background()
	profileRepo := newMockProfileRepo()
	userRepo := newMockProfileUserRepo()
	svc := service.NewProfileService(profileRepo, userRepo)

	t.Run("list empty profiles", func(t *testing.T) {
		userID := uuid.New()
		profiles, err := svc.ListProfiles(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, profiles)
	})

	t.Run("list all active profiles", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID}

		// Create a profile
		details, _ := json.Marshal(map[string]interface{}{"company_name": "Acme"})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Test Brand", details)
		profileRepo.Create(ctx, profile)

		profiles, err := svc.ListProfiles(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, profiles, 1)
		assert.Equal(t, "Test Brand", profiles[0].Name)
	})

	t.Run("exclude soft-deleted profiles", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID}

		// Create an active profile
		details, _ := json.Marshal(map[string]interface{}{"company_name": "Active"})
		active := domain.NewProfile(userID, domain.ProfileTypeBrand, "Active Brand", details)
		profileRepo.Create(ctx, active)

		// Create a soft-deleted profile
		details2, _ := json.Marshal(map[string]interface{}{"company_name": "Deleted"})
		deleted := domain.NewProfile(userID, domain.ProfileTypeEditor, "Deleted Editor", details2)
		deleted.SoftDelete()
		profileRepo.Create(ctx, deleted)

		profiles, err := svc.ListProfiles(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, profiles, 1)
		assert.Equal(t, "Active Brand", profiles[0].Name)
	})
}

// TestGetProfile tests the GetProfile method.
func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	profileRepo := newMockProfileRepo()
	userRepo := newMockProfileUserRepo()
	svc := service.NewProfileService(profileRepo, userRepo)

	t.Run("get own profile", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID}

		details, _ := json.Marshal(map[string]interface{}{"company_name": "Acme"})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Test Brand", details)
		profileRepo.Create(ctx, profile)

		found, err := svc.GetProfile(ctx, userID, profile.ID)
		require.NoError(t, err)
		assert.Equal(t, profile.ID, found.ID)
	})

	t.Run("get profile of another user", func(t *testing.T) {
		ownerID := uuid.New()
		otherID := uuid.New()
		userRepo.users[ownerID] = &domain.User{ID: ownerID}
		userRepo.users[otherID] = &domain.User{ID: otherID}

		details, _ := json.Marshal(map[string]interface{}{"company_name": "Acme"})
		profile := domain.NewProfile(ownerID, domain.ProfileTypeBrand, "Test Brand", details)
		profileRepo.Create(ctx, profile)

		_, err := svc.GetProfile(ctx, otherID, profile.ID)
		assert.ErrorIs(t, err, domain.ErrProfileNotOwned)
	})

	t.Run("get deleted profile", func(t *testing.T) {
		userID := uuid.New()
		userRepo.users[userID] = &domain.User{ID: userID}

		details, _ := json.Marshal(map[string]interface{}{"company_name": "Acme"})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Deleted Brand", details)
		profile.SoftDelete()
		profileRepo.Create(ctx, profile)

		_, err := svc.GetProfile(ctx, userID, profile.ID)
		assert.ErrorIs(t, err, domain.ErrProfileNotFound)
	})
}

// TestValidateProfileType tests the ValidateProfileType function.
func TestValidateProfileType(t *testing.T) {
	t.Run("valid brand details", func(t *testing.T) {
		details := map[string]interface{}{
			"company_name": "Acme",
			"size":         "100-500",
			"industry":     "Technology",
		}
		err := service.ValidateProfileType(domain.ProfileTypeBrand, details)
		require.NoError(t, err)
	})

	t.Run("valid editor details", func(t *testing.T) {
		details := map[string]interface{}{
			"specializations": []interface{}{"video", "photo"},
			"portfolio_url":   "https://portfolio.example.com",
		}
		err := service.ValidateProfileType(domain.ProfileTypeEditor, details)
		require.NoError(t, err)
	})

	t.Run("valid influencer details", func(t *testing.T) {
		details := map[string]interface{}{
			"platforms":       []interface{}{"instagram", "youtube"},
			"follower_counts": map[string]interface{}{"instagram": 10000},
		}
		err := service.ValidateProfileType(domain.ProfileTypeInfluencer, details)
		require.NoError(t, err)
	})

	t.Run("invalid profile type", func(t *testing.T) {
		err := service.ValidateProfileType("invalid", map[string]interface{}{})
		assert.ErrorIs(t, err, domain.ErrInvalidProfileType)
	})
}
