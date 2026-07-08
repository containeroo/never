package checker

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	icmpv4ProtocolNumber int    = 1
	icmpv6ProtocolNumber int    = 58
	icmpv4Network        string = "ip4:icmp"
	icmpv6Network        string = "ip6:ipv6-icmp"
)

// Protocol defines an interface for ICMP-based diagnostics, abstracting ICMPv4 and ICMPv6 behavior.
type Protocol interface {
	// MakeRequest creates an ICMP echo request message with the specified identifier and sequence number.
	// Returns the serialized byte representation of the message or an error if message construction fails.
	MakeRequest(identifier, sequence uint16) ([]byte, error)
	// ValidateReply verifies that an ICMP echo reply message matches the expected identifier and sequence number.
	// Returns an error if the reply is invalid, such as a mismatch in identifier, sequence number, or unexpected message type.
	ValidateReply(reply []byte, identifier, sequence uint16) error
	// Network returns the network type string to be used for listening to ICMP packets, which typically indicates the IP
	// protocol version (e.g., "ip4:icmp" for IPv4 ICMP or "ip6:ipv6-icmp" for IPv6 ICMP).
	Network() string
	// ListenPacket sets up a listener for ICMP packets on the specified network and address, using the provided context.
	// Returns a net.PacketConn for reading and writing packets, or an error if the listener cannot be established.
	ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error)
}

// newProtocol initializes a protocol based on the given address.
func newProtocol(address string) (Protocol, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		// If the address is not an IP, try resolving it as a domain name
		ips, err := net.LookupIP(address)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("invalid or unresolvable address: %s", address)
		}
		ip = ips[0] // Use the first resolved IP address
	}

	if ip.To16() != nil && ip.To4() == nil {
		return &ICMPv6{}, nil
	}

	return &ICMPv4{}, nil
}

// ICMPv4 implements the Protocol interface for IPv4 ICMP.
type ICMPv4 struct {
	conn net.PacketConn
}

// MakeRequest creates an ICMP echo request message.
func (p *ICMPv4) MakeRequest(identifier, sequence uint16) ([]byte, error) {
	body := &icmp.Echo{
		ID:   int(identifier),
		Seq:  int(sequence),
		Data: []byte("HELLO-R-U-THERE"),
	}
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}

	return msg.Marshal(nil)
}

// ValidateReply validates an ICMP echo reply message.
func (p *ICMPv4) ValidateReply(reply []byte, identifier, sequence uint16) error {
	parsedMsg, err := icmp.ParseMessage(icmpv4ProtocolNumber, reply)
	if err != nil {
		return fmt.Errorf("failed to parse ICMPv4 message: %w", err)
	}

	if parsedMsg.Type != ipv4.ICMPTypeEchoReply {
		return fmt.Errorf("unexpected ICMPv4 message type: %v", parsedMsg.Type)
	}

	body, ok := parsedMsg.Body.(*icmp.Echo)
	if !ok || body.ID != int(identifier) || body.Seq != int(sequence) {
		return fmt.Errorf("identifier or sequence mismatch")
	}

	return nil
}

// Network returns the network type for the ICMP protocol.
func (p *ICMPv4) Network() string { return icmpv4Network }

// ListenPacket creates a new ICMPv4 packet connection.
func (p *ICMPv4) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	var lc net.ListenConfig
	conn, err := lc.ListenPacket(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	p.conn = conn
	return conn, nil
}

// ICMPv6 implements the Protocol interface for IPv6 ICMP.
type ICMPv6 struct {
	conn net.PacketConn
}

// MakeRequest creates an ICMP echo request message.
func (p *ICMPv6) MakeRequest(identifier, sequence uint16) ([]byte, error) {
	body := &icmp.Echo{
		ID:   int(identifier),
		Seq:  int(sequence),
		Data: []byte("HELLO-R-U-THERE"),
	}
	msg := icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: body,
	}

	return msg.Marshal(nil)
}

// ValidateReply validates an ICMP echo reply message.
func (p *ICMPv6) ValidateReply(reply []byte, identifier, sequence uint16) error {
	parsedMsg, err := icmp.ParseMessage(icmpv6ProtocolNumber, reply)
	if err != nil {
		return fmt.Errorf("failed to parse ICMPv6 message: %w", err)
	}

	if parsedMsg.Type != ipv6.ICMPTypeEchoReply {
		return fmt.Errorf("unexpected ICMPv6 message type: %v", parsedMsg.Type)
	}

	body, ok := parsedMsg.Body.(*icmp.Echo)
	if !ok || body.ID != int(identifier) || body.Seq != int(sequence) {
		return fmt.Errorf("identifier or sequence mismatch")
	}

	return nil
}

// Network returns the network type for the ICMP protocol.
func (p *ICMPv6) Network() string { return icmpv6Network }

// ListenPacket creates a new ICMPv6 packet connection.
func (p *ICMPv6) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	var lc net.ListenConfig
	conn, err := lc.ListenPacket(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	p.conn = conn
	return p.conn, nil
}
