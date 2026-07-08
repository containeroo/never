package runner

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/never/internal/cli"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake version for testing
const version string = "0.0.0"

// TestRunAllHTTPReady verifies RunAll succeeds when an HTTP target is ready.
func TestRunAllHTTPReady(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=" + server.URL,
		"--http.httpcheck.interval=1s",
		"--http.httpcheck.timeout=1s",
	}

	// Build checkers via flags+factory (same path app would use).
	fs, err := cli.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.Targets, fs.DefaultCheckInterval)
	require.NoError(t, err)

	// Run
	var output strings.Builder
	logger := logging.SetupLogger(logging.LogFormatText, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, -1, logger)
	assert.NoError(t, err)

	// Assert output contains readiness line
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, lines)
	assert.Contains(t, lines[len(lines)-1], "HTTPServer is ready ✓")
}

// TestRunAllTCPReady verifies RunAll succeeds when a TCP target is ready.
func TestRunAllTCPReady(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() // nolint:errcheck

	args := []string{
		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=" + ln.Addr().String(),
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	fs, err := cli.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.Targets, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(logging.LogFormatText, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, -1, logger)
	assert.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, lines)
	assert.Contains(t, lines[len(lines)-1], "TCPServer is ready ✓")
}

// TestRunAllMultipleReady verifies RunAll handles multiple ready targets.
func TestRunAllMultipleReady(t *testing.T) {
	t.Parallel()

	httpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer httpSrv.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close() // nolint:errcheck

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=" + httpSrv.URL,
		"--http.httpcheck.interval=1s",
		"--http.httpcheck.timeout=1s",

		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=" + ln.Addr().String(),
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	fs, err := cli.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.Targets, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(logging.LogFormatText, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, -1, logger)
	assert.NoError(t, err)

	// Order is nondeterministic; assert both readiness messages appear.
	out := output.String()
	assert.Contains(t, out, "HTTPServer is ready ✓")
	assert.Contains(t, out, "TCPServer is ready ✓")
}

// TestRunAllNoCheckers verifies RunAll rejects an empty checker list.
func TestRunAllNoCheckers(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := logging.SetupLogger(logging.LogFormatText, &output)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := RunAll(ctx, nil, -1, logger)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCheckers), "expected ErrNoCheckers, got %v", err)
	assert.EqualError(t, err, "no checkers to run")
}

// TestRunAllPropagatesError verifies checker errors are wrapped with target context.
func TestRunAllPropagatesError(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=http://" + unusedTCPAddr(t),
		"--http.httpcheck.interval=100ms",
		"--http.httpcheck.timeout=100ms",
	}

	fs, err := cli.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.Targets, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(logging.LogFormatText, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	err = RunAll(ctx, checkers, -1, logger)
	require.Error(t, err)

	// The exact inner error can vary (timeout, context deadline), so check the runner prefix.
	assert.Contains(t, err.Error(), "checker 'HTTPServer' failed")
}

// TestRunAllMaxAttempts verifies max-attempts errors propagate through RunAll.
func TestRunAllMaxAttempts(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=http://" + unusedTCPAddr(t),
		"--http.httpcheck.interval=50ms",
		"--http.httpcheck.timeout=50ms",
	}

	fs, err := cli.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.Targets, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(logging.LogFormatText, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, 2, logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checker 'HTTPServer' failed")
}

// unusedTCPAddr returns an unused local TCP address for tests.
func unusedTCPAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	return addr
}
