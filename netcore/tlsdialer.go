// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
)

// TLSDialer dials TLS connections.
//
// Construct using [NewTLSDialer].
//
// A [*TLSDialer] is safe for concurrent use by multiple goroutines as long as
// you don't modify its fields after construction and the underlying fields you
// may set (e.g., DialContextFunc) are also safe.
type TLSDialer struct {
	// Config is the TLS client config to use. If this field is nil, we
	// will try to create a suitable config based on the network and address
	// that are passed to the DialContext method.
	Config *tls.Config

	// DialContextFunc is the optional dialer for creating new TCP and UDP
	// connections. If this field is nil, we use [DefaultDialer].
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

	// Logger is the optional structured logger for emitting
	// structured diagnostic events. If this field is nil, we
	// will not be emitting structured logs.
	Logger *slog.Logger
}

// NewTLSDialer constructs a new [*TLSDialer] with default settings.
func NewTLSDialer() *TLSDialer {
	return &TLSDialer{}
}

// DialContext establishes a new TLS connection.
func (td *TLSDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	config, err := td.config(network, address)
	if err != nil {
		return nil, err
	}

	// TODO(bassosimone): we should use DialContextFunc instead,
	// which means we need to manually dial here
	child := &tls.Dialer{Config: config}

	return child.DialContext(ctx, network, address)
}

func (td *TLSDialer) config(network, address string) (*tls.Config, error) {
	if td.Config != nil {
		config := td.Config.Clone() // make sure we return a cloned config
		return config, nil
	}
	return newTLSConfig(network, address)
}
