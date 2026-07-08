package cli

import (
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	httpExampleURL     = "http://example.com"
	httpWebAddressFlag = "--http.web.address=" + httpExampleURL
)

// TestParseFlagsHTTPMethod verifies HTTP method parsing uses enum validation.
func TestParseFlagsHTTPMethod(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.method=POST",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		assert.Equal(t, http.MethodPost, parsedFlags.Targets[0].HTTPMethod)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.method=INVALID",
		}, "1.0.0")
		assertInvalidFlagValueError(t, err, "--http.web.method", "INVALID", http.MethodGet, http.MethodPost)
	})
}

// TestParseFlagsHTTPTarget verifies HTTP flags are converted into typed target config.
func TestParseFlagsHTTPTarget(t *testing.T) {
	t.Parallel()

	args := []string{
		"--default-interval=5s",
		"--http.web.name=Web",
		httpWebAddressFlag,
		"--http.web.method=POST",
		"--http.web.header=Authorization=Bearer token",
		"--http.web.expected-status-codes=200,204",
		"--http.web.backoff=exponential",
		"--http.web.max-interval=30s",
		"--http.web.max-attempts=3",
	}

	parsedFlags, err := ParseFlags(args, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)

	target := parsedFlags.Targets[0]
	assert.Equal(t, "web", target.ID)
	assert.Equal(t, "Web", target.Name)
	assert.Equal(t, httpExampleURL, target.Address)
	assert.Equal(t, http.MethodPost, target.HTTPMethod)
	assert.Equal(t, []string{"Authorization=Bearer token"}, target.HTTPHeaders)
	assert.Equal(t, []string{"200", "204"}, target.HTTPExpectedStatusCodes)
	assert.Equal(t, 3, target.MaxAttempts)
	assert.Equal(t, backoff.ModeExponential, target.Backoff)
	assert.Equal(t, 30*time.Second, target.MaxInterval)
}

// TestParseFlagsHTTPBackoff verifies HTTP backoff values use enum validation.
func TestParseFlagsHTTPBackoff(t *testing.T) {
	t.Parallel()

	t.Run("exponential", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.backoff=exponential",
			"--http.web.max-interval=30s",
		}, "1.0.0")
		require.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.backoff=invalid",
		}, "1.0.0")
		assertInvalidFlagValueError(t, err, "--http.web.backoff", "invalid", backoff.ModeLinear.String(), backoff.ModeExponential.String())
	})
}

// TestParseFlagsHTTPPerTargetMaxAttempts verifies HTTP target max-attempts override parsing.
func TestParseFlagsHTTPPerTargetMaxAttempts(t *testing.T) {
	t.Parallel()

	parsedFlags, err := ParseFlags([]string{
		"--max-attempts=5",
		httpWebAddressFlag,
		"--http.web.max-attempts=2",
	}, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)
	assert.Equal(t, 5, parsedFlags.MaxAttempts)
	assert.Equal(t, 2, parsedFlags.Targets[0].MaxAttempts)
}
