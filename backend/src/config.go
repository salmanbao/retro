package src

import (
	"os"
)

// Config holds environment configuration for the backend service.
type Config struct {
	DatabaseURL   string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
	FromEmail     string
	BaseURL       string
	TokenSecret   string
	EncryptionKey string // 32 bytes for AES-256 encryption
}

// Load returns a Config populated from environment variables.
func Load() *Config {
	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		SMTPHost:      getEnv("SMTP_HOST", "localhost"),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		FromEmail:     getEnv("FROM_EMAIL", "noreply@viralforge.com"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		TokenSecret:   os.Getenv("TOKEN_SECRET"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
