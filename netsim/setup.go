// SPDX-License-Identifier: GPL-3.0-or-later

package netsim

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/netip"

	"github.com/rbmk-project/dnscore/dnscoretest"
	"github.com/rbmk-project/x/netsim/simpki"
)

// StackConfig contains configuration for creating a new network stack.
type StackConfig struct {
	// DomainNames contains the domain names associated with this stack.
	DomainNames []string

	// Addresses contains the IP addresses for this stack.
	Addresses []string

	// DNSOverUDPHandler optionally specifies a handler for DNS-over-UDP.
	DNSOverUDPHandler DNSHandler

	// ClientResolvers optionally specifies resolvers for client stacks.
	ClientResolvers []string
}

// validate returns an error if the configuration is not valid.
func (cfg *StackConfig) validate() error {
	if len(cfg.Addresses) < 1 {
		return errors.New("at least one address is required")
	}
	return nil
}

// newBaseStack returns the base stack given a [*StackConfig].
func (s *Scenario) newBaseStack(cfg *StackConfig) (*Stack, error) {
	addrs := make([]netip.Addr, len(cfg.Addresses))
	for idx, addr := range cfg.Addresses {
		pa, err := netip.ParseAddr(addr)
		if err != nil {
			return nil, err
		}
		addrs[idx] = pa
	}
	stack := NewStack(addrs...)
	s.router.AddRoute(stack)
	return stack, nil
}

// setupClientResolvers configures the client resolvers for the stack.
func (cfg *StackConfig) setupClientResolvers(stack *Stack) error {
	var ress []netip.AddrPort
	for _, addr := range cfg.ClientResolvers {
		paddr, err := netip.ParseAddrPort(net.JoinHostPort(addr, "53"))
		if err != nil {
			return err
		}
		ress = append(ress, paddr)
	}
	stack.SetResolvers(ress...)
	return nil
}

// mustSetupPKI sets up the PKI database for the stack, if possible.
//
// This method panics on error.
func (s *Scenario) mustSetupPKI(cfg *StackConfig) (tls.Certificate, bool) {
	if len(cfg.DomainNames) <= 0 {
		return tls.Certificate{}, false
	}
	var ipAddr []net.IP
	for _, addr := range cfg.Addresses {
		ipAddr = append(ipAddr, netip.MustParseAddr(addr).AsSlice())
	}
	cert := s.pki.MustNewCert(&simpki.Config{
		CommonName: cfg.DomainNames[0],
		DNSNames:   cfg.DomainNames,
		IPAddrs:    ipAddr,
	})
	return cert, true
}

// mustSetupDNSOverUDP configures the DNS-over-UDP handler for the stack.
func (s *Scenario) mustSetupDNSOverUDP(stack *Stack, cfg *StackConfig) {
	server := &dnscoretest.Server{
		ListenPacket: func(network, address string) (net.PacketConn, error) {
			return stack.ListenPacket(context.Background(), network, "[::]:53")
		},
	}
	<-server.StartUDP(cfg.DNSOverUDPHandler)
	s.pool.Add(server)
}
