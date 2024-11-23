// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/rbmk-project/x/connpool"
	"github.com/rbmk-project/x/netsim"
	"github.com/rbmk-project/x/netsim/simpki"
)

// This example shows how to use [netsim] to simulate an HTTPS
// server that listens for incoming encrypted requests.
func Example_https() {
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

	// Create a PKI for the server and obtain the certificate.
	pki := simpki.MustNew("testdata")
	serverCert := pki.MustNewCert(&simpki.Config{
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

	// Create the HTTP server.
	serverEndpoint := netip.AddrPortFrom(serverAddr, 443)
	listener, err := serverStack.Listen(ctx, "tcp", serverEndpoint.String())
	if err != nil {
		log.Fatal(err)
	}
	cpool.Add(listener)
	serverHTTP := &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("Bonsoir, Elliot!\n"))
		}),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{serverCert},
		},
	}
	go serverHTTP.ServeTLS(listener, "", "")
	cpool.Add(serverHTTP)

	// Create the HTTP client
	clientTxp := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := clientStack.DialContext(ctx, "tcp", addr)
			if err != nil {
				return nil, err
			}
			cpool.Add(conn)
			return conn, nil
		},
		TLSClientConfig: &tls.Config{
			RootCAs: pki.CertPool(),
		},
	}
	clientHTTP := &http.Client{Transport: clientTxp}

	// Get the response body.
	resp, err := clientHTTP.Get("https://8.8.8.8/")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed: %d", resp.StatusCode)
	}
	cpool.Add(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body
	fmt.Printf("%s", string(body))

	// Explicitly close the connections
	cpool.Close()

	// Output:
	// Bonsoir, Elliot!
}
