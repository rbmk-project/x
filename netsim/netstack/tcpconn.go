//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TCP Conn/PacketConn.
//

package netstack

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"
)

// TCPConn is a TCP connection.
//
// The zero value is invalid; construct using [NewTCPConn].
type TCPConn struct {
	buf      bytes.Buffer
	initonce sync.Once
	*Port
	rlock sync.Mutex
}

// NewTCPConn creates a new TCP connection.
func NewTCPConn(p *Port) *TCPConn {
	return &TCPConn{
		buf:      bytes.Buffer{},
		initonce: sync.Once{},
		Port:     p,
		rlock:    sync.Mutex{},
	}
}

// Accept responds to the incoming SYN with SYN|ACK.
func (c *TCPConn) Accept() (err error) {
	c.initonce.Do(func() {
		c.SetDeadline(time.Now().Add(time.Second))
		defer c.SetDeadline(time.Time{})
		err = c.Port.WritePacket(nil, TCPFlagSYN|TCPFlagACK, netip.AddrPort{})
	})
	return
}

// Connect perform the three-way handshake.
func (c *TCPConn) Connect(ctx context.Context) (err error) {
	c.initonce.Do(func() {
		if d, ok := ctx.Deadline(); ok {
			c.SetDeadline(d)
			defer c.SetDeadline(time.Time{})
		}
		err = c.Port.WritePacket(nil, TCPFlagSYN, netip.AddrPort{})
		if err != nil {
			return
		}
		var pkt *Packet
		pkt, err = c.Port.ReadPacket()
		if err != nil {
			return
		}
		if pkt.Flags == TCPFlagRST {
			err = ECONNREFUSED
			return
		}
		if pkt.Flags != TCPFlagSYN|TCPFlagACK {
			err = ECONNABORTED
			return
		}
	})
	return
}

// Ensure [*TCPConn] implements [net.PacketConn].
var _ net.PacketConn = &TCPConn{}

// Ensure [*TCPConn] implements [net.Conn].
var _ net.Conn = &TCPConn{}

// Read implements [net.Conn].
func (c *TCPConn) Read(buf []byte) (int, error) {
	for {
		// if there's buffered data, just read from the buffer
		// holding the lock just in case (even though one is not
		// supposed to invoke [Read] concurrently)
		c.rlock.Lock()
		count, err := c.buf.Read(buf)
		c.rlock.Unlock()
		if count > 0 {
			return count, nil
		}

		// otherwise, attempt to read the next packet
		pkt, err := c.Port.ReadPacket()
		if err != nil {
			return 0, err
		}

		// handle TCP flags
		if pkt.Flags&TCPFlagFIN != 0 {
			return 0, io.EOF
		}
		if pkt.Flags&TCPFlagRST != 0 {
			return 0, ECONNRESET
		}

		// fill the buffer
		//
		// holding the lock just in case (even though one is not
		// supposed to invoke [Read] concurrently)
		c.rlock.Lock()
		c.buf.Write(pkt.Payload)
		c.rlock.Unlock()
	}
}

// Close implements [net.Conn].
func (c *TCPConn) Close() error {
	c.Port.WritePacket(nil, TCPFlagFIN, netip.AddrPort{})
	return c.Port.Close()
}
