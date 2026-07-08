package factory

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateHTTPHeadersMap_DuplicateHeadersAllowed verifies the expected behavior.
func TestCreateHTTPHeadersMap_DuplicateHeadersAllowed(t *testing.T) {
	headers, err := createHTTPHeadersMap([]string{
		"X-Test=one",
		"X-Test=two",
	}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two"}, headers["X-Test"])
}

// TestCreateHTTPHeadersMap_NilHeadersAllowed verifies the expected behavior.
func TestCreateHTTPHeadersMap_NilHeadersAllowed(t *testing.T) {
	headers, err := createHTTPHeadersMap(nil, false)
	require.NoError(t, err)
	assert.Equal(t, http.Header{}, headers)
}

// TestCreateHTTPHeadersMap_ResolvableValue verifies the expected behavior.
func TestCreateHTTPHeadersMap_ResolvableValue(t *testing.T) {
	require.NoError(t, os.Setenv("NEVER_TEST_HEADER", "secret"))
	t.Cleanup(func() {
		_ = os.Unsetenv("NEVER_TEST_HEADER")
	})

	headers, err := createHTTPHeadersMap([]string{
		"Authorization=env:NEVER_TEST_HEADER",
	}, false)
	require.NoError(t, err)
	assert.Equal(t, http.Header{"Authorization": []string{"secret"}}, headers)
}
