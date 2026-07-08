package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertInvalidFlagValueError verifies an enum parse error names the flag, rejected value, and allowed values.
func assertInvalidFlagValueError(t *testing.T, err error, flag string, value string, allowed ...string) {
	t.Helper()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value for flag "+flag)
	assert.Contains(t, err.Error(), "\""+value+"\"")
	assert.Contains(t, err.Error(), "must be one of")
	for _, allowedValue := range allowed {
		assert.Contains(t, err.Error(), allowedValue)
	}
}
