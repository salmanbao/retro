package main

import (
	"context"
	"log"
	"os"

	"viralforge/backend/src/adapter"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost port=5432 user=postgres password=postgres dbname=viralforge_test sslmode=disable"
	}

	ctx := context.Background()

	db, err := adapter.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database")

	store := adapter.NewPostgresStore(db)

	// Run migrations - for a fresh database, AutoMigrate works
	log.Println("Running database migrations...")
	if err := store.AutoMigrate(); err != nil {
		log.Printf("Warning: AutoMigrate issue: %v", err)
	}
	log.Println("Database migration complete")

	// Seed initial data
	if err := store.SeedData(ctx); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	} else {
		log.Println("Database seeded with initial data")
	}

	log.Println("Database setup complete!")
}
