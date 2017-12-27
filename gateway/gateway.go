package gateway

import (
	"crypto/tls"
	"expvar"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/service"
)

const (
	defaultRequestIDResponseHeader = "X-Request-ID"
)

var (
	stats = expvar.NewMap("gateway")
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

	// Logger optionally specifies the logger to use by the
	// Gateway and all of its services.
	// If nil all logging is disabled.
	Logger *zap.Logger

	// RequestIDHeaderName optionally customizes the name of the
	// response header for the request id.
	// If empty "X-Request-Id" will be used.
	RequestIDHeaderName string

	// RequestIDGenerator optionally customizes how request ids
	// are generated.
	// If nil then `service.Request.GenerateID` will be used.
	RequestIDGenerator service.StringExtractor

	// NotFoundHandler optionally customizes the response
	// returned to the user agent when no endpoint is configured
	// service a request path.
	// If nil the default handler will return a 404 status code
	// with an empty body.
	NotFoundHandler service.Handler

	base        http.Server
	auxiliaries []auxillary.Server
	tree        *node

	requestLog *zap.Logger
	panicLog   *zap.Logger
	started    bool
}

func (g *Gateway) init() {
	g.started = true
	stats = stats.Init()
	g.base.Addr = g.Address
	g.base.TLSConfig = g.TLSConfig
	g.base.ReadTimeout = g.ReadTimeout
	g.base.ReadHeaderTimeout = g.ReadHeaderTimeout
	g.base.WriteTimeout = g.WriteTimeout
	g.base.IdleTimeout = g.IdleTimeout
	g.base.MaxHeaderBytes = g.MaxHeaderBytes
	g.base.TLSNextProto = g.TLSNextProto
	g.base.ConnState = connstate
	g.base.Handler = g

	if g.DisableKeepAlive {
		g.base.SetKeepAlivesEnabled(false)
	}

	// xxx - need to handle early shutdown before startup is completed
	if g.HandleInterrupt {
		interrupt := make(chan os.Signal, 1)
		go g.handleInterrupt(interrupt)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	}

	if g.RequestIDHeaderName == "" {
		g.RequestIDHeaderName = defaultRequestIDResponseHeader
	}

	if g.RequestIDGenerator == nil {
		g.RequestIDGenerator = defaultRequestIDGenerator
	}

	if g.NotFoundHandler == nil {
		g.NotFoundHandler = defaultNotFoundHandler
	}

	if g.Logger == nil {
		g.Logger = zap.NewNop()
	}
	g.requestLog = g.Logger.Named("request")
	g.panicLog = g.Logger.Named("panic")

	g.tree = new(node)
}

func connstate(con net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		stats.Add("total_connections", 1)
		stats.Add("connected", 1)
	case http.StateClosed, http.StateHijacked:
		stats.Add("connected", -1)
	}
}

func (g *Gateway) handleInterrupt(interrupt chan os.Signal) {
	select {
	case <-interrupt:
		g.Logger.Info("interrupt received!")
		g.Shutdown()
	}
}
