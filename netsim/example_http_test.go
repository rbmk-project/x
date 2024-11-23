// SPDX-License-Identifier: GPL-3.0-or-later

package netsim_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/rbmk-project/x/netsim"
)

// This example shows how to use [netsim] to simulate an HTTP
// server that listens for incoming cleartext requests.
func Example_http() {
	// Create a new scenario using the given directory to cache
	// the certificates used by the simulated PKI
	scenario := netsim.NewScenario("testdata")
	defer scenario.Close()

	// Create server stack running a HTTP-over-TCP server.
	//
	// This includes:
	//
	// 1. creating, attaching, and enabling routing for a server stack
	//
	// 2. registering the proper domain names and addresses
	//
	// 3. updating the PKI database to include the server's certificate
	scenario.Attach(scenario.MustNewStack(&netsim.StackConfig{
		DomainNames: []string{"dns.google"},
		Addresses:   []string{"8.8.8.8"},
		HTTPHandler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("Bonsoir, Elliot!\n"))
		}),
	}))

	// Create and attach the client stack.
	clientStack := scenario.MustNewStack(&netsim.StackConfig{
		Addresses: []string{"130.192.91.211"},
	})
	scenario.Attach(clientStack)

	// Create the HTTP client
	clientTxp := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := clientStack.DialContext(ctx, "tcp", addr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	defer clientTxp.CloseIdleConnections()
	clientHTTP := &http.Client{Transport: clientTxp}

	// Get the response body.
	resp, err := clientHTTP.Get("http://8.8.8.8/")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body
	fmt.Printf("%s", string(body))

	// Output:
	// Bonsoir, Elliot!
}
