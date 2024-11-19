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
	"log/slog"
	"net"
)

// Dialer dials TCP/UDP connections.
//
// Construct using [NewDialer].
//
// A [*Dialer] is safe for concurrent use by multiple goroutines
// as long as you don't modify its fields after construction and the
// underlying fields you may set (e.g., DialContextFunc) are also safe.
type Dialer struct {
	// DialContextFunc is the optional dialer for creating new
	// TCP and UDP connections. If this field is nil, the default
	// dialer from the [net] package will be used.
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

	// Logger is the optional structured logger for emitting
	// structured diagnostic events. If this field is nil, we
	// will not be emitting structured logs.
	Logger *slog.Logger

	// LookupHostFunc is the optional function to resolve a domain
	// name to IP addresses. If this field is nil, we use the
	// default [*net.Resolver] from the [net] package.
	LookupHostFunc func(ctx context.Context, domain string) ([]string, error)
}

// NewDialer constructs a new [*Dialer] with default settings.
func NewDialer() *Dialer {
	return &Dialer{}
}

// DefaultDialer is the default [*Dialer] used by this package.
var DefaultDialer = NewDialer()

// DialContext establishes a new TCP/UDP connection.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): decide whether we want an overall timeout here

	// resolve the domain name to IP addresses
	domain, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addrs, err := d.maybeLookupHost(ctx, domain)
	if err != nil {
		return nil, err
	}

	// TODO(bassosimone): decide whether we want to use happy eyeballs here

	// attempt using each IP address
	var errv []error
	for _, addr := range addrs {
		address = net.JoinHostPort(addr, port)
		conn, err := d.dialLog(ctx, network, address)
		if conn != nil && err == nil {
			return conn, nil
		}
		errv = append(errv, err)
	}
	return nil, errors.Join(errv...)
}

// maybeLookupHost resolves a domain name to IP addresses unless the domain
// is already an IP address, in which case we short circuit the lookup.
func (d *Dialer) maybeLookupHost(ctx context.Context, domain string) ([]string, error) {
	// handle the case where domain is already an IP address
	if net.ParseIP(domain) != nil {
		return []string{domain}, nil
	}

	// TODO(bassosimone): we should probably ensure we nonetheless
	// include the lookup event inside the logs.
	return d.doLookupHost(ctx, domain)
}

func (d *Dialer) doLookupHost(ctx context.Context, domain string) ([]string, error) {
	// if there is a custom LookupHostFunc, use it
	if d.LookupHostFunc != nil {
		return d.LookupHostFunc(ctx, domain)
	}

	// otherwise fallback to the system resolver
	reso := &net.Resolver{}
	return reso.LookupHost(ctx, domain)
}

func (d *Dialer) dialLog(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): emit structured logs
	return d.dialNet(ctx, network, address)
}

func (d *Dialer) dialNet(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): do we want to automatically wrap the connection?

	// if there's an user provided dialer func, use it
	if d.DialContextFunc != nil {
		return d.DialContextFunc(ctx, network, address)
	}

	// otherwise use the net package
	child := &net.Dialer{}
	child.SetMultipathTCP(false)
	return child.DialContext(ctx, network, address)
}
