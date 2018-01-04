package auxillary

import (
	stdctx "context"
	"expvar"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

const (
	defaultDebugServerPath = "/debug/vars"
)

var (
	backgroundContext = stdctx.Background()
)

type DebugServer struct {
	HTTPServer
	Path string // URL path to listen on, "/debug/vars" if empty

	// Logger optionally specifies the logger to use by the Debug
	// server.
	// If nil all logging is disabled.
	Logger *zap.Logger

	requestLog *zap.Logger
	delegate   http.Handler
}

func (s *DebugServer) init() {
	s.HTTPServer.init()

	if s.Path == "" {
		s.Path = defaultDebugServerPath
	}

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
	defer s.Logger.Sync()
	s.requestLog = s.Logger.Named("request")

	s.delegate = expvar.Handler()
	s.base.Handler = s
}

func (s *DebugServer) Name() string {
	return "debug"
}

func (s *DebugServer) Serve() error {
	s.init()

	if s.UseTLS {
		return s.base.ListenAndServeTLS("", "")
	}

	return s.base.ListenAndServe()
}

func (s *DebugServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(backgroundContext, timeout)
	defer cancel()
	return merry.Wrap(s.base.Shutdown(ctx))
}

func (s *DebugServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ri := ResponseInterceptor{
		Logger:   s.requestLog,
		Delegate: w,
		Start:    time.Now().UTC(),
	}

	ctx := context.New(backgroundContext)
	request := &service.Request{Request: r}

	requestID, idErr := s.RequestIDGenerator(ctx, request)
	if idErr != nil {
		requestID = request.ID()
	}
	if requestID == "" {
		idErr = merry.New("empty request id").WithUserMessage("Request ID Generator returned empty string")
		requestID = request.ID()
	}

	ctx = context.WithRequestID(ctx, requestID)
	ri.Header().Set(s.RequestIDHeaderName, requestID)

	var writeErr error

	if s.Authentication != nil {
		if response := s.Authentication.Service(ctx, request); response != nil {
			writeErr = writeResponse(&ri, response)
			goto finish
		}
		if s.Authorizer != nil {
			if err := s.Authorizer.Authorize(ctx, request); err != nil {
				response := s.Authentication.UnauthorizedHandler(ctx, request)
				writeErr = writeResponse(&ri, response)
				goto finish
			}
		}
	}

	if s.Path == r.URL.Path {
		s.delegate.ServeHTTP(&ri, r)
	} else {
		ri.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ri.WriteHeader(http.StatusNotFound)
	}

finish:
	ri.Flush(ctx, request)

	if idErr != nil {
		s.Logger.Warn("request id generator failed, fell back to default", zap.String("request-id", requestID), zap.Error(idErr))
	}
	if writeErr != nil {
		s.Logger.Error("error serializing response", zap.String("request-id", requestID), zap.Error(writeErr))
	}
}

func writeResponse(w http.ResponseWriter, response service.Response) (err error) {
	for k, vs := range response.Headers() {
		w.Header()[k] = vs
	}
	for k := range response.Trailers() {
		w.Header().Add("Trailer", k)
	}

	w.WriteHeader(response.StatusCode())

	_, err = response.Serialize(w)

	for k, vs := range response.Trailers() {
		w.Header()[k] = vs
	}

	return
}
