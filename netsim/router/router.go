// SPDX-License-Identifier: GPL-3.0-or-later

// Package router provides network routing capabilities for testing
package router

import (
	"errors"
	"net/netip"

	"github.com/rbmk-project/x/netsim/packet"
)

// Router provides routing capabilities.
type Router struct {
	// devs tracks attached [packet.NetworkDevice].
	devs []packet.NetworkDevice

	// srt is the static routing table.
	srt map[netip.Addr]packet.NetworkDevice
}

// New creates a new [*Router].
func New() *Router {
	return &Router{
		devs: make([]packet.NetworkDevice, 0),
		srt:  make(map[netip.Addr]packet.NetworkDevice),
	}
}

// Attach attaches a [packet.NetworkDevice] to the [*Router].
func (r *Router) Attach(dev packet.NetworkDevice) {
	r.devs = append(r.devs, dev)
	go r.readLoop(dev)
}

// AddRoute adds routes for all addresses of the given [packet.NetworkDevice].
func (r *Router) AddRoute(dev packet.NetworkDevice) {
	for _, addr := range dev.Addresses() {
		r.srt[addr] = dev
	}
}

// readLoop reads packets from a [packet.NetworkDevice] until EOF.
func (r *Router) readLoop(dev packet.NetworkDevice) {
	for {
		select {
		case <-dev.EOF():
			return
		case pkt := <-dev.Output():
			r.route(pkt)
		}
	}
}

var (
	// errTTLExceeded is returned when a packet's TTL is exceeded.
	errTTLExceeded = errors.New("TTL exceeded in transit")

	// errNoRouteToHost is returned when there is no route to the host.
	errNoRouteToHost = errors.New("no route to host")

	// errBufferFull is returned when the buffer is full.
	errBufferFull = errors.New("buffer full")
)

// route routes a given packet to its destination.
func (r *Router) route(pkt *packet.Packet) error {
	// Decrement TTL.
	if pkt.TTL <= 0 {
		return errTTLExceeded
	}
	pkt.TTL--

	// Find next hop.
	nextHop := r.srt[pkt.DstAddr]
	if nextHop == nil {
		return errNoRouteToHost
	}

	// Forward packet (non-blocking)
	select {
	case nextHop.Input() <- pkt:
		return nil
	default:
		return errBufferFull
	}
}
