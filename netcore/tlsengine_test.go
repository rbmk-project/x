// SPDX-License-Identifier: GPL-3.0-or-later

package netcore

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// tlsEngineMock is a mocked [TLSEngine] used for testing.
type tlsEngineMock struct{}

var _ TLSEngine = &tlsEngineMock{}

// Name implements [TLSEngine].
func (t *tlsEngineMock) Name() string {
	panic("unimplemented")
}

// NewClientConn implements [TLSEngine].
func (t *tlsEngineMock) NewClientConn(conn net.Conn, config *tls.Config) TLSConn {
	panic("unimplemented")
}

// Parrot implements [TLSEngine].
func (t *tlsEngineMock) Parrot() string {
	panic("unimplemented")
}

func TestNetwork_newTLSEngine(t *testing.T) {
	t.Run("when we have a custom TLSEngine", func(t *testing.T) {
		nx := &Network{
			TLSEngine: &tlsEngineMock{},
		}
		engine := nx.newTLSEngine()
		_, ok := engine.(*tlsEngineMock)
		assert.True(t, ok)
	})

	t.Run("when we have a custom NewTLSClientConn", func(t *testing.T) {
		nx := &Network{
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				panic("unimplemented")
			},
		}
		engine := nx.newTLSEngine()
		name := engine.Name()
		assert.Equal(t, name, "unknown")
	})

	t.Run("when we have both TLSEngine and NewTLSClientConn", func(t *testing.T) {
		nx := &Network{
			TLSEngine: &tlsEngineMock{},
			NewTLSClientConn: func(conn net.Conn, config *tls.Config) TLSConn {
				panic("unimplemented")
			},
		}
		engine := nx.newTLSEngine()
		_, ok := engine.(*tlsEngineMock)
		assert.True(t, ok)
	})

	t.Run("when nothing has been configured", func(t *testing.T) {
		nx := &Network{}
		engine := nx.newTLSEngine()
		_, ok := engine.(*TLSEngineStdlib)
		assert.True(t, ok)
	})
}
