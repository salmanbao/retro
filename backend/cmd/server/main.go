package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"viralforge/backend/src"
	"viralforge/backend/src/adapter"
)

const timeout = 30 * time.Second

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost port=5432 user=postgres password=postgres dbname=solomon sslmode=disable"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		encryptionKey = "default-32-byte-key-for-dev!"
	}

	cfg := &src.Config{
		BaseURL:       baseURL,
		EncryptionKey: encryptionKey,
	}

	// Connect to database
	db, err := adapter.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database")

	store := adapter.NewPostgresStore(db)

	if err := store.SeedData(ctx); err != nil {
		log.Printf("Warning: Failed to seed data (tables may not exist yet): %v", err)
	} else {
		log.Println("Database seeded with initial data")
	}

	// Create email service
	var emailSvc adapter.EmailService
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		emailSvc = adapter.NewSMTPEmailService(adapter.SMTPConfig{
			Host:     smtpHost,
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		})
		log.Println("Using SMTP email service")
	} else {
		emailSvc = &adapter.MockEmailService{}
		log.Println("Using mock email service")
	}

	// Create and start server
	server := src.NewServer(cfg, store, emailSvc)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received")
		cancel()
	}()

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	if err := server.Run(ctx, addr); err != nil {
		log.Printf("Server error: %v", err)
	}
	fmt.Println("Server stopped")
}
