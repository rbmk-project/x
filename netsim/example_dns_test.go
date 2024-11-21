// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"

	"github.com/miekg/dns"
	"github.com/rbmk-project/x/connpool"
	"github.com/rbmk-project/x/netsim"
)

// This example shows how to use [netsim] to simulate a DNS
// server that listens for incoming requests over UDP.
func Example_dnsOverUDP() {
	// Create a pool to close resources when done.
	cpool := connpool.New()
	defer cpool.Close()

	// Create the server stack.
	serverAddr := netip.MustParseAddr("8.8.8.8")
	serverStack := netsim.NewStack(serverAddr)
	cpool.Add(serverStack)

	// Create the client stack.
	clientAddr := netip.MustParseAddr("130.192.91.211")
	clientStack := netsim.NewStack(clientAddr)
	cpool.Add(clientStack)

	// Link the client and the server stacks.
	link := netsim.NewLink(clientStack, serverStack)
	cpool.Add(link)

	// Create a context with a watchdog timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create the server UDP listener.
	serverEndpoint := netip.AddrPortFrom(serverAddr, 53)
	serverConn, err := serverStack.ListenPacket(ctx, "udp", serverEndpoint.String())
	if err != nil {
		log.Fatal(err)
	}
	cpool.Add(serverConn)

	// Start the server in the background.
	serverDNS := &dns.Server{
		PacketConn: serverConn,
		Handler: dns.HandlerFunc(func(rw dns.ResponseWriter, query *dns.Msg) {
			resp := &dns.Msg{}
			resp.SetReply(query)
			resp.Answer = append(resp.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:     "dns.google.",
					Rrtype:   dns.TypeA,
					Class:    dns.ClassINET,
					Ttl:      3600,
					Rdlength: 0,
				},
				A: net.IPv4(8, 8, 8, 8),
			})
			if err := rw.WriteMsg(resp); err != nil {
				log.Fatal(err)
			}
		}),
	}
	go serverDNS.ActivateAndServe()
	defer serverDNS.Shutdown()

	// Create the client connection with the DNS server.
	conn, err := clientStack.DialContext(ctx, "udp", serverEndpoint.String())
	if err != nil {
		log.Fatal(err)
	}
	cpool.Add(conn)

	// Create the query to send
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = []dns.Question{{
		Name:   "dns.google.",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}}

	// Perform the DNS round trip
	clientDNS := &dns.Client{}
	resp, _, err := clientDNS.ExchangeWithConnContext(ctx, query, &dns.Conn{Conn: conn})
	if err != nil {
		log.Fatal(err)
	}

	// Print the responses
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			fmt.Printf("%s\n", a.A.String())
		}
	}

	// Explicitly close the connections
	cpool.Close()

	// Output:
	// 8.8.8.8
}
