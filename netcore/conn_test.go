// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/rbmk-project/common/mocks"
	"github.com/stretchr/testify/assert"
)

func TestConnLocalAddr(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		addr := connLocalAddr(nil)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("nil local address", func(t *testing.T) {
		conn := &mocks.Conn{
			MockLocalAddr: func() net.Addr { return nil },
		}
		addr := connLocalAddr(conn)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("valid address", func(t *testing.T) {
		expectedAddr := &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 1234,
		}
		conn := &mocks.Conn{
			MockLocalAddr: func() net.Addr { return expectedAddr },
		}
		addr := connLocalAddr(conn)
		assert.Equal(t, expectedAddr, addr)
	})
}

func TestConnRemoteAddr(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		addr := connRemoteAddr(nil)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("nil remote address", func(t *testing.T) {
		conn := &mocks.Conn{
			MockRemoteAddr: func() net.Addr { return nil },
		}
		addr := connRemoteAddr(conn)
		assert.Equal(t, "", addr.Network())
		assert.Equal(t, "", addr.String())
	})

	t.Run("valid address", func(t *testing.T) {
		expectedAddr := &net.TCPAddr{
			IP:   net.ParseIP("1.1.1.1"),
			Port: 443,
		}
		conn := &mocks.Conn{
			MockRemoteAddr: func() net.Addr { return expectedAddr },
		}
		addr := connRemoteAddr(conn)
		assert.Equal(t, expectedAddr, addr)
	})
}

func TestMaybeWrapConn(t *testing.T) {
	t.Run("nil connection", func(t *testing.T) {
		nx := &Network{}
		assert.Nil(t, nx.maybeWrapConn(context.Background(), nil))
	})

	t.Run("no logger configured", func(t *testing.T) {
		nx := &Network{}
		conn := &mocks.Conn{}
		wrapped := nx.maybeWrapConn(context.Background(), conn)
		assert.Equal(t, conn, wrapped) // should return unwrapped
	})

	t.Run("no wrapper configured", func(t *testing.T) {
		nx := &Network{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		}
		conn := &mocks.Conn{}
		wrapped := nx.maybeWrapConn(context.Background(), conn)
		assert.Equal(t, conn, wrapped) // should return unwrapped
	})

	t.Run("full wrapping", func(t *testing.T) {
		nx := &Network{
			Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
			WrapConn: WrapConn,
		}
		conn := &mocks.Conn{
			MockLocalAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}
			},
			MockRemoteAddr: func() net.Addr {
				return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
			},
		}
		wrapped := nx.maybeWrapConn(context.Background(), conn)
		assert.NotEqual(t, conn, wrapped) // should return wrapped
		assert.IsType(t, &connWrapper{}, wrapped)
	})
}
