package auxiliary

import (
	stdctx "context"
	"expvar"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

const (
	defaultDebugServerPath = "/debug/vars"
)

var (
	debugStats = new(expvar.Map)
)

// DebugServer serves values from the `expvar` package to the
// configured address and path.
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
	now := time.Now().UTC().Format(startTimeFormat)
	s.HTTPServer.init()
	debugStats = debugStats.Init()
	AuxiliaryStats.Set("debug", debugStats)
	debugStats.Set("hits", new(expvar.Int))
	startTime := new(expvar.String)
	startTime.Set(now)
	debugStats.Set("starttime", startTime)

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

	listener, err := httpx.HTTPListenerForAddress(s.Addr)
	if err != nil {
		return err
	}

	addr := listener.Addr().String()
	addrVar := new(expvar.String)
	addrVar.Set(addr)
	debugStats.Set("addr", addrVar)
	s.Logger.Info("debug service started", zap.String("addr", addr))

	if s.UseTLS {
		return s.base.ServeTLS(listener, "", "")
	}

	return s.base.Serve(listener)
}

func (s *DebugServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(stdctx.Background(), timeout)
	defer cancel()
	return merry.Wrap(s.base.Shutdown(ctx))
}

func (s *DebugServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ri := httpx.NewInterceptor(w)

	debugStats.Add("hits", 1)

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

	var writeErr error
	if response := s.Authenticate(ctx, request); response != nil {
		writeErr = ri.WriteResponse(response)
		goto finish
	}

	if s.Path == r.URL.Path {
		s.delegate.ServeHTTP(ri, r)
	} else {
		ri.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ri.WriteHeader(http.StatusNotFound)
	}

finish:
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
}
