// Package logger настраивает структурированный JSON-логгер на log/slog.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New возвращает JSON-логгер с полем service. level: debug, info, warn, error.
func New(service, level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h).With("service", service)
}
