// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
	"github.com/rbmk-project/x/netsim"
)

// This example shows how to use [netsim] to simulate a DNS
// server that listens for incoming requests over UDP.
func Example_dnsOverUDP() {
	// Create a new scenario using the given directory to cache
	// the certificates used by the simulated PKI
	scenario := netsim.NewScenario("testdata")
	defer scenario.Close()

	// Create server stack emulating dns.google.
	//
	// This includes:
	//
	// 1. creating, attaching, and enabling routing for a server stack
	//
	// 2. registering the proper domain names and addresses
	//
	// 3. updating the PKI database to include the server's certificate
	scenario.Attach(scenario.MustNewGoogleDNSStack())

	// Create and attach the client stack.
	clientStack := scenario.MustNewClientStack()
	scenario.Attach(clientStack)

	// Create a context with a watchdog timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create the client connection with the DNS server.
	conn, err := clientStack.DialContext(ctx, "udp", "8.8.8.8:53")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

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

	// Output:
	// 8.8.8.8
}
