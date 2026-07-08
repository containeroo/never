package cli

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFlagsICMPHostname verifies hostnames are valid ICMP target addresses.
func TestParseFlagsICMPHostname(t *testing.T) {
	t.Parallel()

	parsedFlags, err := ParseFlags([]string{"--icmp.host.address=example.com"}, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)
	assert.Equal(t, checker.ICMP, parsedFlags.Targets[0].Type)
	assert.Equal(t, "example.com", parsedFlags.Targets[0].Address)
}

// TestParseFlagsICMPTimeout verifies ICMP timeout flags are converted into target config.
func TestParseFlagsICMPTimeout(t *testing.T) {
	t.Parallel()

	parsedFlags, err := ParseFlags([]string{
		"--icmp.host.address=example.com",
		"--icmp.host.timeout=3s",
		"--icmp.host.read-timeout=4s",
		"--icmp.host.write-timeout=5s",
	}, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)
	assert.Equal(t, 3*time.Second, parsedFlags.Targets[0].ICMPTimeout)
	assert.Equal(t, 4*time.Second, parsedFlags.Targets[0].ICMPReadTimeout)
	assert.Equal(t, 5*time.Second, parsedFlags.Targets[0].ICMPWriteTimeout)
}
