// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package netcore provides a TCP/UDP dialer and a TLS dialer.

This package is designed to facilitate measuring TCP, UDP, and TLS
connection events via the [log/slog] package.

# Features

- TCP/UDP [*Network.DialContext] method compatible with [net/http].

- TLS [*Network.DialTLSContext] method compatible with [net/http].

# Design Documents

This package is experimental and has no design documents for now.
*/
package netcore
