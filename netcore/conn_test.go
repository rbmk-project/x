// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

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

func TestConnWrapper(t *testing.T) {
	t.Run("Close", func(t *testing.T) {

		// Helper function to create a standard test environment
		setup := func() (*bytes.Buffer, *mocks.Conn, *connWrapper, time.Time) {
			var buf bytes.Buffer
			fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			timeNow := func() time.Time {
				return fixedTime
			}

			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return a
				},
			}))

			mock := &mocks.Conn{
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{Logger: logger, TimeNow: timeNow},
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			return &buf, mock, wrapper, fixedTime
		}

		t.Run("successful close", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock behavior
			mock.MockClose = func() error {
				return nil
			}

			// Perform the close operation
			err := wrapper.Close()
			assert.NoError(t, err)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify closeStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeStart",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify closeDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeDone",
				"err":        nil,
				"errClass":   "",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t0":         fixedTime.Format(time.RFC3339Nano),
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("error on close", func(t *testing.T) {
			buf, mock, wrapper, fixedTime := setup()

			// Configure mock to return error
			expectedErr := errors.New("mocked close error")
			mock.MockClose = func() error {
				return expectedErr
			}

			// Perform the close operation
			err := wrapper.Close()
			assert.ErrorIs(t, err, expectedErr)

			// Verify logging output
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2)

			// Verify closeStart log
			var startLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[0]), &startLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeStart",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, startLog)

			// Verify closeDone log
			var doneLog map[string]interface{}
			err = json.Unmarshal([]byte(logs[1]), &doneLog)
			assert.NoError(t, err)
			assert.Equal(t, map[string]interface{}{
				"level":      "INFO",
				"msg":        "closeDone",
				"err":        expectedErr.Error(),
				"errClass":   "EGENERIC",
				"localAddr":  "127.0.0.1:1234",
				"protocol":   "tcp",
				"remoteAddr": "1.1.1.1:443",
				"t0":         fixedTime.Format(time.RFC3339Nano),
				"t":          fixedTime.Format(time.RFC3339Nano),
			}, doneLog)
		})

		t.Run("idempotent close", func(t *testing.T) {
			buf, mock, wrapper, _ := setup()

			closeCount := 0
			mock.MockClose = func() error {
				closeCount++
				return nil
			}

			// Close multiple times
			err1 := wrapper.Close()
			err2 := wrapper.Close()
			err3 := wrapper.Close()

			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.NoError(t, err3)
			assert.Equal(t, 1, closeCount, "Close should only be called once")

			// Verify we only logged one close operation
			logs := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Len(t, logs, 2, "Should only have one pair of start/done logs")
		})

		t.Run("no logger configured", func(t *testing.T) {
			mock := &mocks.Conn{
				MockClose: func() error {
					return nil
				},
				MockLocalAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
				},
				MockRemoteAddr: func() net.Addr {
					return &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 443}
				},
			}

			wrapper := &connWrapper{
				ctx:      context.Background(),
				conn:     mock,
				laddr:    "127.0.0.1:1234",
				netx:     &Network{}, // no logger configured
				protocol: "tcp",
				raddr:    "1.1.1.1:443",
			}

			err := wrapper.Close()
			assert.NoError(t, err)
		})
	})
}
