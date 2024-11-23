//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// UDP Conn/PacketConn.
//

package netstack

import "net"

// UDPConn is a UDP connection.
//
// The zero value is invalid; construct using [NewUDPConn].
type UDPConn struct {
	*Port
}

// NewUDPConn creates a new UDP connection.
func NewUDPConn(p *Port) *UDPConn {
	return &UDPConn{Port: p}
}

// Ensure [*UDPConn] implements [net.PacketConn].
var _ net.PacketConn = &UDPConn{}

// Ensure [*UDPConn] implements [net.Conn].
var _ net.Conn = &UDPConn{}

// Read implements [net.Conn].
func (c *UDPConn) Read(buf []byte) (int, error) {
	count, _, err := c.Port.ReadFrom(buf)
	return count, err
}
