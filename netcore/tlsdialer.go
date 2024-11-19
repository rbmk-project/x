//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TLS dialing code
//

package netcore

import (
	"context"
	"crypto/tls"
	"net"
)

// TLSConn is the interface implementing [*tls.Conn] as well as
// the conn exported by alternative TLS libraries.
type TLSConn interface {
	ConnectionState() tls.ConnectionState
	HandshakeContext(ctx context.Context) error
	net.Conn
}

// DialTLSContext establishes a new TLS connection.
func (nx *Network) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	// obtain the TLS config to use
	config, err := nx.tlsConfig(network, address)
	if err != nil {
		return nil, err
	}

	// resolve the endpoints to connect to
	endpoints, err := nx.maybeLookupEndpoint(ctx, address)
	if err != nil {
		return nil, err
	}

	// build a TLS dialer
	td := &tlsDialer{config: config, netx: nx}

	// sequentially attempt with each available endpoint
	return nx.sequentialDial(ctx, network, td.dial, endpoints...)
}

type tlsDialer struct {
	config *tls.Config
	netx   *Network
}

func (td *tlsDialer) dial(ctx context.Context, network, address string) (net.Conn, error) {
	// dial and log the results of dialing
	conn, err := td.netx.dialLog(ctx, network, address)
	if err != nil {
		return nil, err
	}

	// create TLS client connection
	tconn := td.netx.newTLSClientConn(conn, td.config)

	// TODO(bassosimone): emit before handshake event

	// perform the TLS handshake
	err = tconn.HandshakeContext(ctx)

	// TODO(bassosimone): emit after handshake event

	// process the results
	if err != nil {
		conn.Close()
		return nil, err
	}
	return tconn, nil
}

// newTLSClientConn creates a new TLS client connection.
func (nx *Network) newTLSClientConn(conn net.Conn, config *tls.Config) TLSConn {
	if nx.NewTLSClientConn != nil {
		return nx.NewTLSClientConn(conn, config)
	}
	return tls.Client(conn, config)
}
