package logger

import (
	"io"
	"log/slog"
)

const (
	EnvDev   = "dev"
	EnvLocal = "local"
	EnvProd  = "prod"
)

// NewLogger creates a new structured logger based on the environment.
func NewLogger(env string, w io.Writer) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case EnvDev:
		logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case EnvLocal:
		logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case EnvProd:
		logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return logger
}
