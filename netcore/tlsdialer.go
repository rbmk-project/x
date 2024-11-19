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

// DialTLSContext establishes a new TLS connection.
func (nx *Network) DialTLSContext(ctx context.Context, network, address string) (net.Conn, error) {
	config, err := nx.tlsConfig(network, address)
	if err != nil {
		return nil, err
	}

	child := &tls.Dialer{Config: config}

	return child.DialContext(ctx, network, address)
}

func (nx *Network) tlsConfig(network, address string) (*tls.Config, error) {
	if nx.TLSConfig != nil {
		config := nx.TLSConfig.Clone() // make sure we return a cloned config
		return config, nil
	}
	return newTLSConfig(network, address)
}
