// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
)

// Network allows dialing and measuring TCP/UDP/TLS connections.
//
// Construct using [NewNetwork].
//
// A [*Network] is safe for concurrent use by multiple goroutines as long as
// you don't modify its fields after construction and the underlying fields you
// may set (e.g., DialContextFunc) are also safe.
type Network struct {
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

	// TLSConfig is the TLS client config to use. If this field is nil, we
	// will try to create a suitable config based on the network and address
	// that are passed to the DialTLSContext method.
	TLSConfig *tls.Config
}

// NewNetwork constructs a new [*Network] with default settings.
func NewNetwork() *Network {
	return &Network{}
}

// DefaultNetwork is the default [*Network] used by this package.
var DefaultNetwork = NewNetwork()
