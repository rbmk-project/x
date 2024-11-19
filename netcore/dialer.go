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
	// TODO(bassosimone): decide whether we want an overall timeout here

	// resolve the domain name to IP addresses
	domain, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addrs, err := nx.maybeLookupHost(ctx, domain)
	if err != nil {
		return nil, err
	}

	// TODO(bassosimone): decide whether we want to use happy eyeballs here

	// attempt using each IP address
	var errv []error
	for _, addr := range addrs {
		address = net.JoinHostPort(addr, port)
		conn, err := nx.dialLog(ctx, network, address)
		if conn != nil && err == nil {
			return conn, nil
		}
		errv = append(errv, err)
	}
	return nil, errors.Join(errv...)
}

// maybeLookupHost resolves a domain name to IP addresses unless the domain
// is already an IP address, in which case we short circuit the lookup.
func (nx *Network) maybeLookupHost(ctx context.Context, domain string) ([]string, error) {
	// handle the case where domain is already an IP address
	if net.ParseIP(domain) != nil {
		return []string{domain}, nil
	}

	// TODO(bassosimone): we should probably ensure we nonetheless
	// include the lookup event inside the logs.
	return nx.doLookupHost(ctx, domain)
}

func (nx *Network) doLookupHost(ctx context.Context, domain string) ([]string, error) {
	// if there is a custom LookupHostFunc, use it
	if nx.LookupHostFunc != nil {
		return nx.LookupHostFunc(ctx, domain)
	}

	// otherwise fallback to the system resolver
	reso := &net.Resolver{}
	return reso.LookupHost(ctx, domain)
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
