package viralforge

import (
	"os"
)

// Config holds environment configuration for the backend service.
type Config struct {
	DatabaseURL  string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	BaseURL      string
	TokenSecret  string
}

// Load returns a Config populated from environment variables.
// Missing required variables cause a fatal error at startup.
func Load() *Config {
	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@viralforge.com"),
		BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
		TokenSecret:  getEnv("TOKEN_SECRET", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}