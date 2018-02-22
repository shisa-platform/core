package auxiliary

import (
	stdctx "context"
	"crypto/tls"
	"expvar"
	"net"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/service"
)

const (
	defaultRequestIDResponseHeader = "X-Request-ID"
	startTimeFormat                = "2006-01-02T15:04:05+00:00"
)

var (
	AuxiliaryStats = new(expvar.Map)
)

type HTTPServer struct {
	Addr             string // TCP address to listen on, ":http" if empty
	UseTLS           bool   // should this server use TLS?
	DisableKeepAlive bool   // Should TCP keep alive be disabled?

	// TLSConfig is optional TLS configuration,
	// This must be non-nil and properly initialized if `UseTLS`
	// is `true`.
	TLSConfig *tls.Config

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
	// If nil then `service.Request.GenerateID` will be used.
	RequestIDGenerator service.StringExtractor

	// Authentication optionally enforces authentication before
	// other request handling.  This is recommended to prevent
	// leaking details about the implementation to unknown user
	// agents.
	Authentication *middleware.Authentication

	// Authorizer optionally enforces authentication before other
	// request handling.  Use of this field requires the
	// `Authentication` field to be configured and return a
	// principal.
	Authorizer Authorizer

	// Router is called by the `ServeHTTP` method to find the
	// correct handler to invoke for the current request.
	// If nil is returned a 404 status code with an empty body is
	// returned to the user agent.
	Router func(context.Context, *service.Request) service.Handler

	// ErrorHook optionally customizes how errors encountered
	// servicing a request are disposed.
	// If nil no action will be taken.
	ErrorHook func(context.Context, *service.Request, merry.Error)

	// CompletionHook optionally customizes the behavior after
	// a request has been serviced.
	// If nil no action will be taken.
	CompletionHook func(context.Context, *service.Request, httpx.ResponseSnapshot)

	// Logger optionally specifies the logger to use by the Debug
	// server.
	// If nil all logging is disabled.
	Logger *zap.Logger

	base     http.Server
	listener net.Listener
}

func (s *HTTPServer) init() {
	s.base.Addr = s.Addr
	s.base.TLSConfig = s.TLSConfig
	s.base.ReadTimeout = s.ReadTimeout
	s.base.ReadHeaderTimeout = s.ReadHeaderTimeout
	s.base.WriteTimeout = s.WriteTimeout
	s.base.IdleTimeout = s.IdleTimeout
	s.base.MaxHeaderBytes = s.MaxHeaderBytes
	s.base.TLSNextProto = s.TLSNextProto

	s.base.Handler = s

	if s.DisableKeepAlive {
		s.base.SetKeepAlivesEnabled(false)
	}

	if s.RequestIDHeaderName == "" {
		s.RequestIDHeaderName = defaultRequestIDResponseHeader
	}

	if s.RequestIDGenerator == nil {
		s.RequestIDGenerator = s.generateRequestID
	}

	if s.Router == nil {
		s.Router = s.route
	}

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
}

func (s *HTTPServer) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}

	return s.Addr
}

func (s *HTTPServer) Authenticate(ctx context.Context, request *service.Request) (response service.Response) {
	if s.Authentication == nil {
		return
	}

	if response = s.Authentication.Service(ctx, request); response != nil {
		return
	}
	if s.Authorizer != nil {
		if ok, err := s.Authorizer.Authorize(ctx, request); err != nil {
			err = err.WithHTTPCode(http.StatusUnauthorized)
			return s.Authentication.ErrorHandler(ctx, request, err)
		} else if !ok {
			return s.Authentication.UnauthorizedHandler(ctx, request)
		}
	}

	return
}

func (s *HTTPServer) Serve() error {
	s.init()
	defer s.Logger.Sync()

	listener, err := httpx.HTTPListenerForAddress(s.Addr)
	if err != nil {
		return err
	}
	s.listener = listener

	if s.UseTLS {
		return s.base.ServeTLS(listener, "", "")
	}

	return s.base.Serve(listener)
}

func (s *HTTPServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(stdctx.Background(), timeout)
	defer cancel()
	s.listener = nil

	return merry.Wrap(s.base.Shutdown(ctx))
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ri := httpx.NewInterceptor(w)

	ctx := context.New(r.Context())
	request := &service.Request{Request: r}

	requestID, idErr := s.RequestIDGenerator(ctx, request)
	if idErr != nil {
		idErr = merry.WithMessage(idErr, "generating request id")
		requestID = request.ID()
	}
	if requestID == "" {
		idErr = merry.New("generator returned empty request id")
		requestID = request.ID()
	}

	ctx = context.WithRequestID(ctx, requestID)
	ri.Header().Set(s.RequestIDHeaderName, requestID)

	var response service.Response
	if response = s.Authenticate(ctx, request); response != nil {
		goto finish
	}

	if handler := s.Router(ctx, request); handler != nil {
		response = handler(ctx, request)
	} else {
		response = service.NewEmpty(http.StatusNotFound)
		response.Headers().Set("Content-Type", "text/plain; charset=utf-8")
	}

finish:
	writeErr := ri.WriteResponse(response)
	snapshot := ri.Flush()

	if s.CompletionHook != nil {
		s.CompletionHook(ctx, request, snapshot)
	}

	if idErr != nil && s.ErrorHook != nil {
		s.ErrorHook(ctx, request, idErr)
	}
	writeErr1 := merry.WithMessage(writeErr, "serializing response")
	if writeErr1 != nil && s.ErrorHook != nil {
		s.ErrorHook(ctx, request, writeErr1)
	}
	respErr := response.Err()
	if respErr != nil && s.ErrorHook != nil {
		s.ErrorHook(ctx, request, merry.WithMessage(respErr, "handler failed"))
	}
}

func (s *HTTPServer) generateRequestID(c context.Context, r *service.Request) (string, merry.Error) {
	return r.ID(), nil
}

func (s *HTTPServer) route(context.Context, *service.Request) service.Handler {
	return nil
}
