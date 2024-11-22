//go:build unix

//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// UNIX errno definitions.
//

package netsim

import "golang.org/x/sys/unix"

const (
	// EADDRINUSE is the address in use error.
	EADDRINUSE = unix.EADDRINUSE

	// ECONNABORTED is the connection aborted error.
	ECONNABORTED = unix.ECONNABORTED

	// ECONNRESET is the connection reset by peer error.
	ECONNRESET = unix.ECONNRESET

	// EHOSTUNREACH is the host unreachable error.
	EHOSTUNREACH = unix.EHOSTUNREACH

	// EINVAL is the invalid argument error.
	EINVAL = unix.EINVAL

	// ENETDOWN is the network is down error.
	ENETDOWN = unix.ENETDOWN

	// ENOBUFS is the no buffer space available error.
	ENOBUFS = unix.ENOBUFS

	// ENOTCONN is the not connected error.
	ENOTCONN = unix.ENOTCONN

	// EPROTONOSUPPORT is the protocol not supported error.
	EPROTONOSUPPORT = unix.EPROTONOSUPPORT
)