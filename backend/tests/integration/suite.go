package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

const (
	testBaseURL = "http://localhost:8080"
	testTimeout = 30 * time.Second
)

// TestSuite is the base integration test suite.
type TestSuite struct {
	T         *testing.T
	Ctx       context.Context
	Client    *fixtures.TestClient
	Container *fixtures.ContainerManager
	DB        interface{} // *gorm.DB - avoid import cycle
	Cancel    context.CancelFunc
}

// NewTestSuite creates a new test suite.
func NewTestSuite(t *testing.T) *TestSuite {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)

	// Check if server is already running
	client, err := fixtures.NewTestClient(testBaseURL)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Check if server is available
	resp, err := client.HttpClient.Get(testBaseURL + "/health")
	if err == nil {
		resp.Body.Close()
		// Server is running, no need for container
		return &TestSuite{
			T:         t,
			Ctx:       ctx,
			Client:    client,
			Container: nil,
			Cancel:    cancel,
		}
	}

	// Server not running - skip integration tests
	cancel()
	t.Skipf("Skipping integration test: server not running at %s. Start server with: DATABASE_URL='host=localhost port=5432 user=postgres password=postgres dbname=viralforge_test sslmode=disable' ./cmd/server", testBaseURL)
	return nil
}

// GenerateValidEmail generates a valid unique email for testing.
// Avoids using BaseURL which contains scheme that breaks email validation.
func (s *TestSuite) GenerateValidEmail(prefix string) string {
	return fmt.Sprintf("test-%s-%d@example.com", prefix, time.Now().UnixNano())
}

// TearDown cleans up after the test.
func (s *TestSuite) TearDown() {
	if s.Container != nil {
		s.Container.Stop(s.Ctx)
	}
	s.Cancel()
}

// SkipIfNoServer skips the test if the server is not running.
func (s *TestSuite) SkipIfNoServer() {
	// Try to reach the health endpoint
	resp, err := s.Client.HttpClient.Get(testBaseURL + "/health")
	if err != nil {
		s.T.Skipf("Skipping integration test: server not running at %s: %v", testBaseURL, err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

// SetupAuth sets up an authenticated user for testing.
func (s *TestSuite) SetupAuth() (*fixtures.User, error) {
	userFac := fixtures.NewUserFactory(s.Client)
	return userFac.CreateRegistered()
}

// SetupTwoUsers sets up two separate authenticated users.
func (s *TestSuite) SetupTwoUsers() (*fixtures.User, *fixtures.User, error) {
	crossUserFac := fixtures.NewCrossUserFactory(s.Client)
	return crossUserFac.CreateTwoUsers()
}

// RunMigrations runs database migrations.
// This is a placeholder - actual implementation needs to import domain models.
func (s *TestSuite) RunMigrations() error {
	// Import all domain packages to register models with GORM AutoMigrate
	// This is done in the actual test file

	// For now, we assume migrations run automatically via the server
	return nil
}

// Log logs a test message.
func (s *TestSuite) Log(format string, args ...interface{}) {
	log.Printf("[TEST] "+format, args...)
}

// AssertResponseOK asserts that the response status is 200 OK.
func (s *TestSuite) AssertResponseOK(resp *fixtures.ErrorResponse) {
	if resp.Error != "" {
		s.T.Errorf("Expected OK, got error: %s - %s", resp.Error, resp.Message)
	}
}

// AssertStatus asserts a specific status code.
func (s *TestSuite) AssertStatus(status int, expected int) {
	if status != expected {
		s.T.Errorf("Expected status %d, got %d", expected, status)
	}
}

// AssertNoError asserts no error occurred.
func (s *TestSuite) AssertNoError(err error) {
	if err != nil {
		s.T.Errorf("Unexpected error: %v", err)
	}
}

// AssertField asserts a field has an expected value.
func (s *TestSuite) AssertField(value, expected interface{}, fieldName string) {
	if value != expected {
		s.T.Errorf("Field %s: expected %v, got %v", fieldName, expected, value)
	}
}

// AssertContains asserts a string contains a substring.
func (s *TestSuite) AssertContains(str, substr string) {
	if !contains(str, substr) {
		s.T.Errorf("Expected string to contain %q, got %q", substr, str)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IntegrationTest is a helper to create integration tests.
func IntegrationTest(name string, fn func(*TestSuite)) func(*testing.T) {
	return func(t *testing.T) {
		suite := NewTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.TearDown()

		// Check if server is running
		suite.SkipIfNoServer()

		suite.T.Logf("Running integration test: %s", name)
		fn(suite)
	}
}

// BenchmarkSuite is the base for benchmarks.
type BenchmarkSuite struct {
	TB        testing.TB
	Ctx       context.Context
	Client    *fixtures.TestClient
	Container *fixtures.ContainerManager
	Cancel    context.CancelFunc
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(tb testing.TB) *BenchmarkSuite {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	var container *fixtures.ContainerManager

	envDSN := os.Getenv("TEST_DATABASE_DSN")
	if envDSN == "" {
		container = fixtures.NewContainerManager()
		if err := container.Start(ctx); err != nil {
			tb.Skipf("Skipping benchmark: could not start container: %v", err)
			cancel()
			return nil
		}
	}

	client, err := fixtures.NewTestClient(testBaseURL)
	if err != nil {
		tb.Fatalf("Failed to create test client: %v", err)
	}

	return &BenchmarkSuite{
		TB:        tb,
		Ctx:       ctx,
		Client:    client,
		Container: container,
		Cancel:    cancel,
	}
}

// TearDown cleans up after the benchmark.
func (s *BenchmarkSuite) TearDown() {
	if s.Container != nil {
		s.Container.Stop(s.Ctx)
	}
	s.Cancel()
}

// SetupTestServer starts a test server for benchmarks.
// This would be called before running benchmarks.
func SetupTestServer(ctx context.Context, db interface{}) (cleanup func(), err error) {
	// This would start the actual server with test database
	// For now, assume server is already running
	return func() {}, nil
}

// GetServerURL returns the test server URL.
func GetServerURL() string {
	if url := os.Getenv("TEST_SERVER_URL"); url != "" {
		return url
	}
	return testBaseURL
}

// RequireDB skips the test if database is not available.
func RequireDB(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping database test")
	}
}

// RequireServer skips the test if the server is not running.
func RequireServer(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/health")
	if err != nil {
		t.Skipf("Server not running at %s: %v", testBaseURL, err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}
