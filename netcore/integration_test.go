// SPDX-License-Identifier: GPL-3.0-or-later

package netcore_test

import (
	"context"
	"testing"
	"time"

	"github.com/rbmk-project/x/netcore"
)

func TestDialerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}

	netx := netcore.NewNetwork()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := netx.DialContext(ctx, "tcp", "example.com:80")
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
}

func TestTLSDialerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}

	netx := netcore.NewNetwork()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := netx.DialTLSContext(ctx, "tcp", "example.com:443")
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
}
