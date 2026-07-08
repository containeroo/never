package checker

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTCPChecker_Valid(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() // nolint:errcheck

	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	assert.Equal(t, checker.Name(), "example")
	assert.Equal(t, checker.Address(), ln.Addr().String())
	assert.Equal(t, checker.Type(), TCP.String())
}

func TestTCPChecker_ValidConnection(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() // nolint:errcheck

	checker, err := newTCPChecker("example", ln.Addr().String(), WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)
	require.NoError(t, err)
}

func TestTCPChecker_FailedConnection(t *testing.T) {
	t.Parallel()

	checker, err := newTCPChecker("example", unusedTCPAddr(t), WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	ctx := context.Background()
	err = checker.Check(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
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

	checker, err := newTCPChecker("example", unusedTCPAddr(t), WithTCPTimeout(1*time.Second))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond)

	err = checker.Check(ctx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expected context deadline exceeded, got %v", err)
}

func unusedTCPAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	return addr
}
