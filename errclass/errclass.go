// SPDX-License-Identifier: GPL-3.0-or-later

/*
Package errclass implements error classification.

The general idea is to classify golang errors to an enum of strings
with names resembling standard Unix error names.

Deprecated: use `github.com/rbmk-project/common/errclass` instead.

# Design Principles

1. Preserve original error in `err` in the structured logs.

2. Add the classified error as the `errClass` field.

3. Use [errors.Is] and [errors.As] for classification.

4. Use string-based classification for readability.

5. Follow Unix-like naming where appropriate.

6. Prefix subsystem-specific errors (`EDNS_`, `ETLS_`).

7. Keep full names for clarity over brevity.

8. Map the nil error to an empty string.

# System and Network Errors

- [ETIMEDOUT] for [context.DeadlineExceeded], [os.ErrDeadlineExceeded]

- [EINTR] for [context.Canceled], [net.ErrClosed]

- [EEOF] for (unexpected) [io.EOF] and [io.ErrUnexpectedEOF] errors

- [ECONNRESET], [ECONNREFUSED], ... for respective syscall errors

The actual system error constants are defined in platform-specific files:

- unix.go for Unix-like systems using x/sys/unix

- windows.go for Windows systems using x/sys/windows

This ensures proper mapping between the standardized error classes and
platform-specific error constants.

# DNS Errors

- [EDNS_NONAME] for errors with the "no such host" suffix

- [EDNS_NODATA] for errors with the "no answer" suffix

# TLS

- [ETLS_HOSTNAME_MISMATCH] for hostname verification failure

- [ETLS_CA_UNKNOWN] for unknown certificate authority

- [ETLS_CERT_INVALID] for invalid certificate

# Fallback

- [EGENERIC] for unclassified errors
*/
package errclass

import (
	"github.com/rbmk-project/common/errclass"
)

const (
	//
	// Errors that we can map using [errors.Is]:
	//

	// EADDRNOTAVAIL is the address not available error.
	EADDRNOTAVAIL = errclass.EADDRNOTAVAIL

	// EADDRINUSE is the address in use error.
	EADDRINUSE = errclass.EADDRINUSE

	// ECONNABORTED is the connection aborted error.
	ECONNABORTED = errclass.ECONNABORTED

	// ECONNREFUSED is the connection refused error.
	ECONNREFUSED = errclass.ECONNREFUSED

	// ECONNRESET is the connection reset by peer error.
	ECONNRESET = errclass.ECONNRESET

	// EHOSTUNREACH is the host unreachable error.
	EHOSTUNREACH = errclass.EHOSTUNREACH

	// EEOF indicates an unexpected EOF.
	EEOF = errclass.EEOF

	// EINVAL is the invalid argument error.
	EINVAL = errclass.EINVAL

	// EINTR is the interrupted system call error.
	EINTR = errclass.EINTR

	// ENETDOWN is the network is down error.
	ENETDOWN = errclass.ENETDOWN

	// ENETUNREACH is the network unreachable error.
	ENETUNREACH = errclass.ENETUNREACH

	// ENOBUFS is the no buffer space available error.
	ENOBUFS = errclass.ENOBUFS

	// ENOTCONN is the not connected error.
	ENOTCONN = errclass.ENOTCONN

	// EPROTONOSUPPORT is the protocol not supported error.
	EPROTONOSUPPORT = errclass.EPROTONOSUPPORT

	// ETIMEDOUT is the operation timed out error.
	ETIMEDOUT = errclass.ETIMEDOUT

	//
	// Errors that we can map using the error message suffix:
	//

	// EDNS_NONAME is the DNS error for "no such host".
	EDNS_NONAME = errclass.EDNS_NONAME

	// EDNS_NODATA 	is the DNS error for "no answer".
	EDNS_NODATA = errclass.EDNS_NODATA

	//
	// Errors that we can map using [errors.As]:
	//

	// ETLS_HOSTNAME_MISMATCH is the TLS error for hostname verification failure.
	ETLS_HOSTNAME_MISMATCH = errclass.ETLS_HOSTNAME_MISMATCH

	// ETLS_CA_UNKNOWN is the TLS error for unknown certificate authority.
	ETLS_CA_UNKNOWN = errclass.ETLS_CA_UNKNOWN

	// ETLS_CERT_INVALID is the TLS error for invalid certificate.
	ETLS_CERT_INVALID = errclass.ETLS_CERT_INVALID

	//
	// Fallback errors:
	//

	// EGENERIC is the generic, unclassified error.
	EGENERIC = errclass.EGENERIC
)

// New is an alias for [errclass.New].
var New = errclass.New
