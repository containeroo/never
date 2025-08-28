package checker

import (
	"context"
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/containeroo/never/internal/testutils"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func isPermissionError(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		return false
	}
	// Best-effort cross-platform check
	// net.OpError -> os.SyscallError -> EPERM/EACCES
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "operation not permitted") || strings.Contains(s, "permission denied")
}

func TestNewProtocol(t *testing.T) {
	t.Parallel()

	t.Run("Valid IPv4 Address", func(t *testing.T) {
		t.Parallel()

		protocol, err := newProtocol("192.168.1.1")
		require.NoError(t, err)

		if _, ok := protocol.(*ICMPv4); !ok {
			t.Fatalf("expected ICMPv4 protocol, got %T", protocol)
		}
	})

	t.Run("Valid IPv6 Address", func(t *testing.T) {
		t.Parallel()

		protocol, err := newProtocol("2001:db8::1")
		require.NoError(t, err)

		if _, ok := protocol.(*ICMPv6); !ok {
			t.Fatalf("expected ICMPv6 protocol, got %T", protocol)
		}
	})

	t.Run("Unresolvable Address", func(t *testing.T) {
		t.Parallel()

		_, err := newProtocol("invalid.domain")

		require.Error(t, err)
		assert.EqualError(t, err, "invalid or unresolvable address: invalid.domain")
	})

	t.Run("Unsupported IP Address", func(t *testing.T) {
		t.Parallel()

		_, err := newProtocol("300.300.300.300")

		require.Error(t, err)
		assert.EqualError(t, err, "invalid or unresolvable address: 300.300.300.300")
	})
}

func TestICMPv4MakeRequest(t *testing.T) {
	t.Parallel()

	t.Run("MakeRequest", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		msg, err := protocol.MakeRequest(1234, 1)

		require.NoError(t, err)
		assert.Len(t, msg, 23)
	})
}

func TestICMPv4_Network(t *testing.T) {
	t.Parallel()

	t.Run("Network", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		assert.Equal(t, protocol.Network(), "ip4:icmp")
	})
}

func TestICMPv4_SetDeadline(t *testing.T) {
	t.Parallel()

	t.Run("SetDeadline Success", func(t *testing.T) {
		t.Parallel()

		mockConn := testutils.MockPacketConn{}
		protocol := &ICMPv4{conn: &mockConn}
		err := protocol.SetDeadline(time.Now().Add(1 * time.Second))

		require.NoError(t, err)
	})

	t.Run("SetDeadline Error", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{conn: nil}
		err := protocol.SetDeadline(time.Now().Add(1 * time.Second))

		require.Error(t, err)
	})
}

func TestICMPv4_ValidateReply(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected Message Type", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a reply with a different identifier
		request[4] = 0xFF

		err := protocol.ValidateReply(request, 1234, 1)

		require.Error(t, err)
		assert.EqualError(t, err, "unexpected ICMPv4 message type: echo")
	})

	t.Run("ValidateReply Success", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a successful reply by modifying the request type to EchoReply
		reply := request
		reply[0] = byte(ipv4.ICMPTypeEchoReply)

		err := protocol.ValidateReply(reply, 1234, 1)
		require.NoError(t, err)
	})
	t.Run("ValidateReply Identifier Mismatch", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a reply with a different identifier
		reply := request
		reply[4] = 0xFF // Modify the identifier

		err := protocol.ValidateReply(reply, 1234, 1)

		require.Error(t, err)
		assert.EqualError(t, err, "unexpected ICMPv4 message type: echo")
	})

	t.Run("Error Parsing Message", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		// Pass an invalid byte slice that cannot be parsed as a valid ICMP message
		reply := []byte{0xff, 0xff, 0xff}

		err := protocol.ValidateReply(reply, 1234, 1)
		require.Error(t, err)
		assert.EqualError(t, err, "failed to parse ICMPv4 message: message too short")
	})

	t.Run("Unexpected Message Type", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a reply with a different identifier
		request[4] = 0xFF

		err := protocol.ValidateReply(request, 1234, 1)
		require.Error(t, err)
		assert.EqualError(t, err, "unexpected ICMPv4 message type: echo")
	})

	t.Run("IdentifierOrSequenceMismatch", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}

		// Create a valid ICMP echo request message
		identifier := uint16(1234)
		sequence := uint16(1)
		validRequest, err := protocol.MakeRequest(identifier, sequence)
		require.NoError(t, err)

		// Modify the request to simulate an incorrect identifier or sequence in the reply
		replyMsg := icmp.Message{
			Type: ipv4.ICMPTypeEchoReply, // Correct type for the reply
			Code: 0,
			Body: &icmp.Echo{
				ID:   int(identifier + 1), // Incorrect ID to force a mismatch
				Seq:  int(sequence + 1),   // Incorrect sequence to force a mismatch
				Data: validRequest[8:],    // Keep the rest of the data the same
			},
		}
		reply, err := replyMsg.Marshal(nil)
		require.NoError(t, err)

		// Call ValidateReply with the modified reply
		err = protocol.ValidateReply(reply, identifier, sequence)
		require.Error(t, err)
		assert.EqualError(t, err, "identifier or sequence mismatch")
	})
}

func TestICMPv4_ListenPacket(t *testing.T) {
	t.Parallel()

	t.Run("Successful ListenPacket", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		conn, err := protocol.ListenPacket(ctx, "ip4:icmp", "localhost")
		if err != nil {
			// If we don't have raw-socket privilege, skip instead of failing.
			if isPermissionError(t, err) {
				t.Skip("skipping: requires raw ICMP privileges (root or CAP_NET_RAW)")
			}
			// Some other error? That's a real failure.
			require.NoError(t, err)
		}
		defer conn.Close() // nolint:errcheck

		require.NotNil(t, conn)
	})

	t.Run("Invalid Network", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := protocol.ListenPacket(ctx, "invalid-network", "localhost")
		require.Error(t, err)
		assert.EqualError(t, err, "failed to listen for ICMP packets: listen invalid-network: unknown network invalid-network")
	})

	t.Run("Invalid Address", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv4{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := protocol.ListenPacket(ctx, "ip4:icmp", "invalid-address")

		require.Error(t, err)

		// Check your wrapper/prefix is present (stable across platforms)
		assert.Contains(t, err.Error(), "failed to listen for ICMP packets: listen ip4:icmp: lookup invalid-address:")

		// Unwrap the root cause: accept NXDOMAIN (IsNotFound) or SERVFAIL (IsTemporary)
		var dnsErr *net.DNSError
		if assert.ErrorAs(t, err, &dnsErr) {
			assert.True(t, dnsErr.IsNotFound || dnsErr.IsTemporary,
				"expected NXDOMAIN or temporary DNS failure, got: %+v", dnsErr)
		}
	})
}

func TestICMPv6MakeRequest(t *testing.T) {
	t.Parallel()

	t.Run("MakeRequest", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		msg, err := protocol.MakeRequest(1234, 1)

		require.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Len(t, msg, 23)
	})
}

func TestICMPv6_Network(t *testing.T) {
	t.Parallel()

	t.Run("Network", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		assert.Equal(t, protocol.Network(), "ip6:ipv6-icmp")
	})
}

func TestICMPv6_SetDeadline(t *testing.T) {
	t.Parallel()

	t.Run("SetDeadline Success", func(t *testing.T) {
		t.Parallel()

		mockConn := testutils.MockPacketConn{}
		protocol := &ICMPv6{conn: &mockConn}
		err := protocol.SetDeadline(time.Now().Add(1 * time.Second))

		require.NoError(t, err)
	})

	t.Run("SetDeadline Error", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{conn: nil}
		err := protocol.SetDeadline(time.Now().Add(1 * time.Second))

		require.Error(t, err)
	})
}

func TestICMPv6_ValidateReply(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected Message Type", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a reply with a different identifier
		request[4] = 0xFF

		err := protocol.ValidateReply(request, 1234, 1)

		require.Error(t, err)
		assert.EqualError(t, err, "unexpected ICMPv6 message type: echo request")
	})

	t.Run("ValidateReply Success", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a successful reply by modifying the request type to EchoReply
		reply := request
		reply[0] = byte(ipv6.ICMPTypeEchoReply)

		err := protocol.ValidateReply(reply, 1234, 1)
		require.NoError(t, err)
	})
	t.Run("ValidateReply Identifier Mismatch", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a reply with a different identifier
		reply := request
		reply[4] = 0xFF // Modify the identifier

		err := protocol.ValidateReply(reply, 1234, 1)

		require.Error(t, err)
		assert.EqualError(t, err, "unexpected ICMPv6 message type: echo request")
	})

	t.Run("Error Parsing Message", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		// Pass an invalid byte slice that cannot be parsed as a valid ICMP message
		reply := []byte{0xff, 0xff, 0xff}

		err := protocol.ValidateReply(reply, 1234, 1)
		require.Error(t, err)
		assert.EqualError(t, err, "failed to parse ICMPv6 message: message too short")
	})

	t.Run("Unexpected Message Type", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}
		request, _ := protocol.MakeRequest(1234, 1)

		// Simulate a reply with a different identifier
		request[4] = 0xFF

		err := protocol.ValidateReply(request, 1234, 1)

		require.Error(t, err)
		assert.EqualError(t, err, "unexpected ICMPv6 message type: echo request")
	})

	t.Run("IdentifierOrSequenceMismatch", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}

		// Create a valid ICMP echo request message
		identifier := uint16(1234)
		sequence := uint16(1)
		validRequest, err := protocol.MakeRequest(identifier, sequence)

		require.NoError(t, err)

		// Modify the request to simulate an incorrect identifier or sequence in the reply
		replyMsg := icmp.Message{
			Type: ipv6.ICMPTypeEchoReply, // Correct type for the reply
			Code: 0,
			Body: &icmp.Echo{
				ID:   int(identifier + 1), // Incorrect ID to force a mismatch
				Seq:  int(sequence + 1),   // Incorrect sequence to force a mismatch
				Data: validRequest[8:],    // Keep the rest of the data the same
			},
		}
		reply, err := replyMsg.Marshal(nil)
		require.NoError(t, err)

		// Call ValidateReply with the modified reply
		err = protocol.ValidateReply(reply, identifier, sequence)
		require.Error(t, err)
		assert.EqualError(t, err, "identifier or sequence mismatch")
	})
}

func TestICMPv6_ListenPacket(t *testing.T) {
	t.Parallel()

	t.Run("Successful ListenPacket", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		conn, err := protocol.ListenPacket(ctx, "ip6:ipv6-icmp", "localhost")
		if err != nil {
			// If we don't have raw-socket privilege, skip instead of failing.
			if isPermissionError(t, err) {
				t.Skip("skipping: requires raw ICMP privileges (root or CAP_NET_RAW)")
			}
			// Some other error? That's a real failure.
			require.NoError(t, err)
		}
		defer conn.Close() // nolint:errcheck

		require.NotNil(t, conn)
	})

	t.Run("Invalid Network", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := protocol.ListenPacket(ctx, "invalid-network", "localhost")

		require.Error(t, err)
		assert.EqualError(t, err, "failed to listen for ICMP packets: listen invalid-network: unknown network invalid-network")
	})

	t.Run("Invalid Address", func(t *testing.T) {
		t.Parallel()

		protocol := &ICMPv6{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := protocol.ListenPacket(ctx, "ip6:ipv6-icmp", "invalid-address")

		require.Error(t, err)

		// Check your wrapper/prefix is present (stable across platforms)
		assert.Contains(t, err.Error(), "failed to listen for ICMP packets: listen ip6:ipv6-icmp: lookup invalid-address:")

		// Unwrap the root cause: accept NXDOMAIN (IsNotFound) or SERVFAIL (IsTemporary)
		var dnsErr *net.DNSError
		if assert.ErrorAs(t, err, &dnsErr) {
			assert.True(t, dnsErr.IsNotFound || dnsErr.IsTemporary,
				"expected NXDOMAIN or temporary DNS failure, got: %+v", dnsErr)
		}
	})
}
