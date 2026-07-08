package factory_test

import (
	"testing"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildCheckers verifies the expected behavior.
func TestBuildCheckers(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:                        "mygroup",
				Type:                      checker.HTTP,
				Name:                      "mygroup",
				Address:                   "http://example.com",
				Interval:                  5 * time.Second,
				MaxAttempts:               3,
				HTTPMethod:                "GET",
				HTTPHeaders:               []string{"Content-Type=application/json"},
				HTTPAllowDuplicateHeaders: true,
				HTTPExpectedStatusCodes:   []string{"200"},
				HTTPSkipTLSVerify:         true,
				HTTPTimeout:               33 * time.Second,
			},
		}, 9*time.Second)

		require.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "http://example.com", checkers[0].Checker.Address())
		assert.Equal(t, 5*time.Second, checkers[0].Interval)
		assert.Equal(t, 3, checkers[0].MaxAttempts)
	})

	t.Run("Invalid Check Type", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      "mygroup",
				Type:    checker.CheckType("invalid"),
				Address: "invalid-address",
			},
		}, 2*time.Second)

		assert.Nil(t, checkers)
		assert.EqualError(t, err, "unsupported check type: invalid")
	})

	t.Run("Invalid Header Parsing", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          "mygroup",
				Type:        checker.HTTP,
				Address:     "http://example.com",
				HTTPMethod:  "GET",
				HTTPHeaders: []string{"InvalidHeaderFormat"},
			},
		}, 2*time.Second)

		require.Error(t, err)
		assert.EqualError(t, err, "invalid \"--http.mygroup.header\": invalid header format: \"InvalidHeaderFormat\"")
		assert.Nil(t, checkers)
		assert.ErrorContains(t, err, "invalid \"--http.mygroup.header\"")
	})

	t.Run("Inalid HTTP Status codes", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:                      "myid",
				Type:                    checker.HTTP,
				Address:                 "http://example.com",
				HTTPMethod:              "GET",
				HTTPExpectedStatusCodes: []string{"201-200"},
			},
		}, 2*time.Second)

		require.Error(t, err)
		assert.Len(t, checkers, 0)
	})

	t.Run("Valid HTTP Status codes", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:                      "mygroup",
				Type:                    checker.HTTP,
				Address:                 "http://example.com",
				HTTPMethod:              "GET",
				HTTPExpectedStatusCodes: []string{"200,201"},
			},
		}, 2*time.Second)

		require.NoError(t, err)
		assert.Len(t, checkers, 1)
	})

	t.Run("HTTP Backoff", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:          "mygroup",
				Type:        checker.HTTP,
				Address:     "http://example.com",
				HTTPMethod:  "GET",
				Backoff:     backoff.ModeExponential,
				MaxInterval: 30 * time.Second,
			},
		}, 2*time.Second)

		require.NoError(t, err)
		require.Len(t, checkers, 1)
		assert.Equal(t, backoff.ModeExponential, checkers[0].Backoff)
		assert.Equal(t, 30*time.Second, checkers[0].MaxInterval)
	})

	t.Run("Valid TCP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:         "mygroup",
				Type:       checker.TCP,
				Address:    "127.0.0.1:8080",
				TCPTimeout: 3 * time.Second,
			},
		}, 2*time.Second)

		require.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "127.0.0.1:8080", checkers[0].Checker.Address())
	})

	t.Run("Valid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:               "mygroup",
				Type:             checker.ICMP,
				Address:          "8.8.8.8",
				ICMPTimeout:      2 * time.Second,
				ICMPReadTimeout:  2 * time.Second,
				ICMPWriteTimeout: 2 * time.Second,
			},
		}, 2*time.Second)

		require.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "8.8.8.8", checkers[0].Checker.Address())
	})

	t.Run("Invalid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		checkers, err := factory.BuildCheckers([]factory.TargetConfig{
			{
				ID:      "mygroup",
				Type:    checker.ICMP,
				Address: "://invalid-url",
			},
		}, 2*time.Second)

		assert.Nil(t, checkers)
		require.Error(t, err)
	})
}
