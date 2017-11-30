package gateway

import (
	"crypto/tls"
	"errors"
	"expvar"
	"net/http"
	"time"

	"github.com/percolate/shisa/server"
)

var (
	stats               = expvar.NewMap("gateway")
	AlreadyStartedError = errors.New("Server already started, cannot register auxillary server.")
)

type Gateway struct {
	Name             string        // The name of the Gateway for in logging
	Address          string        // TCP address to listen on, ":http" if empty
	Trace            bool          // Should trace-level logging be enabled?
	HandleInterrupt  bool          // Should SIGINT and SIGTERM interrupts be handled?
	DisableKeepAlive bool          // Should TCP keep alive be disabled?
	GracePeriod      time.Duration // Timeout for graceful shutdown of open connections
	TLSConfig        *tls.Config   // optional TLS config, used by ServeTLS and ListenAndServeTLS

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int

	// TLSNextProto optionally specifies a function to take over
	// ownership of the provided TLS connection when an NPN/ALPN
	// protocol upgrade has occurred. The map key is the protocol
	// name negotiated. The Handler argument should be used to
	// handle HTTP requests and will initialize the Request's TLS
	// and RemoteAddr if not already set. The connection is
	// automatically closed when the function returns.
	// If TLSNextProto is not nil, HTTP/2 support is not enabled
	// automatically.
	TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler)

	base    http.Server
	aux     []server.Server
	started bool
}
