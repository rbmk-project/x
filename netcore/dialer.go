//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
//
// Cleartext conn dialer.
//

package netcore

import (
	"context"
	"errors"
	"net"
)

// DialContext establishes a new TCP/UDP connection.
func (nx *Network) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): decide whether we want an overall timeout here,
	// which I don't think is fine, because it's not granular enough.

	// resolve the endpoints to connect to
	endpoints, err := nx.maybeLookupEndpoint(ctx, address)
	if err != nil {
		return nil, err
	}

	// sequentially attempt with each available endpoint
	return nx.sequentialDial(ctx, network, nx.dialLog, endpoints...)
}

// dialContextFunc is a function used to dial a connection.
type dialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

// sequentialDial attempts to dial the endpoints in sequence until one
// of them succeeds. It returns the first successfully established network
// connection, on success, and the union of all errors, otherwise.
func (nx *Network) sequentialDial(
	ctx context.Context,
	network string,
	fx dialContextFunc,
	endpoints ...string,
) (net.Conn, error) {
	// TODO(bassosimone): decide whether we want to sort IPv4 before IPv6
	// here, and whether we want another method for happy eyeballs.
	var errv []error
	for _, endpoint := range endpoints {
		conn, err := fx(ctx, network, endpoint)
		if conn != nil && err == nil {
			return conn, nil
		}
		errv = append(errv, err)
	}
	return nil, errors.Join(errv...)
}

func (nx *Network) dialLog(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): emit structured logs
	return nx.dialNet(ctx, network, address)
}

func (nx *Network) dialNet(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): do we want to automatically wrap the connection?

	// if there's an user provided dialer func, use it
	if nx.DialContextFunc != nil {
		return nx.DialContextFunc(ctx, network, address)
	}

	// otherwise use the net package
	child := &net.Dialer{}
	child.SetMultipathTCP(false)
	return child.DialContext(ctx, network, address)
}
