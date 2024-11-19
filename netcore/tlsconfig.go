//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// TLS config code.
//

package netcore

import (
	"crypto/tls"
	"net"
)

// newTLSConfig is a best-effort attempt at creating a suitable TLS config
// for TCP and UDP transports using the network and address.
func newTLSConfig(network, address string) (*tls.Config, error) {
	sni, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		RootCAs:    nil, // TODO(bassosimone): bundle Mozilla CA store
		NextProtos: []string{},
		ServerName: sni,
	}
	switch {
	case port == "443" && network == "tcp":
		config.NextProtos = []string{"h2", "http/1.1"}
	case port == "443" && network == "udp":
		config.NextProtos = []string{"h3"}
	case port == "853" && network == "tcp":
		config.NextProtos = []string{"doh"}
	}

	return config, nil
}
