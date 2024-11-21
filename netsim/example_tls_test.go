// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"

	"github.com/rbmk-project/x/connpool"
	"github.com/rbmk-project/x/netsim"
	"github.com/rbmk-project/x/netsim/simpki"
)

// This example shows how to use [netsim] to simulate a TLS
// server that listens for incoming encrypted requests.
func Example_tls() {
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

	// Create the server listener
	serverEndpoint := netip.AddrPortFrom(serverAddr, 443)
	listener, err := serverStack.Listen(ctx, "tcp", serverEndpoint.String())
	if err != nil {
		log.Fatal(err)
	}
	cpool.Add(listener)

	// Create a PKI for the server and obtain the certificate.
	pki := simpki.MustNewPKI("testdata")
	serverCert := pki.MustNewCert(&simpki.PKICertConfig{
		CommonName: "dns.google",
		DNSNames: []string{
			"dns.google.com",
			"dns.google",
		},
		IPAddrs: []net.IP{
			net.IPv4(8, 8, 8, 8),
			net.IPv4(8, 8, 4, 4),
		},
	})

	// Start the server in the background.
	go func() {
		clp := connpool.New()
		defer clp.Close()

		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		clp.Add(conn)

		tconn := tls.Server(conn, &tls.Config{
			Certificates: []tls.Certificate{serverCert},
		})
		clp.Add(tconn)
		if err := tconn.HandshakeContext(ctx); err != nil {
			log.Fatal(err)
		}

		if _, err := tconn.Write([]byte("Bonsoir, Elliot!\n")); err != nil {
			log.Fatal(err)
		}
	}()

	// Connect to the server
	conn, err := clientStack.DialContext(ctx, "tcp", serverEndpoint.String())
	if err != nil {
		log.Fatal(err)
	}
	cpool.Add(conn)

	// Perform the TLS handshake
	tconn := tls.Client(conn, &tls.Config{
		RootCAs:    pki.CertPool(),
		ServerName: "dns.google",
	})
	cpool.Add(tconn)
	if err := tconn.HandshakeContext(ctx); err != nil {
		log.Fatal(err)
	}

	// Get the response body.
	buffer := make([]byte, 1024)
	count, err := tconn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body
	fmt.Printf("%s", string(buffer[:count]))

	// Output:
	// Bonsoir, Elliot!
}
