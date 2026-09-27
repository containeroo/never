package cli

import (
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	httpExampleURL     = "http://example.com"
	httpWebAddressFlag = "--http.web.address=" + httpExampleURL
)

// TestParseFlagsHTTPAddressDetail verifies address-detail is configured per HTTP target.
func TestParseFlagsHTTPAddressDetail(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		flag string
		want checker.HTTPAddressDetail
	}{
		{name: "default", want: checker.HTTPAddressOrigin},
		{name: "origin", flag: "--http.web.address-detail=origin", want: checker.HTTPAddressOrigin},
		{name: "path", flag: "--http.web.address-detail=path", want: checker.HTTPAddressPath},
		{name: "query", flag: "--http.web.address-detail=query", want: checker.HTTPAddressQuery},
		{name: "full", flag: "--http.web.address-detail=full", want: checker.HTTPAddressFull},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := []string{httpWebAddressFlag}
			if tt.flag != "" {
				args = append(args, tt.flag)
			}

			cfg, err := ParseFlags(args, "1.0.0")
			require.NoError(t, err)
			require.Len(t, cfg.Targets, 1)
			httpConfig, ok := cfg.Targets[0].Config.(checker.HTTPConfig)
			require.True(t, ok)
			assert.Equal(t, tt.want, httpConfig.AddressDetail)
		})
	}

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.address-detail=invalid",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid value for flag --http.web.address-detail")
		assert.ErrorContains(t, err, `"invalid"`)
		assert.ErrorContains(t, err, "must be one of")
		assert.ErrorContains(t, err, "origin")
		assert.ErrorContains(t, err, "path")
		assert.ErrorContains(t, err, "query")
		assert.ErrorContains(t, err, "full")
	})

	t.Run("per target", func(t *testing.T) {
		t.Parallel()

		cfg, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.address-detail=path",
			"--http.api.address=https://api.example.com",
			"--http.api.address-detail=full",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 2)

		apiConfig, ok := cfg.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		webConfig, ok := cfg.Targets[1].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Equal(t, checker.HTTPAddressFull, apiConfig.AddressDetail)
		assert.Equal(t, checker.HTTPAddressPath, webConfig.AddressDetail)
	})
}

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
		cfg, ok := parsedFlags.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Equal(t, http.MethodPost, cfg.Method)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.method=INVALID",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid value for flag --http.web.method")
		assert.ErrorContains(t, err, `"INVALID"`)
		assert.ErrorContains(t, err, "must be one of")
		assert.ErrorContains(t, err, http.MethodGet)
		assert.ErrorContains(t, err, http.MethodPost)
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
		"--http.web.expected-status-codes=200-202,204",
		"--http.web.follow-redirects=false",
		"--http.web.max-redirects=3",
		"--http.web.backoff=exponential",
		"--http.web.max-interval=30s",
		"--http.web.max-attempts=3",
		"--http.web.address-detail=query",
	}

	parsedFlags, err := ParseFlags(args, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)

	target := parsedFlags.Targets[0]
	assert.Equal(t, "web", target.ID)
	assert.Equal(t, "Web", target.Name)
	assert.Equal(t, httpExampleURL, target.Address)
	cfg, ok := target.Config.(checker.HTTPConfig)
	require.True(t, ok)
	assert.Equal(t, http.MethodPost, cfg.Method)
	assert.Equal(t, http.Header{"Authorization": {"Bearer token"}}, cfg.Headers)
	assert.Equal(t, []int{200, 201, 202, 204}, cfg.ExpectedStatusCodes)
	assert.False(t, cfg.FollowRedirects)
	assert.Equal(t, 3, cfg.MaxRedirects)
	assert.Equal(t, "never/1.0.0", cfg.UserAgent)
	assert.Equal(t, checker.HTTPAddressQuery, cfg.AddressDetail)
	assert.Equal(t, 3, target.MaxAttempts)
	assert.Equal(t, backoff.ModeExponential, target.Backoff)
	assert.Equal(t, 30*time.Second, target.MaxInterval)
}

// TestParseFlagsHTTPMaxRedirects verifies redirect limit parsing and validation.
func TestParseFlagsHTTPMaxRedirects(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{httpWebAddressFlag}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg, ok := parsedFlags.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Equal(t, defaultHTTPMaxRedirects, cfg.MaxRedirects)
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.max-redirects=0",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg, ok := parsedFlags.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Zero(t, cfg.MaxRedirects)
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.max-redirects=-1",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "must not be negative")
	})
}

// TestParseFlagsHTTPFollowRedirects verifies redirect following defaults to enabled and can be disabled.
func TestParseFlagsHTTPFollowRedirects(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{httpWebAddressFlag}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg, ok := parsedFlags.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.True(t, cfg.FollowRedirects)
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.follow-redirects=false",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		cfg, ok := parsedFlags.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.False(t, cfg.FollowRedirects)
	})
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
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid value for flag --http.web.backoff")
		assert.ErrorContains(t, err, `"invalid"`)
		assert.ErrorContains(t, err, "must be one of")
		assert.ErrorContains(t, err, backoff.ModeLinear.String())
		assert.ErrorContains(t, err, backoff.ModeExponential.String())
	})
}

// TestParseFlagsHTTPPerTargetMaxAttempts verifies per-target values override or inherit the global limit.
func TestParseFlagsHTTPPerTargetMaxAttempts(t *testing.T) {
	t.Parallel()

	t.Run("inherits global when unset", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			"--max-attempts=5",
			httpWebAddressFlag,
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		assert.Equal(t, 5, parsedFlags.Targets[0].MaxAttempts)
	})

	t.Run("zero disables target limit", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			"--max-attempts=5",
			httpWebAddressFlag,
			"--http.web.max-attempts=0",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		assert.Zero(t, parsedFlags.Targets[0].MaxAttempts)
	})

	t.Run("positive overrides global", func(t *testing.T) {
		t.Parallel()

		parsedFlags, err := ParseFlags([]string{
			"--max-attempts=5",
			httpWebAddressFlag,
			"--http.web.max-attempts=2",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, parsedFlags.Targets, 1)
		assert.Equal(t, 2, parsedFlags.Targets[0].MaxAttempts)
	})
}

// TestParseFlagsHTTPInputParsing verifies HTTP-specific raw values are parsed at the CLI boundary.
func TestParseFlagsHTTPInputParsing(t *testing.T) {
	t.Run("invalid header", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=InvalidHeader",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid HTTP header: invalid header format: InvalidHeader")
	})

	t.Run("duplicate header", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=X-Test=one",
			"--http.web.header=X-Test=two",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "duplicate header key found: X-Test")
	})

	t.Run("duplicate header allowed", func(t *testing.T) {
		t.Parallel()

		cfg, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=X-Test=one",
			"--http.web.header=X-Test=two",
			"--http.web.allow-duplicate-headers=true",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		httpConfig, ok := cfg.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Equal(t, []string{"one", "two"}, httpConfig.Headers.Values("X-Test"))
	})

	t.Run("header value containing comma", func(t *testing.T) {
		t.Parallel()

		cfg, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=Cache-Control=no-cache, no-store",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		httpConfig, ok := cfg.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Equal(t, []string{"no-cache, no-store"}, httpConfig.Headers.Values("Cache-Control"))
	})

	t.Run("header names are case insensitive", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=X-Test=one",
			"--http.web.header=x-test=two",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "duplicate header key found: X-Test")
	})

	t.Run("resolves header values", func(t *testing.T) {
		t.Setenv("NEVER_TEST_HEADER", "secret")

		cfg, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=Authorization=env:NEVER_TEST_HEADER",
		}, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		httpConfig, ok := cfg.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		assert.Equal(t, "secret", httpConfig.Headers.Get("Authorization"))
	})

	t.Run("invalid expected status codes", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.expected-status-codes=299-200",
		}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid status range: 299-200")
	})
}

// TestParseFlagsHTTPRetryValidation verifies retry input is rejected before reaching the factory.
func TestParseFlagsHTTPRetryValidation(t *testing.T) {
	t.Parallel()

	t.Run("negative interval", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{httpWebAddressFlag, "--http.web.interval=-1s"}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "must not be negative")
	})

	t.Run("negative max interval", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{httpWebAddressFlag, "--http.web.max-interval=-1s"}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "must not be negative")
	})

	t.Run("invalid max attempts", func(t *testing.T) {
		t.Parallel()

		_, err := ParseFlags([]string{httpWebAddressFlag, "--http.web.max-attempts=-1"}, "1.0.0")
		require.Error(t, err)
		assert.ErrorContains(t, err, "must not be negative")
	})
}
