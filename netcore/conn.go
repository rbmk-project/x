//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/measurexlite/conn.go
//
// Conn wrapper.
//

package netcore

import "net"

// connLocalAddr is a safe way to get the local address of a connection.
func connLocalAddr(conn net.Conn) net.Addr {
	if conn != nil && conn.LocalAddr() != nil {
		return conn.LocalAddr()
	}
	return emptyAddr{}
}

// connRemoteAddr is a safe way to get the remote address of a connection.
func connRemoteAddr(conn net.Conn) net.Addr {
	if conn != nil && conn.RemoteAddr() != nil {
		return conn.RemoteAddr()
	}
	return emptyAddr{}
}

// emptyAddr is an empty [net.Addr].
type emptyAddr struct{}

// Network implements [net.Addr].
func (emptyAddr) Network() string { return "" }

// String implements [net.Addr].
func (emptyAddr) String() string { return "" }
