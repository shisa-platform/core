package httpx

import (
	"net"
	"time"

	"github.com/ansel1/merry"
)

// HTTPListenerForAddress returns a TCP listener for the given
// address.
// If the address is empty `":http"` is used.
func HTTPListenerForAddress(addr string) (net.Listener, merry.Error) {
	if addr == "" {
		addr = ":http"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, merry.Prepend(err, "open HTTP listener")
	}

	return tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS
// so dead TCP connections (e.g. closing laptop mid-download)
// eventually go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (l tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := l.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)

	return tc, nil
}
