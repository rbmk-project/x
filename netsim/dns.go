// SPDX-License-Identifier: GPL-3.0-or-later

package netsim

import (
	"net"

	"github.com/miekg/dns"
	"github.com/rbmk-project/common/runtimex"
	"github.com/rbmk-project/dnscore/dnscoretest"
)

// DNSHandler is an alias for dnscoretest.DNSHandler.
type DNSHandler = dnscoretest.Handler

// dnsDatabase is the global DNS database.
type dnsDatabase struct {
	names map[string][]dns.RR
}

// newDNSDatabase creates a new DNS database.
func newDNSDatabase() *dnsDatabase {
	return &dnsDatabase{
		names: make(map[string][]dns.RR),
	}
}

// AddCNAME adds a CNAME alias.
//
// This method IS NOT goroutine safe.
func (dd *dnsDatabase) AddCNAME(name, alias string) {
	header := dns.RR_Header{
		Name:     dns.CanonicalName(name),
		Rrtype:   dns.TypeCNAME,
		Class:    dns.ClassINET,
		Ttl:      3600,
		Rdlength: 0,
	}

	rr := &dns.CNAME{
		Hdr:    header,
		Target: dns.CanonicalName(alias),
	}

	dd.names[name] = append(dd.names[name], rr)
}

// AddFromConfig adds DNS records from a [*StackConfig].
//
// This method IS NOT goroutine safe.
func (dd *dnsDatabase) AddFromConfig(cfg *StackConfig) {
	for _, name := range cfg.DomainNames {
		name = dns.CanonicalName(name)
		for _, addr := range cfg.Addresses {
			// Make sure the string is a valid IP address
			ipAddr := net.ParseIP(addr)
			runtimex.Assert(ipAddr != nil, "invalid IP address")

			// Create the common DNS header
			header := dns.RR_Header{
				Name:     dns.CanonicalName(name),
				Rrtype:   0,
				Class:    dns.ClassINET,
				Ttl:      3600,
				Rdlength: 0,
			}

			// Create the DNS record to add
			var rr dns.RR
			switch ipAddr.To4() {
			case nil:
				header.Rrtype = dns.TypeAAAA
				rr = &dns.AAAA{Hdr: header, AAAA: ipAddr}
			default:
				header.Rrtype = dns.TypeA
				rr = &dns.A{Hdr: header, A: ipAddr}
			}

			dd.names[name] = append(dd.names[name], rr)
		}
	}
}

// Ensure [*dnsDatabase] implements [dnsHandler].
var _ DNSHandler = (*dnsDatabase)(nil)

// Handler implements [dnsHandler] using [*dnsDatabase].
//
// This method is goroutine safe as long as one does not
// modify the database while handling queries.
func (dd *dnsDatabase) Handle(rw dnscoretest.ResponseWriter, rawQuery []byte) {
	// Parse the incoming query and make sure it's a
	// query containing just one question.
	var (
		response = &dns.Msg{}
		query    = &dns.Msg{}
	)
	if err := query.Unpack(rawQuery); err != nil {
		return
	}
	if query.Response || query.Opcode != dns.OpcodeQuery || len(query.Question) != 1 {
		return
	}
	response.SetReply(query)

	// Get the RRs if possible
	var (
		q0   = query.Question[0]
		name = dns.CanonicalName(q0.Name)
	)
	switch {
	case q0.Qclass != dns.ClassINET:
		response.Rcode = dns.RcodeRefused
	case q0.Qtype == dns.TypeA ||
		q0.Qtype == dns.TypeAAAA ||
		q0.Qtype == dns.TypeCNAME:
		var found bool
		response.Answer, found = dd.lookup(q0.Qtype, name)
		if !found {
			response.Rcode = dns.RcodeNameError
		}
	default:
		response.Rcode = dns.RcodeNameError
	}

	// Write the response
	rawResp, err := response.Pack()
	if err != nil {
		return
	}
	rw.Write(rawResp)
}

// lookup returns the DNS records for a domain name.
//
// This method is goroutine safe as long as one does not
// modify the database while handling queries.
func (dd *dnsDatabase) lookup(qtype uint16, name string) ([]dns.RR, bool) {
	const maxloops = 10
	var rrs []dns.RR
	for idx := 0; idx < maxloops; idx++ {

		// Search whether the current name is in the database.
		var interim []dns.RR
		interim, found := dd.names[name]
		if !found {
			return nil, false
		}

		// We have definitely found something related.
		rrs = append(rrs, interim...)

		// Check whether we have found the desired record.
		for _, rr := range interim {
			if qtype == rr.Header().Rrtype {
				return rrs, true
			}
		}

		// Otherwise, follow CNAME redirects.
		var cname string
		for _, rr := range interim {
			if rr, ok := rr.(*dns.CNAME); ok {
				cname = rr.Target
				break
			}
		}
		if cname == "" {
			return nil, false
		}

		// Continue searching from the CNAME target.
		name = cname
	}

	return nil, false
}
