package factory_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	targetID        = "mygroup"
	testHTTPAddress = "http://example.com"
)

// TestBuildCheckers verifies the expected behavior.
func TestBuildCheckers(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Name:        targetID,
				Address:     testHTTPAddress,
				Interval:    5 * time.Second,
				MaxAttempts: 3,
				Config: checker.HTTPConfig{
					Method:              http.MethodGet,
					Headers:             http.Header{"Content-Type": {"application/json"}},
					ExpectedStatusCodes: []int{http.StatusOK},
					FollowRedirects:     true,
					MaxRedirects:        4,
					SkipTLSVerify:       true,
					Timeout:             33 * time.Second,
					AddressDetail:       checker.HTTPAddressOrigin,
				},
			},
		}, 9*time.Second)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, testHTTPAddress, checkers[0].Checker.Address())
		assert.Equal(t, 5*time.Second, checkers[0].Interval)
		assert.Equal(t, 3, checkers[0].MaxAttempts)
	})

	t.Run("Rejects missing checker config", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{ID: targetID, Address: testHTTPAddress},
		}, 2*time.Second)

		assert.Nil(t, checkers)
		assert.EqualError(t, err, `target "mygroup" has no checker config`)
	})

	t.Run("HTTP Backoff", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Address:     testHTTPAddress,
				Backoff:     backoff.ModeExponential,
				MaxInterval: 30 * time.Second,
				Config:      checker.DefaultHTTPConfig(),
			},
		}, 2*time.Second)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, backoff.ModeExponential, checkers[0].Backoff)
		assert.Equal(t, 30*time.Second, checkers[0].MaxInterval)
	})

	t.Run("Valid TCP Checker", func(t *testing.T) {
		t.Parallel()

		address := testutils.LocalhostAddr("8080")
		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: address,
				Config:  checker.TCPConfig{Timeout: 3 * time.Second},
			},
		}, 2*time.Second)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, address, checkers[0].Checker.Address())
	})

	t.Run("Valid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      targetID,
				Address: testutils.LocalhostIPv4,
				Config: checker.ICMPConfig{
					ReadTimeout:  2 * time.Second,
					WriteTimeout: 2 * time.Second,
				},
			},
		}, 2*time.Second)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, testutils.LocalhostIPv4, checkers[0].Checker.Address())
	})
}

func TestBuildCheckersDefaultsNameToID(t *testing.T) {
	t.Parallel()

	checkers, err := factory.BuildCheckers([]factory.TargetConfig{
		{
			ID:      "database",
			Address: testutils.LocalhostAddr("5432"),
			Config:  checker.DefaultTCPConfig(),
		},
	}, time.Second)

	require.NoError(t, err)
	require.Len(t, checkers, 1)
	assert.Equal(t, "database", checkers[0].Checker.Name())
}

// TestBuildCheckersRetrySettings verifies the factory preserves already-resolved retry settings.
func TestBuildCheckersRetrySettings(t *testing.T) {
	t.Parallel()

	t.Run("endless retries", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Address:     testHTTPAddress,
				MaxAttempts: 0,
				Config:      checker.DefaultHTTPConfig(),
			},
		}, 3*time.Second)
		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, 3*time.Second, checkers[0].Interval)
		assert.Zero(t, checkers[0].MaxAttempts)
		assert.Equal(t, backoff.ModeLinear, checkers[0].Backoff)
		assert.Zero(t, checkers[0].MaxInterval)
	})

	t.Run("finite retries", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          targetID,
				Address:     testHTTPAddress,
				Interval:    2 * time.Second,
				MaxAttempts: 3,
				Backoff:     backoff.ModeExponential,
				MaxInterval: 30 * time.Second,
				Config:      checker.DefaultHTTPConfig(),
			},
		}, 3*time.Second)
		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, 2*time.Second, checkers[0].Interval)
		assert.Equal(t, 3, checkers[0].MaxAttempts)
		assert.Equal(t, backoff.ModeExponential, checkers[0].Backoff)
		assert.Equal(t, 30*time.Second, checkers[0].MaxInterval)
	})
}
