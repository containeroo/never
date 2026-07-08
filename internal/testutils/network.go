package testutils

import (
	"net"
	"testing"
)

const (
	// LocalhostIPv4 is the deterministic IPv4 loopback address used in tests.
	LocalhostIPv4 = "127.0.0.1"
	// LocalhostIPv6 is the deterministic IPv6 loopback address used in tests.
	LocalhostIPv6 = "::1"
)

// LocalhostAddr returns a host:port address on the IPv4 loopback interface.
func LocalhostAddr(port string) string {
	return net.JoinHostPort(LocalhostIPv4, port)
}

// ListenLocalTCP opens a local TCP listener on an unused port.
// The caller owns the listener and must close it.
func ListenLocalTCP(t testing.TB) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", LocalhostAddr("0"))
	if err != nil {
		t.Fatalf("listen local TCP: %v", err)
	}

	return listener
}

// LocalTCPAddr returns the address of a local TCP listener on an unused port.
// The listener is closed before returning, so this should only be used for
// tests that need an address that is expected to refuse connections.
func LocalTCPAddr(t testing.TB) string {
	t.Helper()

	listener := ListenLocalTCP(t)
	addr := listener.Addr().String()

	if err := listener.Close(); err != nil {
		t.Fatalf("close local TCP listener: %v", err)
	}

	return addr
}

// LocalHTTPURL returns an HTTP URL using a local TCP address that should refuse connections.
func LocalHTTPURL(t testing.TB) string {
	t.Helper()

	return "http://" + LocalTCPAddr(t)
}
