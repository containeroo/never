package logging

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogger(t *testing.T) {
	t.Parallel()

	// Test that the logger includes the version and outputs correctly.
	t.Run("Logger includes version and writes to output", func(t *testing.T) {
		t.Parallel()

		var output strings.Builder
		version := "1.0.0"

		logger := SetupLogger(version, &output)
		assert.NotNil(t, logger)

		logger.Info("Test log message")
		logOutput := output.String()
		assert.Contains(t, logOutput, "Test log message")
		assert.Contains(t, logOutput, "version=1.0.0")
	})

	// Test that the logger writes output to the correct writer.
	t.Run("Logger writes to specified output", func(t *testing.T) {
		t.Parallel()

		var output strings.Builder
		version := "2.0.0"

		logger := SetupLogger(version, &output)
		assert.NotNil(t, logger)

		logger.Warn("This is a warning")

		logOutput := output.String()
		assert.Contains(t, logOutput, "This is a warning")
	})
}

func TestSetupLoggerJSONFormat(t *testing.T) {
	t.Parallel()

	var output strings.Builder
	logger := SetupLogger("1.0.0", &output, Config{Format: FormatJSON})

	logger.Info("json log message")

	logOutput := output.String()
	assert.Contains(t, logOutput, `"msg":"json log message"`)
	assert.Contains(t, logOutput, `"version":"1.0.0"`)
}
