package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetupLogger verifies logger setup for supported and fallback formats.
func TestSetupLogger(t *testing.T) {
	t.Parallel()

	t.Run("json logger writes json", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatJSON, &buf)

		logger.Info("hello", "key", "value")

		assert.Contains(t, buf.String(), `"msg":"hello"`)
		assert.Contains(t, buf.String(), `"key":"value"`)
	})

	t.Run("text logger writes text", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormatText, &buf)

		logger.Info("hello", "key", "value")

		assert.Contains(t, buf.String(), "msg=hello")
		assert.Contains(t, buf.String(), "key=value")
	})

	t.Run("invalid format falls back to json", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		logger := SetupLogger(LogFormat("invalid"), &buf)

		logger.LogAttrs(context.Background(), slog.LevelInfo, "fallback")

		assert.Contains(t, buf.String(), `"msg":"fallback"`)
	})
}
