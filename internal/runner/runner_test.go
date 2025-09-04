package runner

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/never/internal/flag"
	"github.com/containeroo/never/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake version for testing
const version string = "0.0.0"

func TestRunAllHTTPReady(t *testing.T) {
	t.Parallel()

	// Start a tiny HTTP server on a dedicated port.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Addr: ":18081", Handler: mux}
	go func() { _ = server.ListenAndServe() }()
	defer server.Close() // nolint:errcheck

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=http://localhost:18081",
		"--http.httpcheck.interval=1s",
		"--http.httpcheck.timeout=1s",
	}

	// Build checkers via flags+factory (same path app would use).
	fs, err := flag.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.DynamicGroups, fs.DefaultCheckInterval)
	require.NoError(t, err)

	// Run
	var output strings.Builder
	logger := logging.SetupLogger(version, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, logger)
	assert.NoError(t, err)

	// Assert output contains readiness line
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, lines)
	assert.Contains(t, lines[len(lines)-1], "HTTPServer is ready ✓")
}

func TestRunAllTCPReady(t *testing.T) {
	t.Parallel()

	// Start a TCP listener
	ln, err := net.Listen("tcp", "localhost:18082")
	require.NoError(t, err)
	defer ln.Close() // nolint:errcheck

	args := []string{
		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=localhost:18082",
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	fs, err := flag.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.DynamicGroups, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(version, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, logger)
	assert.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, lines)
	assert.Contains(t, lines[len(lines)-1], "TCPServer is ready ✓")
}

func TestRunAllMultipleReady(t *testing.T) {
	t.Parallel()

	// HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	httpSrv := &http.Server{Addr: ":18085", Handler: mux}
	go func() { _ = httpSrv.ListenAndServe() }()
	defer httpSrv.Close() // nolint:errcheck

	// TCP listener
	ln, err := net.Listen("tcp", "localhost:18086")
	require.NoError(t, err)
	defer ln.Close() // nolint:errcheck

	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=http://localhost:18085",
		"--http.httpcheck.interval=1s",
		"--http.httpcheck.timeout=1s",

		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=localhost:18086",
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	fs, err := flag.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.DynamicGroups, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(version, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = RunAll(ctx, checkers, logger)
	assert.NoError(t, err)

	// Order is nondeterministic; assert both readiness messages appear.
	out := output.String()
	assert.Contains(t, out, "HTTPServer is ready ✓")
	assert.Contains(t, out, "TCPServer is ready ✓")
}

func TestRunAllNoCheckers(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := logging.SetupLogger(version, &output)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := RunAll(ctx, nil, logger)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCheckers), "expected ErrNoCheckers, got %v", err)
	// optional exact message assertion:
	assert.EqualError(t, err, "no checkers to run")
}

func TestRunAllPropagatesError(t *testing.T) {
	t.Parallel()

	// Point to a non-existent HTTP server; set small timeouts.
	args := []string{
		"--http.httpcheck.name=HTTPServer",
		"--http.httpcheck.address=http://localhost:19999",
		"--http.httpcheck.interval=100ms",
		"--http.httpcheck.timeout=100ms",
	}

	fs, err := flag.ParseFlags(args, version)
	require.NoError(t, err)

	checkers, err := factory.BuildCheckers(fs.DynamicGroups, fs.DefaultCheckInterval)
	require.NoError(t, err)

	var output strings.Builder
	logger := logging.SetupLogger(version, &output)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	err = RunAll(ctx, checkers, logger)
	require.Error(t, err)

	// The exact inner error can vary (timeout, context deadline), so check the runner prefix.
	assert.Contains(t, err.Error(), "checker 'HTTPServer' failed")
}
