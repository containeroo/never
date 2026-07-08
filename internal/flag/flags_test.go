package flag

import (
	"errors"
	"testing"
	"time"

	"github.com/containeroo/tinyflags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()

	t.Run("Successful Parsing", func(t *testing.T) {
		t.Parallel()

		args := []string{"--default-interval=5s"}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, 5*time.Second, parsedFlags.DefaultCheckInterval)
	})

	t.Run("Handle Help Flag", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--help"}, "1.0.0")
		require.Error(t, err)
	})

	t.Run("Show Version Flag", func(t *testing.T) {
		t.Parallel()

		args := []string{"--version"}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)
		var verr *tinyflags.VersionRequested
		assert.True(t, errors.As(err, &verr), "expected VersionRequested error")
		assert.EqualError(t, err, "1.0.0")
	})

	t.Run("Invalid Duration Flag", func(t *testing.T) {
		t.Parallel()

		args := []string{"--default-interval=invalid"}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)

		assert.EqualError(t, err, "invalid value for flag --default-interval: time: invalid duration \"invalid\"")
	})

	t.Run("HTTP Method Invalid", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--http.web.address=http://example.com",
			"--http.web.method=INVALID",
		}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "INVALID")
	})

	t.Run("ICMP Hostname Allowed", func(t *testing.T) {
		t.Parallel()

		args := []string{"--icmp.host.address=example.com"}
		_, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
	})

	t.Run("ICMP Timeout Allowed", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--icmp.host.address=example.com",
			"--icmp.host.timeout=3s",
		}
		_, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
	})

	t.Run("Log Format JSON", func(t *testing.T) {
		t.Parallel()

		args := []string{"--log-format=json"}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "json", parsedFlags.LogFormat)
	})

	t.Run("Log Format Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--log-format=xml"}, "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "xml")
	})

	t.Run("Log Level Warn", func(t *testing.T) {
		t.Parallel()

		args := []string{"--log-level=warn"}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "warn", parsedFlags.LogLevel)
	})

	t.Run("Log Level Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{"--log-level=trace"}, "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "trace")
	})

	t.Run("HTTP Backoff Exponential", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--http.web.address=http://example.com",
			"--http.web.backoff=exponential",
			"--http.web.max-interval=30s",
		}

		_, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
	})

	t.Run("HTTP Backoff Invalid", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--http.web.address=http://example.com",
			"--http.web.backoff=linear",
		}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "linear")
	})

	t.Run("Max Attempts Valid", func(t *testing.T) {
		t.Parallel()

		args := []string{"--max-attempts=3"}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, 3, parsedFlags.MaxAttempts)
	})

	t.Run("Max Attempts Endless", func(t *testing.T) {
		t.Parallel()

		args := []string{"--max-attempts=-1"}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, -1, parsedFlags.MaxAttempts)
	})

	t.Run("Max Attempts Invalid Zero", func(t *testing.T) {
		t.Parallel()

		args := []string{"--max-attempts=0"}

		_, err := ParseFlags(args, "1.0.0")
		require.Error(t, err)
	})

	t.Run("Per-Target Max Attempts", func(t *testing.T) {
		t.Parallel()

		args := []string{
			"--max-attempts=5",
			"--http.web.address=http://example.com",
			"--http.web.max-attempts=2",
		}

		parsedFlags, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, 5, parsedFlags.MaxAttempts)
	})
}
