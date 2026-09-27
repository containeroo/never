package cli

import (
	"errors"
	"testing"
	"time"

	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFlagsDefaultInterval verifies the global default interval flag is parsed.
func TestParseFlagsDefaultInterval(t *testing.T) {
	t.Parallel()

	parsedFlags, err := ParseFlags([]string{"--default-interval=5s"}, "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, parsedFlags.DefaultCheckInterval)
}

// TestParseFlagsDefaultIntervalRejectsNegative verifies retry validation stays at the CLI boundary.
func TestParseFlagsDefaultIntervalRejectsNegative(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--default-interval=-1s"}, "1.0.0")
	require.Error(t, err)
	assert.ErrorContains(t, err, "must not be negative")
}

// TestParseFlagsHelp verifies help requests are returned as tinyflags errors.
func TestParseFlagsHelp(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--help"}, "1.0.0")
	require.Error(t, err)
}

// TestParseFlagsVersion verifies version requests include the configured version.
func TestParseFlagsVersion(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--version"}, "1.0.0")
	require.Error(t, err)

	_, ok := errors.AsType[*tinyflags.VersionRequested](err)
	assert.True(t, ok, "expected VersionRequested error")
	assert.EqualError(t, err, "1.0.0")
}

// TestParseFlagsInvalidDuration verifies invalid duration values are rejected.
func TestParseFlagsInvalidDuration(t *testing.T) {
	t.Parallel()

	_, err := ParseFlags([]string{"--default-interval=invalid"}, "1.0.0")
	require.Error(t, err)
	assert.EqualError(t, err, "invalid value for flag --default-interval: time: invalid duration \"invalid\"")
}

// TestParseFlagsMaxAttempts verifies global max-attempts parsing and validation.
func TestParseFlagsMaxAttempts(t *testing.T) {
	t.Parallel()

	t.Run("default is endless", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags(nil, "1.0.0")
		require.NoError(t, err)
		assert.Zero(t, parsedFlags.MaxAttempts)
	})

	t.Run("positive", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{"--max-attempts=3"}, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, 3, parsedFlags.MaxAttempts)
	})

	t.Run("zero is endless", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{"--max-attempts=0"}, "1.0.0")
		require.NoError(t, err)
		assert.Zero(t, parsedFlags.MaxAttempts)
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--max-attempts=-1"}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "must not be negative")
	})
}

// TestParseFlagsLogFormat verifies log format parsing uses the configured enum values.
func TestParseFlagsLogFormat(t *testing.T) {
	t.Parallel()

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{"--log-format=json"}, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, logging.LogFormatJSON, parsedFlags.LogFormat)
	})

	t.Run("text", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{"--log-format=text"}, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, logging.LogFormatText, parsedFlags.LogFormat)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--log-format=xml"}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid value for flag --log-format")
		assert.ErrorContains(t, err, `"xml"`)
		assert.ErrorContains(t, err, "must be one of")
		assert.ErrorContains(t, err, "json")
		assert.ErrorContains(t, err, "text")
	})
}
