package gateway

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGatewayInit(t *testing.T) {
	config := tls.Config{}
	nextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	cut := &Gateway{
		Address:           ":9001",
		DisableKeepAlive:  true,
		TLSConfig:         &config,
		ReadTimeout:       time.Millisecond * 5,
		ReadHeaderTimeout: time.Millisecond * 10,
		WriteTimeout:      time.Millisecond * 15,
		IdleTimeout:       time.Millisecond * 20,
		MaxHeaderBytes:    1024,
		TLSNextProto:      nextProto,
	}
	cut.init()

	assert.Equal(t, cut.Address, cut.base.Addr)
	assert.Equal(t, cut.TLSConfig, cut.base.TLSConfig)
	assert.Equal(t, cut.ReadTimeout, cut.base.ReadTimeout)
	assert.Equal(t, cut.ReadHeaderTimeout, cut.base.ReadHeaderTimeout)
	assert.Equal(t, cut.WriteTimeout, cut.base.WriteTimeout)
	assert.Equal(t, cut.IdleTimeout, cut.base.IdleTimeout)
	assert.Equal(t, cut.MaxHeaderBytes, cut.base.MaxHeaderBytes)
	assert.Equal(t, cut.TLSNextProto, cut.base.TLSNextProto)
	assert.NotNil(t, cut.base.ConnState)
	assert.Equal(t, cut, cut.base.Handler)
	assert.Equal(t, defaultRequestIDResponseHeader, cut.RequestIDHeaderName)
	assert.NotNil(t, cut.RequestIDGenerator)
	assert.NotNil(t, cut.NotFoundHandler)
	assert.NotNil(t, cut.Logger)
}
