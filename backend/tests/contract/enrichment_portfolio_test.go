package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// mockEnrichmentProfileRepo is a mock implementation of ProfileRepository.
type mockEnrichmentProfileRepo struct {
	profiles map[uuid.UUID]*domain.Profile
	byUser   map[uuid.UUID][]*domain.Profile
}

func newMockEnrichmentProfileStore() *mockEnrichmentProfileRepo {
	return &mockEnrichmentProfileRepo{
		profiles: make(map[uuid.UUID]*domain.Profile),
		byUser:   make(map[uuid.UUID][]*domain.Profile),
	}
}

func (m *mockEnrichmentProfileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	m.profiles[profile.ID] = profile
	m.byUser[profile.UserID] = append(m.byUser[profile.UserID], profile)
	return nil
}

func (m *mockEnrichmentProfileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := m.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (m *mockEnrichmentProfileRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	if profiles, ok := m.byUser[userID]; ok {
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

func (m *mockEnrichmentProfileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	m.profiles[profile.ID] = profile
	return nil
}

func (m *mockEnrichmentProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := m.profiles[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
	}
	return nil
}

// mockEnrichmentProfileEnrichmentRepo is a mock for ProfileEnrichmentRepository.
type mockEnrichmentProfileEnrichmentRepo struct {
	data map[uuid.UUID]*domain.ProfileEnrichment
}

func newMockProfileEnrichmentRepo() *mockEnrichmentProfileEnrichmentRepo {
	return &mockEnrichmentProfileEnrichmentRepo{
		data: make(map[uuid.UUID]*domain.ProfileEnrichment),
	}
}

func (m *mockEnrichmentProfileEnrichmentRepo) Create(ctx context.Context, enrichment *domain.ProfileEnrichment) error {
	m.data[enrichment.ProfileID] = enrichment
	return nil
}

func (m *mockEnrichmentProfileEnrichmentRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.ProfileEnrichment, error) {
	if e, ok := m.data[profileID]; ok {
		return e, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (m *mockEnrichmentProfileEnrichmentRepo) Update(ctx context.Context, enrichment *domain.ProfileEnrichment) error {
	m.data[enrichment.ProfileID] = enrichment
	return nil
}

// mockPortfolioItemRepo is a mock for PortfolioItemRepository.
type mockPortfolioItemRepo struct {
	items     map[uuid.UUID]*domain.PortfolioItem
	byProfile map[uuid.UUID][]*domain.PortfolioItem
}

func newMockPortfolioItemRepo() *mockPortfolioItemRepo {
	return &mockPortfolioItemRepo{
		items:     make(map[uuid.UUID]*domain.PortfolioItem),
		byProfile: make(map[uuid.UUID][]*domain.PortfolioItem),
	}
}

func (m *mockPortfolioItemRepo) Create(ctx context.Context, item *domain.PortfolioItem) error {
	m.items[item.ID] = item
	m.byProfile[item.ProfileID] = append(m.byProfile[item.ProfileID], item)
	return nil
}

func (m *mockPortfolioItemRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.PortfolioItem, error) {
	if item, ok := m.items[id]; ok {
		if item.DeletedAt != nil {
			return nil, domain.ErrPortfolioItemNotFound
		}
		return item, nil
	}
	return nil, domain.ErrPortfolioItemNotFound
}

func (m *mockPortfolioItemRepo) ByProfileID(ctx context.Context, profileID uuid.UUID, limit, offset int) ([]*domain.PortfolioItem, error) {
	var result []*domain.PortfolioItem
	for _, item := range m.byProfile[profileID] {
		if item.DeletedAt == nil {
			result = append(result, item)
		}
	}
	// Apply offset and limit
	if offset >= len(result) {
		return []*domain.PortfolioItem{}, nil
	}
	if limit > 0 {
		end := offset + limit
		if end > len(result) {
			end = len(result)
		}
		result = result[offset:end]
	} else {
		result = result[offset:]
	}
	return result, nil
}

func (m *mockPortfolioItemRepo) Update(ctx context.Context, item *domain.PortfolioItem) error {
	m.items[item.ID] = item
	return nil
}

func (m *mockPortfolioItemRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if item, ok := m.items[id]; ok {
		now := time.Now()
		item.DeletedAt = &now
		item.UpdatedAt = now
	}
	return nil
}

func (m *mockPortfolioItemRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	var count int64
	for _, item := range m.byProfile[profileID] {
		if item.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

// TestT081_PortfolioCRUDAsEditor tests full portfolio CRUD flow as Editor.
func TestT081_PortfolioCRUDAsEditor(t *testing.T) {
	profileRepo := newMockEnrichmentProfileStore()
	enrichmentRepo := newMockProfileEnrichmentRepo()
	portfolioRepo := newMockPortfolioItemRepo()

	// Create an editor profile
	userID := uuid.New()
	editorProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeEditor,
		Name:   "Editor Profile",
	}
	profileRepo.Create(context.Background(), editorProfile)

	// Create enrichment for the profile
	enrichmentRepo.Create(context.Background(), &domain.ProfileEnrichment{
		ProfileID: editorProfile.ID,
	})

	// Create services
	portfolioSvc := service.NewPortfolioService(portfolioRepo, profileRepo)

	// Create handler
	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)

	// Setup chi router
	router := chi.NewRouter()

	// Register portfolio routes with ownership middleware
	router.Route("/api/v1/profiles/{id}/portfolio", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, editorProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Get("/", portfolioHandler.GetPortfolio)
		r.Post("/", portfolioHandler.CreatePortfolioItem)
		r.Patch("/{itemId}", portfolioHandler.UpdatePortfolioItem)
		r.Delete("/{itemId}", portfolioHandler.DeletePortfolioItem)
	})

	// Test POST /api/v1/profiles/{id}/portfolio (create)
	createReq := handler.CreatePortfolioItemRequest{
		Title:        "Test Portfolio Item",
		Description:  "A test portfolio item",
		ThumbnailURL: "https://example.com/thumb.jpg",
		DisplayOrder: 1,
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "POST should create portfolio item")

	// Parse response to get item ID
	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	itemID, ok := createResp["id"].(string)
	assert.True(t, ok, "Response should contain item id")

	// Test GET /api/v1/profiles/{id}/portfolio (list)
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "GET should list portfolio items")

	var listResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.Len(t, listResp, 1, "Should have one portfolio item")
	assert.Equal(t, "Test Portfolio Item", listResp[0]["title"])

	// Test PATCH /api/v1/profiles/{id}/portfolio/{itemId} (update)
	itemUUID, _ := uuid.Parse(itemID)
	displayOrderInt := 2
	updateReq := handler.UpdatePortfolioItemRequest{
		DisplayOrder: &displayOrderInt,
	}
	body, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest(http.MethodPatch, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio/"+itemUUID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "PATCH should update portfolio item")

	// Verify update
	updatedItem, err := portfolioRepo.ByID(context.Background(), itemUUID)
	require.NoError(t, err)
	assert.Equal(t, 2, updatedItem.DisplayOrder, "Display order should be updated")

	// Test DELETE /api/v1/profiles/{id}/portfolio/{itemId}
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio/"+itemUUID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code, "DELETE should return 204")

	// Verify soft delete
	_, err = portfolioRepo.ByID(context.Background(), itemUUID)
	assert.Error(t, err, "Deleted item should not be found")
}

// TestT081_PortfolioRejectionForNonEditor tests that non-Editor profiles cannot access portfolio.
func TestT081_PortfolioRejectionForNonEditor(t *testing.T) {
	profileRepo := newMockEnrichmentProfileStore()
	portfolioRepo := newMockPortfolioItemRepo()

	// Create a brand profile (not editor)
	userID := uuid.New()
	brandProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeBrand,
		Name:   "Brand Profile",
	}
	profileRepo.Create(context.Background(), brandProfile)

	// Create services
	portfolioSvc := service.NewPortfolioService(portfolioRepo, profileRepo)
	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)

	router := chi.NewRouter()

	// Register portfolio routes
	router.Route("/api/v1/profiles/{id}/portfolio", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate context with brand profile
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, brandProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		// Note: RequireEditor middleware would block this, but for this test
		// we just verify the service rejects non-editor
		r.Get("/", portfolioHandler.GetPortfolio)
	})

	// Create mock middleware that simulates brand profile
	profileTypeMw := middleware.NewProfileTypeMiddleware(profileRepo)

	router2 := chi.NewRouter()
	router2.Route("/api/v1/profiles/{id}/portfolio", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, brandProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Use(profileTypeMw.RequireEditor)
		r.Get("/", portfolioHandler.GetPortfolio)
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/profiles/"+brandProfile.ID.String()+"/portfolio", nil)
	w := httptest.NewRecorder()
	router2.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "Brand profile should be rejected for portfolio access")
}

// TestT080_PortfolioOrderingWithGaps tests display_order gaps preserved after delete.
func TestT080_PortfolioOrderingWithGaps(t *testing.T) {
	profileRepo := newMockEnrichmentProfileStore()
	portfolioRepo := newMockPortfolioItemRepo()

	// Create an editor profile
	userID := uuid.New()
	editorProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeEditor,
		Name:   "Editor Profile",
	}
	profileRepo.Create(context.Background(), editorProfile)

	// Create portfolio service
	portfolioSvc := service.NewPortfolioService(portfolioRepo, profileRepo)
	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)

	router := chi.NewRouter()
	router.Route("/api/v1/profiles/{id}/portfolio", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, editorProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Get("/", portfolioHandler.GetPortfolio)
		r.Post("/", portfolioHandler.CreatePortfolioItem)
		r.Delete("/{itemId}", portfolioHandler.DeletePortfolioItem)
	})

	// Create 3 portfolio items with display_order 1, 2, 3
	itemIDs := []uuid.UUID{}
	for i := 1; i <= 3; i++ {
		createReq := handler.CreatePortfolioItemRequest{
			Title:        "Item " + string(rune('0'+i)),
			DisplayOrder: i, // Note: DisplayOrder is int, not *int
		}
		body, _ := json.Marshal(createReq)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		itemID, _ := uuid.Parse(resp["id"].(string))
		itemIDs = append(itemIDs, itemID)
	}

	// Delete item with display_order = 2 (itemIDs[1])
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio/"+itemIDs[1].String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// List remaining items
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/profiles/"+editorProfile.ID.String()+"/portfolio", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var items []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &items)
	assert.Len(t, items, 2, "Should have 2 items after delete")

	// Verify the remaining items have correct display_order (gaps preserved)
	orders := make([]int, len(items))
	for i, item := range items {
		orders[i] = int(item["display_order"].(float64))
	}
	assert.Contains(t, orders, 1, "Item with display_order 1 should remain")
	assert.Contains(t, orders, 3, "Item with display_order 3 should remain")
	assert.NotContains(t, orders, 2, "Gap at display_order 2 should be preserved")
}
