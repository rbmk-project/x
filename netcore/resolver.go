//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Adapted from: https://github.com/ooni/probe-cli/blob/v3.20.1/internal/netxlite/dialer.go
//
// Internal code for DNS lookups.
//

package netcore

import (
	"context"
	"net"
)

// maybeLookupEndpoint resolves the domain name inside an endpoint into
// a list of TCP/UDP endpoints. If the domain name is already an IP
// address, we short circuit the lookup.
func (nx *Network) maybeLookupEndpoint(ctx context.Context, endpoint string) ([]string, error) {
	domain, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, err
	}

	addrs, err := nx.maybeLookupHost(ctx, domain)
	if err != nil {
		return nil, err
	}

	var endpoints []string
	for _, addr := range addrs {
		endpoints = append(endpoints, net.JoinHostPort(addr, port))
	}
	return endpoints, nil
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

// doLookupHost performs the DNS lookup.
func (nx *Network) doLookupHost(ctx context.Context, domain string) ([]string, error) {
	// if there is a custom LookupHostFunc, use it
	if nx.LookupHostFunc != nil {
		return nx.LookupHostFunc(ctx, domain)
	}

	// otherwise fallback to the system resolver
	reso := &net.Resolver{}
	return reso.LookupHost(ctx, domain)
}
