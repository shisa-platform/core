package httpx

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPListenerForAddress(t *testing.T) {
	listener, err := HTTPListenerForAddress(":0")
	assert.NoError(t, err)
	assert.NotNil(t, listener)

	assert.NoError(t, listener.Close())
}

func TestHTTPListenerForAddressDefault(t *testing.T) {
	listener, err := HTTPListenerForAddress("")
	assert.Error(t, err)
	assert.Nil(t, listener)
}

func TestHTTPListenerForAddressBadPort(t *testing.T) {
	listener, err := HTTPListenerForAddress(":25")
	assert.Error(t, err)
	assert.Nil(t, listener)
}

func TestAccept(t *testing.T) {
	listener, err := HTTPListenerForAddress(":0")
	assert.NoError(t, err)
	assert.NotNil(t, listener)

	var conn net.Conn
	go func() {
		var listenErr error
		conn, listenErr = listener.Accept()
		assert.NoError(t, listenErr)
		assert.NotNil(t, conn)
	}()

	addr := listener.Addr()
	timeout := time.Millisecond * 250
	client, err1 := net.DialTimeout(addr.Network(), addr.String(), timeout)
	assert.NoError(t, err1)
	assert.NotNil(t, client)

	listener.Close()
	if client != nil {
		assert.NoError(t, client.Close())
	}
	if conn != nil {
		assert.NoError(t, conn.Close())
	}
}

func TestAcceptError(t *testing.T) {
	listener, err := HTTPListenerForAddress(":0")
	assert.NoError(t, err)
	assert.NotNil(t, listener)

	tcp := listener.(tcpKeepAliveListener)
	tcp.SetDeadline(time.Now().Add((time.Millisecond * 250)))

	conn, listenErr := listener.Accept()
	assert.Error(t, listenErr)
	assert.Nil(t, conn)
}
