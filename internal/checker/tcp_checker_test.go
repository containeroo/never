package checker

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTCPChecker_Valid(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:7080")
	if err != nil {
		t.Fatalf("failed to start TCP server: %q", err)
	}
	defer ln.Close() // nolint:errcheck

	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	assert.Equal(t, checker.Name(), "example")
	assert.Equal(t, checker.Address(), ln.Addr().String())
	assert.Equal(t, checker.Type(), TCP.String())
}

func TestTCPChecker_ValidConnection(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:7081")
	if err != nil {
		t.Fatalf("failed to start TCP server: %q", err)
	}
	defer ln.Close() // nolint:errcheck

	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)
	require.NoError(t, err)
}

func TestTCPChecker_FailedConnection(t *testing.T) {
	t.Parallel()

	checker, err := newTCPChecker("example", "127.0.0.1:7090", WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)

	require.Error(t, err)
	assert.EqualError(t, err, "dial tcp 127.0.0.1:7090: connect: connection refused")
}

func TestTCPChecker_InvalidAddress(t *testing.T) {
	t.Parallel()

	checker, err := newTCPChecker("example", "invalid-address", WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)

	require.Error(t, err)
	assert.EqualError(t, err, "dial tcp: address invalid-address: missing port in address")
}

func TestTCPChecker_Timeout(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:7082")
	defer ln.Close() // nolint:errcheck,staticcheck
	require.NoError(t, err)

	// Simulate a timeout by setting an impossibly short timeout
	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Nanosecond))
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)

	require.Error(t, err)
	assert.EqualError(t, err, "dial tcp 127.0.0.1:7082: i/o timeout")
}
