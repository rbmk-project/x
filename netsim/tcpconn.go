//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TCP Conn/PacketConn.
//

package netsim

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
	// prevent concurrent goroutines from messing with the read buffer
	c.rlock.Lock()
	defer c.rlock.Unlock()

	for {
		// if there's buffered data, just read from the buffer
		if count, err := c.buf.Read(buf); count > 0 && err == nil {
			return count, nil
		}

		// otherwise, attempt to read the next packet
		pkt, err := c.Port.ReadPacket()
		if err != nil {
			return 0, nil
		}

		// handle TCP flags
		if pkt.Flags&TCPFlagFIN != 0 {
			return 0, io.EOF
		}
		if pkt.Flags&TCPFlagRST != 0 {
			return 0, ECONNRESET
		}

		// fill the buffer
		c.buf.Write(pkt.Payload)
	}
}

// Close implements [net.Conn].
func (c *TCPConn) Close() error {
	c.Port.WritePacket(nil, TCPFlagFIN, netip.AddrPort{})
	return c.Port.Close()
}
