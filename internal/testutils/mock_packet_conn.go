package testutils

import (
	"context"
	"net"
	"time"
)

// MockProtocol is a mock implementation of the Protocol interface for testing.
type MockProtocol struct {
	MakeRequestFunc   func(identifier, sequence uint16) ([]byte, error)
	ValidateReplyFunc func(reply []byte, identifier, sequence uint16) error
	NetworkFunc       func() string
	ListenPacketFunc  func(ctx context.Context, network, address string) (net.PacketConn, error)
}

// MakeRequest is a mock implementation of the Protocol.MakeRequest method.
func (m *MockProtocol) MakeRequest(identifier, sequence uint16) ([]byte, error) {
	if m.MakeRequestFunc != nil {
		return m.MakeRequestFunc(identifier, sequence)
	}
	return nil, nil
}

// ValidateReply is a mock implementation of the Protocol.ValidateReply method.
func (m *MockProtocol) ValidateReply(reply []byte, identifier, sequence uint16) error {
	if m.ValidateReplyFunc != nil {
		return m.ValidateReplyFunc(reply, identifier, sequence)
	}
	return nil
}

// Network is a mock implementation of the Protocol.Network method.
func (m *MockProtocol) Network() string {
	if m.NetworkFunc != nil {
		return m.NetworkFunc()
	}
	return ""
}

// ListenPacket is a mock implementation of the Protocol.ListenPacket method.
func (m *MockProtocol) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	if m.ListenPacketFunc != nil {
		return m.ListenPacketFunc(ctx, network, address)
	}
	return nil, nil
}

// MockPacketConn is a mock implementation of net.PacketConn for testing purposes.
type MockPacketConn struct {
	SetDeadlineFunc      func(t time.Time) error
	SetReadDeadlineFunc  func(t time.Time) error
	SetWriteDeadlineFunc func(t time.Time) error
	WriteToFunc          func(b []byte, addr net.Addr) (int, error)
	ReadFromFunc         func(b []byte) (int, net.Addr, error)
	CloseFunc            func() error
	LocalAddrFunc        func() net.Addr
	RemoteAddrFunc       func() net.Addr
}

// SetDeadline is a mock implementation of the net.PacketConn.
func (m *MockPacketConn) SetDeadline(t time.Time) error {
	if m.SetDeadlineFunc != nil {
		return m.SetDeadlineFunc(t)
	}
	return nil
}

// SetReadDeadline is a mock implementation of the net.Packe
func (m *MockPacketConn) SetReadDeadline(t time.Time) error {
	if m.SetReadDeadlineFunc != nil {
		return m.SetReadDeadlineFunc(t)
	}
	return nil
}

// SetWriteDeadline is a mock implementation of the net.PacketConn.
func (m *MockPacketConn) SetWriteDeadline(t time.Time) error {
	if m.SetWriteDeadlineFunc != nil {
		return m.SetWriteDeadlineFunc(t)
	}
	return nil
}

// WriteTo is a mock implementation of the net.PacketConn.
func (m *MockPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if m.WriteToFunc != nil {
		return m.WriteToFunc(b, addr)
	}
	return len(b), nil
}

// ReadFrom is a mock implementation of the net.Packe
func (m *MockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if m.ReadFromFunc != nil {
		return m.ReadFromFunc(b)
	}
	return 0, nil, nil
}

// Close is a mock implementation of the net.PacketConn.
func (m *MockPacketConn) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// LocalAddr is a mock implementation of the net
func (m *MockPacketConn) LocalAddr() net.Addr {
	if m.LocalAddrFunc != nil {
		return m.LocalAddrFunc()
	}
	return &net.IPAddr{}
}

// RemoteAddr is a mock implementation of the net
func (m *MockPacketConn) RemoteAddr() net.Addr {
	if m.RemoteAddrFunc != nil {
		return m.RemoteAddrFunc()
	}
	return &net.IPAddr{}
}
