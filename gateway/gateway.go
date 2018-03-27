package gateway

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"expvar"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/auxiliary"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/sd"
)

const (
	defaultName                    = "gateway"
	defaultRequestIDResponseHeader = "X-Request-ID"
	timeFormat                     = "2006-01-02T15:04:05+00:00"
)

var (
	gatewayExpvar = expvar.NewMap(defaultName)
)

type CheckURLHook func() (*url.URL, merry.Error)

func (h CheckURLHook) InvokeSafely() (u *url.URL, err merry.Error, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			exception = merry.Prepend(err1, "panic in check url hook")
			return
		}

		exception = merry.Errorf("panic in check url hook: \"%v\"", arg)
	}()

	u, err = h()

	return
}

type Gateway struct {
	Name             string        // The name of the Gateway for registration
	Addr             string        // TCP address to listen on, ":http" if empty
	HandleInterrupt  bool          // Should SIGINT and SIGTERM interrupts be handled?
	DisableKeepAlive bool          // Should TCP keep alive be disabled?
	GracePeriod      time.Duration // Timeout for graceful shutdown of open connections
	TLSConfig        *tls.Config   `json:"-"` // optional TLS config, used by ServeTLS

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
	// If nil then `httpx.Request.GenerateID` will be used.
	RequestIDGenerator httpx.StringExtractor `json:"-"`

	// Handlers define handlers to run on all request before
	// any other dispatch or validation.
	// Example uses would be rate limiting or authentication.
	Handlers []httpx.Handler `json:"-"`

	// HandlersTimeout is the maximum amout of time to wait for
	// all gateway-level handlers to complete.
	// If the timeout is exceeded the entire request is aborted.
	HandlersTimeout time.Duration

	// InternalServerErrorHandler optionally customizes the
	// response returned to the user agent when the gateway
	// encounters an error trying to service the requst before
	// the corresponding endpoint has been determined.
	// If nil the default handler will return a 500 status code
	// with an empty body.
	InternalServerErrorHandler httpx.ErrorHandler `json:"-"`

	// NotFoundHandler optionally customizes the response
	// returned to the user agent when no endpoint is configured
	// service a request path.
	// If nil the default handler will return a 404 status code
	// with an empty body.
	NotFoundHandler httpx.Handler `json:"-"`

	// Registrar implements sd.Registrar and registers
	// the gateway service with a service registry, using the Gateway's
	// `Name` field. If nil, no registration occurs.
	Registrar sd.Registrar

	// CheckURLHook provides the *url.URL to be used in
	// the Registrar's `AddCheck` method. If `Registrar` is nil,
	// this hook is not called. If this hook is nil, no check is registered
	// via `AddCheck`.
	CheckURLHook CheckURLHook

	// ErrorHook optionally customizes how errors encountered
	// servicing a request are disposed.
	// If nil the request id and error are sent to the standard
	// library `log.Println` function.
	ErrorHook httpx.ErrorHook `json:"-"`

	// CompletionHook optionally customizes the behavior after
	// a request has been serviced.
	// If nil no action will be taken.
	CompletionHook httpx.CompletionHook `json:"-"`

	base     http.Server
	listener net.Listener
	tree     *node

	started bool
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

	g.base.Addr = g.Addr
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

	if g.Name == "" {
		g.Name = defaultName
	}

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
		signal.Stop(interrupt)
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
