package auxiliary

import (
	stdctx "context"
	"crypto/tls"
	"expvar"
	"net"
	"net/http"
	"time"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/middleware"
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
	// If nil then `httpx.Request.GenerateID` will be used.
	RequestIDGenerator httpx.StringExtractor

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
	Router Router

	// ErrorHook optionally customizes how errors encountered
	// servicing a request are disposed.
	// If nil no action will be taken.
	ErrorHook httpx.ErrorHook

	// CompletionHook optionally customizes the behavior after
	// a request has been serviced.
	// If nil no action will be taken.
	CompletionHook httpx.CompletionHook

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
}

func (s *HTTPServer) Address() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}

	return s.Addr
}

func (s *HTTPServer) Authenticate(ctx context.Context, request *httpx.Request) (response httpx.Response) {
	if s.Authentication == nil {
		return
	}

	if response = s.Authentication.Service(ctx, request); response != nil {
		return
	}

	return s.Authorize(ctx, request)
}

func (s *HTTPServer) Authorize(ctx context.Context, request *httpx.Request) (response httpx.Response) {
	if s.Authorizer == nil {
		return
	}

	if ok, err := s.authn(ctx, request); err != nil {
		err = err.WithHTTPCode(http.StatusForbidden)
		response = s.Authentication.HandleError(ctx, request, err)
	} else if !ok {
		response = s.Authentication.HandleUnauthorized(ctx, request)
	}

	return
}

func (s *HTTPServer) authn(ctx context.Context, request *httpx.Request) (ok bool, err merry.Error) {
	defer errorx.CapturePanic(&err, "panic in auxiliary authorizer")

	return s.Authorizer.Authorize(ctx, request)
}

func (s *HTTPServer) Listen() (err error) {
	s.init()

	if s.listener, err = httpx.HTTPListenerForAddress(s.Addr); err != nil {
		err = merry.Prepend(err, "auxiliary server: listen")
	}

	return
}

func (s *HTTPServer) Serve() error {
	if s.listener == nil {
		return merry.New("auxiliary server: check invariants: call Listen method before Serve")
	}

	if s.UseTLS {
		return s.base.ServeTLS(s.listener, "", "")
	}

	err := s.base.Serve(s.listener)
	if merry.Is(err, http.ErrServerClosed) {
		return nil
	}

	return merry.Prepend(err, "auxiliary server: abnormal termination")
}

func (s *HTTPServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(stdctx.Background(), timeout)
	defer cancel()
	s.listener = nil

	return merry.Prepend(s.base.Shutdown(ctx), "auxiliary server: shutdown")
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ri := httpx.NewInterceptor(w)

	ctx := context.New(r.Context())

	request := httpx.GetRequest(r)
	defer httpx.PutRequest(request)

	requestID, idErr := s.generateRequestID(ctx, request)

	ctx = ctx.WithRequestID(requestID)
	ri.Header().Set(s.RequestIDHeaderName, requestID)

	var response httpx.Response
	if response = s.Authenticate(ctx, request); response != nil {
		goto finish
	}

	response = s.route(ctx, request)

finish:
	writeErr := ri.WriteResponse(response)
	snapshot := ri.Flush()

	s.invokeCompletionHookSafely(ctx, request, snapshot)

	if idErr != nil {
		s.invokeErrorHookSafely(ctx, request, idErr)
	}

	writeErr = merry.Prepend(writeErr, "auxiliary server: route: serialize response")
	if writeErr != nil {
		s.invokeErrorHookSafely(ctx, request, writeErr)
	}

	respErr := response.Err()
	if respErr != nil {
		respErr1 := merry.Prepend(respErr, "auxiliary server: route: handler failed")
		s.invokeErrorHookSafely(ctx, request, respErr1)
	}
}

func (s *HTTPServer) generateRequestID(ctx context.Context, request *httpx.Request) (string, merry.Error) {
	if s.RequestIDGenerator == nil {
		return request.ID(), nil
	}

	requestID, err, exception := s.RequestIDGenerator.InvokeSafely(ctx, request)
	if exception != nil {
		err = exception.Prepend("auxiliary server: route: generate request id")
		requestID = request.ID()
	} else if err != nil {
		err = err.Prepend("auxiliary server: route: generate request id")
		requestID = request.ID()
	} else if requestID == "" {
		err = merry.New("auxiliary server: route: generate request id: empty value")
		requestID = request.ID()
	}

	return requestID, err
}

func (s *HTTPServer) route(ctx context.Context, request *httpx.Request) (response httpx.Response) {
	if s.Router == nil {
		err := merry.New("auxiliary server: check invariants: router is nil")
		s.invokeErrorHookSafely(ctx, request, err)
		response = httpx.NewEmpty(http.StatusNotFound)
		response.Headers().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	handler, exception := s.Router.InvokeSafely(ctx, request)
	if exception != nil {
		s.invokeErrorHookSafely(ctx, request, exception)
		response = httpx.NewEmpty(http.StatusInternalServerError)
		response.Headers().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	if handler == nil {
		response = httpx.NewEmpty(http.StatusNotFound)
		response.Headers().Set("Content-Type", "text/plain; charset=utf-8")
		return
	}

	response, exception = handler.InvokeSafely(ctx, request)
	if exception != nil {
		s.invokeErrorHookSafely(ctx, request, exception)
		response = httpx.NewEmpty(http.StatusInternalServerError)
		response.Headers().Set("Content-Type", "text/plain; charset=utf-8")
	}

	return
}

func (s *HTTPServer) invokeErrorHookSafely(ctx context.Context, request *httpx.Request, err merry.Error) {
	s.ErrorHook.InvokeSafely(ctx, request, err)
}

func (s *HTTPServer) invokeCompletionHookSafely(ctx context.Context, request *httpx.Request, snapshot httpx.ResponseSnapshot) {
	if ex := s.CompletionHook.InvokeSafely(ctx, request, snapshot); ex != nil {
		exception := ex.Prepend("auxiliary server: route: run CompletionHook")
		s.invokeErrorHookSafely(ctx, request, exception)
	}
}
