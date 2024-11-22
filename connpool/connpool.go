// SPDX-License-Identifier: GPL-3.0-or-later

// Package connpool contains a pool of connections.
package connpool

import (
	"errors"
	"io"
	"slices"
	"sync"
)

// Pool is a pool of connections.
//
// Construct using [New].
type Pool struct {
	// conns contains the connections to close.
	conns []io.Closer

	// mu provides mutual exclusion.
	mu sync.Mutex
}

// New constructs a new [*Pool] instance.
func New() *Pool {
	return &Pool{}
}

// Add adds a given [net.Conn] to the pool.
func (p *Pool) Add(conn io.Closer) {
	p.mu.Lock()
	p.conns = append(p.conns, conn)
	p.mu.Unlock()
}

// Close closes all the connections inside the pool iterating
// in backward order. Therefore, if one registers a TCP connection
// and then the corresponding TLS connection, the TLS connection
// is closed first. The returned error is the join of all the
// errors that occurred when closing connections.
func (p *Pool) Close() error {
	// Lock and copy the connections to close.
	p.mu.Lock()
	conns := p.conns
	p.conns = nil
	p.mu.Unlock()

	// Close all the connections.
	var errv []error
	for _, conn := range slices.Backward(conns) {
		if err := conn.Close(); err != nil {
			errv = append(errv, err)
		}
	}
	return errors.Join(errv...)
}
