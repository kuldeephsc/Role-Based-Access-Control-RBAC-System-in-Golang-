package logger

import (
	"log/slog"
	"os"
)

// New returns a structured JSON logger. Every service that wants to log
// takes one of these via constructor injection rather than reaching for a
// global -- keeps logging testable and makes the dependency explicit.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
