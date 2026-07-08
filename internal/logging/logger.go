package logging

import (
	"io"
	"log/slog"
)

// LogFormat defines the supported log formats.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// SetupLogger configures a structured logger.
func SetupLogger(logFormat LogFormat, output io.Writer) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}

	var handler slog.Handler
	switch logFormat {
	case LogFormatJSON:
		handler = slog.NewJSONHandler(output, handlerOpts)
	case LogFormatText:
		handler = slog.NewTextHandler(output, handlerOpts)
	default:
		// Default to JSON if an invalid format is provided.
		handler = slog.NewJSONHandler(output, handlerOpts)
	}

	return slog.New(handler)
}
