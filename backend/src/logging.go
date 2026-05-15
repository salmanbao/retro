package src

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger creates a structured logger for the application.
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// NewNoOpLogger returns a logger that discards all output.
func NewNoOpLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{}))
}
