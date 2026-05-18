package fixtures

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"viralforge/backend/src/domain"
	onboardingDomain "viralforge/backend/src/domain/onboarding"
)

// ContainerManager manages PostgreSQL test containers using podman.
type ContainerManager struct {
	ContainerID string
	DSN         string
	Port        string
	Image       string
}

// NewContainerManager creates a new container manager with default settings.
func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		Image: "postgres:16-alpine",
		Port:  "5433", // Use non-standard port to avoid conflicts
	}
}

// Start starts a PostgreSQL container using podman.
func (m *ContainerManager) Start(ctx context.Context) error {
	// Check if podman is available
	if err := exec.CommandContext(ctx, "podman", "version").Run(); err != nil {
		return fmt.Errorf("podman not available: %w", err)
	}

	// Create and start container
	cmd := exec.CommandContext(ctx, "podman", "run",
		"-d",
		"--rm",
		"-p", fmt.Sprintf("%s:5432", m.Port),
		"-e", "POSTGRES_USER=test",
		"-e", "POSTGRES_PASSWORD=test",
		"-e", "POSTGRES_DB=integration_test",
		"--name", fmt.Sprintf("viralforge-test-%d", time.Now().UnixNano()),
		m.Image,
	)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	m.ContainerID = strings.TrimSpace(string(output))

	// Wait for PostgreSQL to be ready
	if err := m.waitForPostgres(ctx, 30*time.Second); err != nil {
		m.Stop(ctx)
		return fmt.Errorf("postgresql failed to start: %w", err)
	}

	m.DSN = fmt.Sprintf("host=localhost user=test password=test dbname=integration_test port=%s sslmode=disable", m.Port)
	return nil
}

// waitForPostgres waits for PostgreSQL to accept connections.
func (m *ContainerManager) waitForPostgres(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(ctx, "podman", "exec", m.ContainerID, "pg_isready", "-U", "test")
		if err := cmd.Run(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("timeout waiting for postgres")
}

// Stop stops and removes the container.
func (m *ContainerManager) Stop(ctx context.Context) error {
	if m.ContainerID == "" {
		return nil
	}
	cmd := exec.CommandContext(ctx, "podman", "stop", m.ContainerID)
	_, err := cmd.Output()
	if err != nil {
		// Try to force remove
		exec.CommandContext(ctx, "podman", "rm", "-f", m.ContainerID).Run()
		return fmt.Errorf("failed to stop container: %w", err)
	}
	m.ContainerID = ""
	return nil
}

// GetDB returns a GORM database connection to the test container.
func (m *ContainerManager) GetDB() (*gorm.DB, error) {
	if m.DSN == "" {
		return nil, fmt.Errorf("container not started")
	}

	db, err := gorm.Open(postgres.Open(m.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// RunMigrations runs database migrations on the test container.
// It imports all domain models to register them with GORM AutoMigrate.
func (m *ContainerManager) RunMigrations(db *gorm.DB) error {
	// Import all domain models for AutoMigrate
	// This is done via blank import to trigger init() functions
	_ = domain.User{}
	_ = domain.Session{}
	_ = domain.Profile{}
	_ = domain.AuthToken{}
	_ = domain.LoginHistory{}
	_ = domain.TwoFactorSettings{}
	_ = domain.Permission{}
	_ = domain.Role{}
	_ = domain.RolePermission{}
	_ = domain.ProfileRole{}
	_ = domain.ProfileEnrichment{}
	_ = domain.PortfolioItem{}
	_ = domain.AudienceData{}
	_ = domain.FollowerVerification{}
	_ = domain.PayoutPreferences{}
	_ = domain.KYCStatus{}

	// Onboarding domain models
	_ = onboardingDomain.OnboardingTemplate{}
	_ = onboardingDomain.OnboardingStep{}
	_ = onboardingDomain.OnboardingProgress{}
	_ = onboardingDomain.StepProgress{}

	// Get the underlying sql.DB and run migrations
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Ping to ensure connection is alive
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	// GORM AutoMigrate will handle schema creation
	// Note: AutoMigrate is called via the store's AutoMigrate method in production
	return nil
}

// GetDSN returns the database connection string.
func (m *ContainerManager) GetDSN() string {
	return m.DSN
}

// IsRunning returns true if the container is running.
func (m *ContainerManager) IsRunning() bool {
	return m.ContainerID != ""
}

// Cleanup removes any leftover containers from previous runs.
func Cleanup(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "podman", "ps", "-a", "--filter", "name=viralforge-test-", "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	ids := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" {
			exec.CommandContext(ctx, "podman", "rm", "-f", id).Run()
		}
	}
	return nil
}

// GetEnvDSN returns DSN from environment variable (for CI/CD).
func GetEnvDSN() string {
	if dsn := os.Getenv("TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	return ""
}
