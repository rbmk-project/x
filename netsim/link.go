//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Packet and related definitions.
//

package netsim

import (
	"log"
	"sync"

	"github.com/rbmk-project/x/netsim/packet"
)

// LinkStack is the [*Stack] as seen by a [*Link].
type LinkStack = packet.NetworkDevice

// Link models a link between two [*Stack] instances.
//
// The zero value is not ready to use; construct using [NewLink].
type Link struct {
	// eof unblocks any blocking channel operation.
	eof chan struct{}

	// eofOnce ensures we close just once.
	eofOnce sync.Once
}

// NewLink creates a new [*Link] using two [*Stack] and
// sets up moving packets between the two stacks. Use Close
// to shut down background goroutines.
func NewLink(left, right LinkStack) *Link {
	lnk := &Link{
		eof:     make(chan struct{}),
		eofOnce: sync.Once{},
	}
	go lnk.move(left, right)
	go lnk.move(right, left)
	return lnk
}

// Close stops background goroutines moving traffic.
func (lnk *Link) Close() error {
	lnk.eofOnce.Do(func() { close(lnk.eof) })
	return nil
}

type readableStack interface {
	EOF() <-chan struct{}
	Output() <-chan *Packet
}

type writableStack interface {
	EOF() <-chan struct{}
	Input() chan<- *Packet
}

// move moves packets from the left stack to the right stack.
func (lnk *Link) move(left readableStack, right writableStack) {
	for {
		// Read from left stack.
		select {
		case <-lnk.eof:
			return
		case <-left.EOF():
			return
		case pkt := <-left.Output():

			// Write to right stack.
			log.Printf("INFLIGHT %s", pkt)
			select {
			case <-lnk.eof:
				return
			case <-right.EOF():
				return
			case right.Input() <- pkt:
				// success
			}

		}
	}
}
