package cli

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFlagsTCPTarget verifies TCP flags are converted into typed target config.
func TestParseFlagsTCPTarget(t *testing.T) {
	t.Parallel()

	parsedFlags, err := ParseFlags([]string{
		"--tcp.db.name=Database",
		"--tcp.db.address=example.com:5432",
		"--tcp.db.timeout=3s",
		"--tcp.db.interval=4s",
	}, "1.0.0")
	require.NoError(t, err)
	require.Len(t, parsedFlags.Targets, 1)

	target := parsedFlags.Targets[0]
	assert.Equal(t, checker.TCP, target.Type)
	assert.Equal(t, "db", target.ID)
	assert.Equal(t, "Database", target.Name)
	assert.Equal(t, "example.com:5432", target.Address)
	assert.Equal(t, 3*time.Second, target.TCPTimeout)
	assert.Equal(t, 4*time.Second, target.Interval)
}
