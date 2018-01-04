package auxillary

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/service"
)

const (
	defaultRequestIDResponseHeader = "X-Request-ID"
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

	base http.Server
}

func (s *HTTPServer) Address() string {
	return s.Addr
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

	if s.DisableKeepAlive {
		s.base.SetKeepAlivesEnabled(false)
	}

	if s.RequestIDHeaderName == "" {
		s.RequestIDHeaderName = defaultRequestIDResponseHeader
	}

	if s.RequestIDGenerator == nil {
		s.RequestIDGenerator = s.generateRequestID
	}
}

func (s *HTTPServer) generateRequestID(c context.Context, r *service.Request) (string, merry.Error) {
	return r.ID(), nil
}

// ResponseInterceptor implements `http.ResponseWriter` to capture
// and log the response sent to a user agent.
// Use it to wrap a standard  `http.ResponseWriter` and log the
// request to a logger named "request" at the `Info` level.  All
// fields are required.
type ResponseInterceptor struct {
	Logger   *zap.Logger         // The logger for requests
	Delegate http.ResponseWriter // The underlying writer
	Start    time.Time           // The response start time
	status   int
	size     int
}

func (i *ResponseInterceptor) Header() http.Header {
	return i.Delegate.Header()
}

func (i *ResponseInterceptor) Write(data []byte) (int, error) {
	size, err := i.Delegate.Write(data)
	i.size += size

	return size, err
}

func (i *ResponseInterceptor) WriteHeader(status int) {
	i.status = status
	i.Delegate.WriteHeader(status)
}

// Flush logs the requst to a logger named "request" at the
// `Info` level.
// No logging is peformed of either the logger doesn't exist or
// the `Info` level is not configured.  If the underlying writer
// implements `http.Flusher` then the `Flush` method will be
// called.
func (i *ResponseInterceptor) Flush(ctx context.Context, r *service.Request) {
	if ce := i.Logger.Check(zap.InfoLevel, "request"); ce != nil {
		end := time.Now().UTC()
		elapsed := end.Sub(i.Start)
		if i.status == 0 {
			i.status = http.StatusOK
		}
		fs := make([]zapcore.Field, 9, 10)
		fs[0] = zap.String("request-id", ctx.RequestID())
		fs[1] = zap.String("client-ip-address", r.ClientIP())
		fs[2] = zap.String("method", r.Method)
		fs[3] = zap.String("uri", r.URL.RequestURI())
		fs[4] = zap.Int("status-code", i.status)
		fs[5] = zap.Int("response-size", i.size)
		fs[6] = zap.String("user-agent", r.UserAgent())
		fs[7] = zap.Time("start", i.Start)
		fs[8] = zap.Duration("elapsed", elapsed)
		if u := ctx.Actor(); u != nil {
			fs = append(fs, zap.String("user-id", u.ID()))
		}
		ce.Write(fs...)
	}

	if f, ok := i.Delegate.(http.Flusher); ok {
		f.Flush()
	}
}
