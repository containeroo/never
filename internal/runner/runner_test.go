package runner

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/never/internal/cli"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/logging"
	"github.com/containeroo/never/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake version for testing
const version string = "0.0.0"

const (
	httpServerNameFlag     = "--http.httpcheck.name=HTTPServer"
	httpServerIntervalFlag = "--http.httpcheck.interval=1s"
	httpServerTimeoutFlag  = "--http.httpcheck.timeout=1s"
	httpServerReadyLog     = "HTTPServer is ready ✓"
	tcpServerNameFlag      = "--tcp.tcptest.name=TCPServer"
	tcpServerIntervalFlag  = "--tcp.tcptest.interval=1s"
	tcpServerTimeoutFlag   = "--tcp.tcptest.timeout=1s"
	tcpServerReadyLog      = "TCPServer is ready ✓"
)

func httpAddressFlag(address string) string { return "--http.httpcheck.address=" + address }

func tcpAddressFlag(address string) string { return "--tcp.tcptest.address=" + address }

// TestRunAllHTTPReady verifies RunAll succeeds when an HTTP target is ready.
func TestRunAllHTTPReady(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	args := []string{
		httpServerNameFlag,
		httpAddressFlag(server.URL),
		httpServerIntervalFlag,
		httpServerTimeoutFlag,
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
	assert.Contains(t, lines[len(lines)-1], httpServerReadyLog)
}

// TestRunAllTCPReady verifies RunAll succeeds when a TCP target is ready.
func TestRunAllTCPReady(t *testing.T) {
	t.Parallel()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	args := []string{
		tcpServerNameFlag,
		tcpAddressFlag(listener.Addr().String()),
		tcpServerIntervalFlag,
		tcpServerTimeoutFlag,
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
	assert.Contains(t, lines[len(lines)-1], tcpServerReadyLog)
}

// TestRunAllMultipleReady verifies RunAll handles multiple ready targets.
func TestRunAllMultipleReady(t *testing.T) {
	t.Parallel()

	httpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer httpSrv.Close()

	listener := testutils.ListenLocalTCP(t)
	defer listener.Close() // nolint:errcheck

	args := []string{
		httpServerNameFlag,
		httpAddressFlag(httpSrv.URL),
		httpServerIntervalFlag,
		httpServerTimeoutFlag,

		tcpServerNameFlag,
		tcpAddressFlag(listener.Addr().String()),
		tcpServerIntervalFlag,
		tcpServerTimeoutFlag,
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
	assert.Contains(t, out, httpServerReadyLog)
	assert.Contains(t, out, tcpServerReadyLog)
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
		httpServerNameFlag,
		httpAddressFlag(testutils.LocalHTTPURL(t)),
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
		httpServerNameFlag,
		httpAddressFlag(testutils.LocalHTTPURL(t)),
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
