package auxillary

import (
	"context"
	"expvar"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"
)

const (
	defaultDebugServerPath = "/debug/vars"
)

type DebugServer struct {
	HTTPServer
	Path string // URL path to listen on, "/debug/vars" if empty

	// Logger optionally specifies the logger to use by the Debug
	// server.
	// If nil all logging is disabled.
	Logger *zap.Logger

	delegate http.Handler
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return merry.Wrap(s.base.Shutdown(ctx))
}

func (s *DebugServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Path == r.URL.Path {
		s.delegate.ServeHTTP(w, r)
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
