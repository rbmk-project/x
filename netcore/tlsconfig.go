// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"crypto/tls"
	"net"
)

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
