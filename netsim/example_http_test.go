// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/rbmk-project/x/connpool"
	"github.com/rbmk-project/x/netsim"
)

// This example shows how to use [netsim] to simulate an HTTP
// server that listens for incoming cleartext requests.
func Example_http() {
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

	// Create the HTTP server.
	serverEndpoint := netip.AddrPortFrom(serverAddr, 80)
	listener, err := serverStack.Listen(ctx, "tcp", serverEndpoint.String())
	if err != nil {
		log.Fatal(err)
	}
	cpool.Add(listener)
	serverHTTP := &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("Bonsoir, Elliot!\n"))
		}),
	}
	go serverHTTP.Serve(listener)
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
	}
	clientHTTP := &http.Client{Transport: clientTxp}

	// Get the response body.
	resp, err := clientHTTP.Get("http://8.8.8.8/")
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
