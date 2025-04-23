package checker

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"
)

const (
	defaultICMPReadTimeout  time.Duration = 1 * time.Second
	defaultICMPWriteTimeout time.Duration = 1 * time.Second
)

// ICMPChecker implements the Checker interface for ICMP checks.
type ICMPChecker struct {
	name         string
	address      string
	readTimeout  time.Duration
	writeTimeout time.Duration
	protocol     Protocol
}

func (c *ICMPChecker) Address() string { return c.address }
func (c *ICMPChecker) Name() string    { return c.name }
func (c *ICMPChecker) Type() string    { return ICMP.String() }

func (c *ICMPChecker) Check(ctx context.Context) error {
	dst, err := net.ResolveIPAddr(c.protocol.Network(), c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve IP address '%s': %w", c.address, err)
	}

	conn, err := c.protocol.ListenPacket(ctx, c.protocol.Network(), "")
	if err != nil {
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	defer conn.Close() // nolint:errcheck

	id := uint16(os.Getpid() & 0xffff)                       // Create a unique identifier
	seq := uint16(atomic.AddUint32(new(uint32), 1) & 0xffff) // Create a unique sequence number

	msg, err := c.protocol.MakeRequest(id, seq)
	if err != nil {
		return fmt.Errorf("failed to create ICMP request: %w", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err := conn.WriteTo(msg, dst); err != nil {
		return fmt.Errorf("failed to send ICMP request: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return fmt.Errorf("failed to read ICMP reply: %w", err)
	}

	if err := c.protocol.ValidateReply(reply[:n], id, seq); err != nil {
		return fmt.Errorf("failed to validate ICMP reply: %w", err)
	}

	return nil
}

// newICMPChecker initializes a new ICMPChecker with functional options.
func newICMPChecker(name, address string, opts ...Option) (*ICMPChecker, error) {
	checker := &ICMPChecker{
		name:         name,
		address:      address,
		readTimeout:  defaultICMPReadTimeout,
		writeTimeout: defaultICMPWriteTimeout,
	}

	for _, opt := range opts {
		opt.apply(checker)
	}

	protocol, err := newProtocol(checker.address)
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP protocol: %w", err)
	}
	checker.protocol = protocol

	return checker, nil
}

// WithICMPReadTimeout sets the read timeout for the ICMPChecker.
func WithICMPReadTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if icmpChecker, ok := c.(*ICMPChecker); ok {
			icmpChecker.readTimeout = timeout
		}
	})
}

// WithICMPWriteTimeout sets the write timeout for the ICMPChecker.
func WithICMPWriteTimeout(timeout time.Duration) Option {
	return OptionFunc(func(c Checker) {
		if icmpChecker, ok := c.(*ICMPChecker); ok {
			icmpChecker.writeTimeout = timeout
		}
	})
}
