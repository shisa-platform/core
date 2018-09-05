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

	"github.com/shisa-platform/core/auxiliary"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/sd"
)

const (
	defaultName                    = "gateway"
	defaultRequestIDResponseHeader = "X-Request-ID"
	timeFormat                     = "2006-01-02T15:04:05+00:00"
)

var (
	gatewayExpvar = expvar.NewMap(defaultName)
)

type URLHook func() (*url.URL, merry.Error)

func (h URLHook) InvokeSafely() (u *url.URL, err merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in check url hook")

	u, err = h()

	return
}

type Gateway struct {
	Name             string        // The name of the Gateway for registration
	Addr             string        // TCP address to listen on, ":http" if empty
	HandleInterrupt  bool          // Should SIGINT and SIGTERM interrupts be handled?
	DisableKeepAlive bool          // Should TCP keep alive be disabled?
	GracePeriod      time.Duration // Timeout for graceful shutdown of open connections
	TLSConfig        *tls.Config   // optional TLS config, used by ServeTLS

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

	// RequestIDHeaderName optionally customizes the name of the
	// response header for the request id.
	// If empty "X-Request-Id" will be used.
	RequestIDHeaderName string

	// RequestIDGenerator optionally customizes how request ids
	// are generated.
	// If nil then `httpx.Request.GenerateID` will be used.
	RequestIDGenerator httpx.StringExtractor

	// Handlers define handlers to run on all request before
	// any other dispatch or validation.
	// Example uses would be rate limiting or authentication.
	Handlers []httpx.Handler

	// HandlersTimeout is the maximum amount of time to wait for
	// all gateway-level handlers to complete.
	// If the timeout is exceeded the entire request is aborted.
	HandlersTimeout time.Duration

	// InternalServerErrorHandler optionally customizes the
	// response returned to the user agent when the gateway
	// encounters an error trying to service the requst before
	// the corresponding endpoint has been determined.
	// If nil the default handler will return a 500 status code
	// with an empty body.
	InternalServerErrorHandler httpx.ErrorHandler

	// NotFoundHandler optionally customizes the response
	// returned to the user agent when no endpoint is configured
	// service a request path.
	// If nil the default handler will return a 404 status code
	// with an empty body.
	NotFoundHandler httpx.Handler

	// Registrar implements sd.Registrar and registers
	// the gateway service with a service registry, using the
	// Gateway's `Name` field. If nil, no registration occurs.
	Registrar sd.Registrar

	// RegistrationURLHook provides the *url.URL to be used in
	// the Registrar's `Register` method. If `Registrar` is nil,
	// this hook is not called. If this hook is nil, no service is
	// registered via `Register`.
	RegistrationURLHook URLHook

	// CheckURLHook provides the *url.URL to be used in
	// the Registrar's `AddCheck` method. If `Registrar` is nil,
	// this hook is not called. If this hook is nil, no check is
	// registered via `AddCheck`.
	CheckURLHook URLHook

	// ErrorHook optionally customizes how errors encountered
	// servicing a request are disposed.
	// If nil the request id and error are sent to the standard
	// library `log.Println` function.
	ErrorHook httpx.ErrorHook

	// CompletionHook optionally customizes the behavior after
	// a request has been serviced.
	// If nil no action will be taken.
	CompletionHook httpx.CompletionHook

	base      http.Server
	listener  net.Listener
	tree      *node
	started   bool
	interrupt chan os.Signal
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

	if g.RequestIDHeaderName == "" {
		g.RequestIDHeaderName = defaultRequestIDResponseHeader
	}

	if g.Name == "" {
		g.Name = defaultName
	}

	g.tree = new(node)

	if g.HandleInterrupt {
		g.interrupt = make(chan os.Signal, 1)
		go g.handleInterrupt()
		signal.Notify(g.interrupt, syscall.SIGINT, syscall.SIGTERM)
	}
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

func (g *Gateway) handleInterrupt() {
	select {
	case sig := <-g.interrupt:
		if sig != syscall.SIGINT && sig != syscall.SIGTERM {
			return
		}
		signal.Stop(g.interrupt)
		g.Shutdown()
	}
}

// String implements `expvar.Var.String`
func (g *Gateway) String() string {
	repr := map[string]interface{}{
		"Name":              g.Name,
		"Addr":              g.Address(),
		"HandleInterrupt":   g.HandleInterrupt,
		"DisableKeepAlive":  g.DisableKeepAlive,
		"GracePeriod":       g.GracePeriod.String(),
		"ReadTimeout":       g.ReadTimeout.String(),
		"ReadHeaderTimeout": g.ReadHeaderTimeout.String(),
		"WriteTimeout":      g.WriteTimeout.String(),
		"IdleTimeout":       g.IdleTimeout.String(),
		"MaxHeaderBytes":    g.MaxHeaderBytes,
	}

	if g.TLSConfig == nil {
		repr["TLSConfig"] = "unset"
	} else {
		repr["TLSConfig"] = "configured"
	}
	if len(g.TLSNextProto) == 0 {
		repr["TLSNextProto"] = "unset"
	} else {
		repr["TLSNextProto"] = "configured"
	}

	repr["RequestIDHeaderName"] = g.RequestIDHeaderName
	if g.RequestIDGenerator == nil {
		repr["RequestIDGenerator"] = "unset"
	} else {
		repr["RequestIDGenerator"] = "configured"
	}

	repr["Handlers"] = len(g.Handlers)
	repr["HandlersTimeout"] = g.HandlersTimeout.String()

	if g.InternalServerErrorHandler == nil {
		repr["InternalServerErrorHandler"] = "unset"
	} else {
		repr["InternalServerErrorHandler"] = "configured"
	}
	if g.NotFoundHandler == nil {
		repr["NotFoundHandler"] = "unset"
	} else {
		repr["NotFoundHandler"] = "configured"
	}
	if g.Registrar == nil {
		repr["Registrar"] = "unset"
	} else {
		repr["Registrar"] = "configured"
	}
	if g.CheckURLHook == nil {
		repr["CheckURLHook"] = "unset"
	} else {
		repr["CheckURLHook"] = "configured"
	}
	if g.ErrorHook == nil {
		repr["ErrorHook"] = "unset"
	} else {
		repr["ErrorHook"] = "configured"
	}
	if g.CompletionHook == nil {
		repr["CompletionHook"] = "unset"
	} else {
		repr["CompletionHook"] = "configured"
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(repr)
	return buf.String()
}
