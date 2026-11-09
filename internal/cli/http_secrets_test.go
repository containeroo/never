package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/containeroo/never/internal/checker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPSecretHeaders verifies secret references reach requests through both CLI and environment configuration.
func TestHTTPSecretHeaders(t *testing.T) {
	received := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	checkHeaders := func(t *testing.T, args []string) {
		t.Helper()
		cfg, err := ParseFlags(args, "1.0.0")
		require.NoError(t, err)
		require.Len(t, cfg.Targets, 1)
		httpConfig, ok := cfg.Targets[0].Config.(checker.HTTPConfig)
		require.True(t, ok)
		check, err := httpConfig.NewChecker(cfg.Targets[0].Name, cfg.Targets[0].Address)
		require.NoError(t, err)
		require.NoError(t, check.Check(context.Background()))
		headers := <-received
		assert.Equal(t, "Bearer secret=token,with-comma", headers.Get("Authorization"))
		assert.Equal(t, "plain", headers.Get("X-Public"))
	}

	t.Run("environment secret", func(t *testing.T) {
		t.Setenv("NEVER_TEST_AUTHORIZATION", "Bearer secret=token,with-comma")
		checkHeaders(t, []string{
			"--http.web.address=" + server.URL,
			"--http.web.header=Authorization=env:NEVER_TEST_AUTHORIZATION",
			"--http.web.header=X-Public=plain",
		})
	})

	t.Run("mounted secret", func(t *testing.T) {
		secretFile := filepath.Join(t.TempDir(), "authorization")
		require.NoError(t, os.WriteFile(secretFile, []byte("Bearer secret=token,with-comma\n"), 0o600))
		checkHeaders(t, []string{
			"--http.web.address=" + server.URL,
			"--http.web.header=Authorization=file:" + secretFile,
			"--http.web.header=X-Public=plain",
		})
	})

	t.Run("key-value secret", func(t *testing.T) {
		keyFile := filepath.Join(t.TempDir(), "secrets.env")
		require.NoError(t, os.WriteFile(keyFile, []byte("AUTHORIZATION=Bearer secret=token,with-comma\n"), 0o600))
		checkHeaders(t, []string{
			"--http.web.address=" + server.URL,
			"--http.web.header=Authorization=file:" + keyFile + "//AUTHORIZATION",
			"--http.web.header=X-Public=plain",
		})
	})

	t.Run("environment configuration", func(t *testing.T) {
		t.Setenv("NEVER_TEST_AUTHORIZATION", "Bearer secret=token,with-comma")
		t.Setenv("NEVER__HTTP_WEB_ADDRESS", server.URL)
		t.Setenv("NEVER__HTTP_WEB_HEADER", "Authorization=env:NEVER_TEST_AUTHORIZATION\nX-Public=plain")
		checkHeaders(t, nil)
	})
}

// TestHTTPSecretHeaderResolutionFailure verifies unresolved secrets fail before checks start.
func TestHTTPSecretHeaderResolutionFailure(t *testing.T) {
	t.Run("missing environment variable", func(t *testing.T) {
		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=Authorization=env:NEVER_TEST_MISSING_SECRET",
		}, "1.0.0")
		require.ErrorContains(t, err, "invalid HTTP header: failed to resolve variable in header")
	})

	t.Run("missing secret file", func(t *testing.T) {
		_, err := ParseFlags([]string{
			httpWebAddressFlag,
			"--http.web.header=Authorization=file:" + filepath.Join(t.TempDir(), "missing"),
		}, "1.0.0")
		require.ErrorContains(t, err, "invalid HTTP header: failed to resolve variable in header")
	})
}
