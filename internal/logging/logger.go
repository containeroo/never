package logging

import (
	"io"
	"log/slog"
)

const (
	// FormatText writes logs in slog text format.
	FormatText string = "text"
	// FormatJSON writes logs in slog JSON format.
	FormatJSON string = "json"
)

// Config controls logger output behavior.
type Config struct {
	Format string
}

// SetupLogger configures the application logger.
func SetupLogger(version string, output io.Writer, configs ...Config) *slog.Logger {
	cfg := Config{Format: FormatText}
	if len(configs) > 0 {
		cfg = configs[0]
	}
	if cfg.Format == "" {
		cfg.Format = FormatText
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{})
	default:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{})
	}

	return slog.New(handler).With(slog.String("version", version))
}
