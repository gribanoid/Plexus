package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Configure sets the process-wide JSON slog default with level and service attribute.
func Configure(service, level string) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	}).WithAttrs([]slog.Attr{
		slog.String("service", service),
	})))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
