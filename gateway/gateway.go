package gateway

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"expvar"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/auxiliary"
	"github.com/percolate/shisa/service"
)

const (
	defaultRequestIDResponseHeader = "X-Request-ID"
	timeFormat                     = "2006-01-02T15:04:05+00:00"
)

var (
	gatewayExpvar = expvar.NewMap("gateway")
)

type Gateway struct {
	Name             string        // The name of the Gateway for in logging
	Address          string        // TCP address to listen on, ":http" if empty
	HandleInterrupt  bool          // Should SIGINT and SIGTERM interrupts be handled?
	DisableKeepAlive bool          // Should TCP keep alive be disabled?
	GracePeriod      time.Duration // Timeout for graceful shutdown of open connections
	TLSConfig        *tls.Config   `json:"-"` // optional TLS config, used by ServeTLS and ListenAndServeTLS

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
	TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler) `json:"-"`

	// RequestIDHeaderName optionally customizes the name of the
	// response header for the request id.
	// If empty "X-Request-Id" will be used.
	RequestIDHeaderName string

	// RequestIDGenerator optionally customizes how request ids
	// are generated.
	// If nil then `service.Request.GenerateID` will be used.
	RequestIDGenerator service.StringExtractor `json:"-"`

	// Handlers define handlers to run on all request before
	// any other dispatch or validation.
	// Example uses would be rate limiting or authentication.
	Handlers []service.Handler `json:"-"`

	// InternalServerErrorHandler optionally customizes the
	// response returned to the user agent when the gateway
	// encounters an error trying to service the requst before
	// the corresponding endpoint has been determined.
	// If nil the default handler will return a 500 status code
	// with an empty body.
	InternalServerErrorHandler service.ErrorHandler `json:"-"`

	// NotFoundHandler optionally customizes the response
	// returned to the user agent when no endpoint is configured
	// service a request path.
	// If nil the default handler will return a 404 status code
	// with an empty body.
	NotFoundHandler service.Handler `json:"-"`

	// Register is a function hook that optionally registers
	// the gateway service with a service registry
	Register func(name, addr string) merry.Error

	// Logger optionally specifies the logger to use by the
	// Gateway.
	// If nil all logging is disabled.
	Logger *zap.Logger `json:"-"`

	base        http.Server
	auxiliaries []auxiliary.Server
	tree        *node

	requestLog *zap.Logger
	panicLog   *zap.Logger
	started    bool
}

func (g *Gateway) init() {
	start := time.Now().UTC()

	gatewayExpvar = gatewayExpvar.Init()
	startTime := new(expvar.String)
	startTime.Set(start.Format(timeFormat))
	gatewayExpvar.Set("start-time", startTime)
	gatewayExpvar.Set("uptime", expvar.Func(func() interface{} {
		now := time.Now().UTC()
		return now.Sub(start).String()
	}))
	gatewayExpvar.Set("settings", g)
	gatewayExpvar.Set("auxiliary", auxiliary.AuxiliaryStats)

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

	if g.InternalServerErrorHandler == nil {
		g.InternalServerErrorHandler = defaultInternalServerErrorHandler
	}

	if g.Register == nil {
		g.Register = func(name, addr string) merry.Error {
			return nil
		}
	}

	if g.Logger == nil {
		g.Logger = zap.NewNop()
	}
	g.requestLog = g.Logger.Named("request")
	g.panicLog = g.Logger.Named("panic")

	g.tree = new(node)
	g.started = true
}

func connstate(con net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		gatewayExpvar.Add("total_connections", 1)
		gatewayExpvar.Add("connected", 1)
	case http.StateClosed, http.StateHijacked:
		gatewayExpvar.Add("connected", -1)
	}
}

func (g *Gateway) handleInterrupt(interrupt chan os.Signal) {
	select {
	case <-interrupt:
		g.Logger.Info("interrupt received!")
		g.Shutdown()
	}
}

// String implements `expvar.Var.String`
func (g *Gateway) String() string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(g)
	return buf.String()
}
