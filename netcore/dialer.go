// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"context"
	"log/slog"
	"net"
)

// Dialer dials TCP/UDP connections.
//
// Construct using [NewDialer].
//
// A [*Dialer] is safe for concurrent use by multiple goroutines
// as long as you don't modify its fields after construction and the
// underlying fields you may set (e.g., DialContext) are also safe.
type Dialer struct {
	// DialContextFunc is the optional dialer for creating new
	// TCP and UDP connections. If this field is nil, the default
	// dialer from the [net] package will be used.
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

	// Logger is the optional structured logger for emitting
	// structured diagnostic events. If this field is nil, we
	// will not be emitting structured logs.
	Logger *slog.Logger
}

// NewDialer constructs a new [*Dialer] with default settings.
func NewDialer() *Dialer {
	return &Dialer{}
}

// DefaultDialer is the default [*Dialer] used by this package.
var DefaultDialer = NewDialer()

// DialContext establishes a new TCP/UDP connection.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// TODO(bassosimone): decouple DNS lookup and dialing
	child := &net.Dialer{}
	child.SetMultipathTCP(false)
	return child.DialContext(ctx, network, address)
}
