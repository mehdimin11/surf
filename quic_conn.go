package surf

import "net"

// quicPacketConn wraps a net.Conn to implement net.PacketConn interface needed by QUIC.
// This wrapper is used for SOCKS5 UDP proxy connections that need to work with QUIC.
type quicPacketConn struct {
	net.Conn
	remoteAddr net.Addr
}

// ReadFrom implements net.PacketConn interface.
// All data is read from the underlying connection and attributed to the remote address.
func (qpc *quicPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	n, err := qpc.Conn.Read(p)
	return n, qpc.remoteAddr, err
}

// WriteTo implements net.PacketConn interface.
// All data is written to the underlying connection, ignoring the provided address
// since SOCKS5 proxy handles routing.
func (qpc *quicPacketConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	return qpc.Conn.Write(p)
}

// SetReadBuffer sets the read buffer size if the underlying connection supports it.
func (qpc *quicPacketConn) SetReadBuffer(bytes int) error {
	if udpConn, ok := qpc.Conn.(*net.UDPConn); ok {
		return udpConn.SetReadBuffer(bytes)
	}
	return nil
}

// SetWriteBuffer sets the write buffer size if the underlying connection supports it.
func (qpc *quicPacketConn) SetWriteBuffer(bytes int) error {
	if udpConn, ok := qpc.Conn.(*net.UDPConn); ok {
		return udpConn.SetWriteBuffer(bytes)
	}
	return nil
}

// newQUICPacketConn creates a new QUIC packet connection wrapper.
func newQUICPacketConn(conn net.Conn, remoteAddr net.Addr) *quicPacketConn {
	return &quicPacketConn{
		Conn:       conn,
		remoteAddr: remoteAddr,
	}
}