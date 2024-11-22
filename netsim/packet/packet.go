// SPDX-License-Identifier: GPL-3.0-or-later

// Package packet contains [*Packet] and the related definitions.
package packet

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// IPProtocol is the protocol number of an IP packet.
type IPProtocol uint8

// String returns the string representation of the IP protocol.
func (p IPProtocol) String() string {
	switch p {
	case IPProtocolTCP:
		return "tcp"

	case IPProtocolUDP:
		return "udp"

	default:
		return "unknown"
	}
}

const (
	// IPProtocolTCP is the TCP protocol number.
	IPProtocolTCP = 6

	// IPProtocolUDP is the UDP protocol number.
	IPProtocolUDP = 17
)

// TCPFlags is a set of TCP flags.
type TCPFlags uint8

// String returns the string representation of the TCP flags.
func (flags TCPFlags) String() string {
	var builder strings.Builder

	if flags&TCPFlagFIN != 0 {
		builder.WriteString("F")
	} else {
		builder.WriteString(".")
	}

	if flags&TCPFlagSYN != 0 {
		builder.WriteString("S")
	} else {
		builder.WriteString(".")
	}

	if flags&TCPFlagRST != 0 {
		builder.WriteString("R")
	} else {
		builder.WriteString(".")
	}

	if flags&TCPFlagPSH != 0 {
		builder.WriteString("P")
	} else {
		builder.WriteString(".")
	}

	if flags&TCPFlagACK != 0 {
		builder.WriteString("A")
	} else {
		builder.WriteString(".")
	}

	return builder.String()
}

const (
	// TCPFlagFIN is the FIN flag.
	TCPFlagFIN = 1

	// TCPFlagSYN is the SYN flag.
	TCPFlagSYN = 2

	// TCPFlagRST is the RST flag.
	TCPFlagRST = 4

	// TCPFlagPSH is the PSH flag.
	TCPFlagPSH = 8

	// TCPFlagACK is the ACK flag.
	TCPFlagACK = 16
)

// Packet is a network packet.
type Packet struct {
	// SrcAddr is the source address.
	SrcAddr netip.Addr

	// DstAddr is the destination address.
	DstAddr netip.Addr

	// IPProtocol is the protocol number.
	IPProtocol IPProtocol

	// SrcPort is the source port.
	SrcPort uint16

	// DstPort is the destination port.
	DstPort uint16

	// TCPFlags is the TCP flags.
	Flags TCPFlags

	// Payload is the packet payload.
	Payload []byte
}

// String returns the string representation of the packet.
func (p *Packet) String() string {
	switch p.IPProtocol {
	case IPProtocolTCP:
		return p.stringTCP()
	default:
		return p.stringOtherwise()
	}
}

// stringOtherwise returns the string representation of the packet for non-TCP protocols.
func (p *Packet) stringOtherwise() string {
	return fmt.Sprintf(
		"%s -> %s %s length=%d",
		net.JoinHostPort(p.SrcAddr.String(), fmt.Sprintf("%d", p.SrcPort)),
		net.JoinHostPort(p.DstAddr.String(), fmt.Sprintf("%d", p.DstPort)),
		p.IPProtocol.String(),
		len(p.Payload),
	)
}

// stringTCP returns the string representation of the packet for TCP protocol.
func (p *Packet) stringTCP() string {
	return fmt.Sprintf(
		"%s -> %s %s flags=%s length=%d",
		net.JoinHostPort(p.SrcAddr.String(), fmt.Sprintf("%d", p.SrcPort)),
		net.JoinHostPort(p.DstAddr.String(), fmt.Sprintf("%d", p.DstPort)),
		p.IPProtocol.String(),
		p.Flags.String(),
		len(p.Payload),
	)
}

// NetworkDevice is a network device to read/write [*Packet].
type NetworkDevice interface {
	// EOF returns a channel that is closed when the device is closed.
	EOF() <-chan struct{}

	// Input returns a channel to send [*Packet] to the device.
	Input() chan<- *Packet

	// Output returns a channel to receive [*Packet] from the device.
	Output() <-chan *Packet
}