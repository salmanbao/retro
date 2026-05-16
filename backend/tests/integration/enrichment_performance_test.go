package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"viralforge/backend/src/adapter"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// PerformanceBenchmarkSuite runs performance benchmarks for enrichment endpoints.
type PerformanceBenchmarkSuite struct {
	suite.Suite
	db     *adapter.PostgresStore
	server *httptest.Server
	router *chi.Mux

	userID        uuid.UUID
	profileID     uuid.UUID
	enrichmentID  uuid.UUID

	profileEnrichmentSvc *service.ProfileEnrichmentService
	portfolioSvc         *service.PortfolioService
}

func (s *PerformanceBenchmarkSuite) SetupSuite() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/viralforge?sslmode=disable"
	}

	// Connect to PostgreSQL
	gormDB, err := adapter.Connect(context.Background(), databaseURL)
	if err != nil {
		s.T().Skipf("Skipping benchmark test: database not available: %v", err)
		return
	}
	s.db = adapter.NewPostgresStore(gormDB)

	if err := s.db.AutoMigrate(); err != nil {
		s.T().Skipf("Skipping benchmark test: migration failed: %v", err)
		return
	}

	// Create test data
	ctx := context.Background()

	// Create user
	s.userID = uuid.New()
	user := &domain.User{
		ID:           s.userID,
		Email:        fmt.Sprintf("perf-%s@example.com", uuid.New().String()[:8]),
		PasswordHash: "hash",
		Verified:     true,
	}
	err = s.db.UserRepository().Create(ctx, user)
	if err != nil {
		s.T().Skipf("Skipping benchmark test: failed to create user: %v", err)
		return
	}

	// Create profile
	s.profileID = uuid.New()
	profile := &domain.Profile{
		ID:     s.profileID,
		UserID: s.userID,
		Type:   domain.ProfileTypeEditor,
		Name:   "Perf Test Editor",
	}
	err = s.db.ProfileRepository().Create(ctx, profile)
	if err != nil {
		s.T().Skipf("Skipping benchmark test: failed to create profile: %v", err)
		return
	}

	// Create profile enrichment
	s.enrichmentID = uuid.New()
	enrichment := &domain.ProfileEnrichment{
		ID:        s.enrichmentID,
		ProfileID: s.profileID,
		Bio:       "Performance test bio",
		Location:  "New York",
		Languages: []string{"en", "es"},
		Timezone:  "America/New_York",
	}
	err = s.db.ProfileEnrichmentRepository().Create(ctx, enrichment)
	if err != nil {
		s.T().Skipf("Skipping benchmark test: failed to create enrichment: %v", err)
		return
	}

	// Create services
	s.profileEnrichmentSvc = service.NewProfileEnrichmentService(
		s.db.ProfileEnrichmentRepository(),
		s.db.ProfileRepository(),
	)
	portfolioSvc := service.NewPortfolioService(
		s.db.PortfolioItemRepository(),
		s.db.ProfileRepository(),
	)
	s.portfolioSvc = portfolioSvc
	portfolioHandler := handler.NewPortfolioHandler(portfolioSvc)

	// Setup chi router for benchmark
	s.router = chi.NewRouter()
	s.router.Route("/api/v1/profiles/{id}/details", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, s.profileID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		enrichmentHandler := handler.NewProfileEnrichmentHandler(s.profileEnrichmentSvc)
		r.Get("/", enrichmentHandler.GetDetails)
	})
	s.router.Route("/api/v1/profiles/{id}/portfolio", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, s.profileID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Get("/", portfolioHandler.GetPortfolio)
	})

	s.server = httptest.NewServer(s.router)
}

func (s *PerformanceBenchmarkSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.db != nil {
		ctx := context.Background()
		if s.profileID != uuid.Nil {
			s.db.ProfileRepository().Delete(ctx, s.profileID)
		}
		if s.userID != uuid.Nil {
			// Clean up user
			if user, err := s.db.UserRepository().ByID(ctx, s.userID); err == nil {
				s.db.UserRepository().Update(ctx, user)
			}
		}
	}
}

func TestPerformanceBenchmarkSuite(t *testing.T) {
	suite.Run(t, new(PerformanceBenchmarkSuite))
}

// TestT086_GETDetails_ResponseTime measures GET /details response time.
func (s *PerformanceBenchmarkSuite) TestT086_GETDetails_ResponseTime() {
	url := fmt.Sprintf("%s/api/v1/profiles/%s/details", s.server.URL, s.profileID.String())

	// Verify target: < 200ms
	var total time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		resp, err := http.Get(url)
		total += time.Since(start)
		if err != nil {
			s.T().Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			s.T().Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	}
	avgMs := total.Seconds() * 1000 / 100
	s.T().Logf("Average response time: %.2fms (target: <200ms)", avgMs)

	if avgMs >= 200 {
		s.T().Errorf("Response time %.2fms exceeds 200ms target", avgMs)
	}
}

// TestT086_GETPortfolio_ResponseTime measures GET /portfolio response time.
func (s *PerformanceBenchmarkSuite) TestT086_GETPortfolio_ResponseTime() {
	// Create some portfolio items first
	ctx := context.Background()
	portfolioSvc := service.NewPortfolioService(
		s.db.PortfolioItemRepository(),
		s.db.ProfileRepository(),
	)
	for i := 0; i < 10; i++ {
		portfolioSvc.Create(ctx, s.profileID, fmt.Sprintf("Item %d", i), "", "", "", "", i)
	}

	url := fmt.Sprintf("%s/api/v1/profiles/%s/portfolio", s.server.URL, s.profileID.String())

	var total time.Duration
	for i := 0; i < 100; i++ {
		start := time.Now()
		resp, err := http.Get(url)
		total += time.Since(start)
		if err != nil {
			s.T().Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}
	avgMs := total.Seconds() * 1000 / 100
	s.T().Logf("Average GET /portfolio response time: %.2fms", avgMs)
}

// TestT086_ConcurrentRequests measures concurrent request handling.
func (s *PerformanceBenchmarkSuite) TestT086_ConcurrentRequests() {
	url := fmt.Sprintf("%s/api/v1/profiles/%s/details", s.server.URL, s.profileID.String())

	count := int32(0)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				return
			}
			resp.Body.Close()
			atomic.AddInt32(&count, 1)
		}()
	}
	wg.Wait()
	s.T().Logf("Completed %d concurrent requests", count)
}

// TestT086_PerformanceTargets verifies all performance targets are met.
func (s *PerformanceBenchmarkSuite) TestT086_PerformanceTargets() {
	url := fmt.Sprintf("%s/api/v1/profiles/%s/details", s.server.URL, s.profileID.String())

	// Measure 100 requests
	var times []float64
	for i := 0; i < 100; i++ {
		start := time.Now()
		resp, err := http.Get(url)
		elapsed := time.Since(start).Seconds() * 1000
		if err != nil {
			s.T().Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			s.T().Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
		times = append(times, elapsed)
	}

	// Calculate average
	var sum, max float64
	for _, t := range times {
		sum += t
		if t > max {
			max = t
		}
	}
	avg := sum / float64(len(times))

	s.T().Logf("GET /details average: %.2fms, max: %.2fms (target: <200ms avg)", avg, max)
	require.Less(s.T(), avg, 200.0, "Average response time should be under 200ms")
}

// TestT086_P99Latency calculates P99 latency.
func (s *PerformanceBenchmarkSuite) TestT086_P99Latency() {
	url := fmt.Sprintf("%s/api/v1/profiles/%s/details", s.server.URL, s.profileID.String())

	// Collect 1000 samples
	var times []float64
	for i := 0; i < 1000; i++ {
		start := time.Now()
		resp, err := http.Get(url)
		elapsed := time.Since(start).Seconds() * 1000
		if err != nil {
			continue
		}
		resp.Body.Close()
		times = append(times, elapsed)
	}

	// Calculate P99
	sort.Float64s(times)
	p99Index := int(float64(len(times)) * 0.99)
	if p99Index >= len(times) {
		p99Index = len(times) - 1
	}
	p99 := times[p99Index]

	s.T().Logf("P99 latency: %.2fms", p99)
	require.Less(s.T(), p99, 300.0, "P99 latency should be under 300ms")
}