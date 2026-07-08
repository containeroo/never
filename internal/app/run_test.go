package app

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake version for testing
const version string = "0.0.0"

func TestRunHTTPReady(t *testing.T) {
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

	var stdOut, stdErr strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := Run(ctx, version, args, &stdOut, &stdErr)
	require.NoError(t, err)

	stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
	last := len(stdOutEntries) - 1

	assert.Contains(t, stdOutEntries[last], "HTTPServer is ready ✓")
}

func TestRunTCPReady(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close() // nolint:errcheck

	args := []string{
		"--tcp.tcptest.name=TCPServer",
		"--tcp.tcptest.address=" + listener.Addr().String(),
		"--tcp.tcptest.interval=1s",
		"--tcp.tcptest.timeout=1s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr strings.Builder

	err = Run(ctx, version, args, &stdOut, &stdErr)
	require.NoError(t, err)

	stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
	last := len(stdOutEntries) - 1

	assert.Contains(t, stdOutEntries[last], "TCPServer is ready ✓")
}

func TestRunConfigErrorMissingTarget(t *testing.T) {
	t.Parallel()

	args := []string{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "no checkers to run")
}

func TestRunConfigErrorUnsupportedCheckType(t *testing.T) {
	t.Parallel()

	args := []string{
		"--target.unsupported.name=TestService",
		"--target.unsupported.address=localhost:8080",
		"--target.unsupported.interval=1s",
		"--target.unsupported.timeout=1s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "unknown dynamic group \"target\" in flag --target.unsupported.name=TestService\nunknown dynamic group \"target\" in flag --target.unsupported.address=localhost:8080\nunknown dynamic group \"target\" in flag --target.unsupported.interval=1s\nunknown dynamic group \"target\" in flag --target.unsupported.timeout=1s")
}

func TestRunConfigErrorInvalidHeaders(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--http.invalidheaders.address=http://localhost:8080",
		"--http.invalidheaders.interval=1s",
		"--http.invalidheaders.timeout=1s",
		"--http.invalidheaders.header=InvalidHeader",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "invalid \"--http.invalidheaders.header\": invalid header format: \"InvalidHeader\"")
}

func TestRunParseError(t *testing.T) {
	t.Parallel()

	args := []string{
		"--http.invalidheaders.name=TestService",
		"--invalid",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var stdOut, stdErr bytes.Buffer

	err := Run(ctx, version, args, &stdOut, &stdErr)

	require.Error(t, err)
	assert.EqualError(t, err, "unknown flag --invalid")
}
