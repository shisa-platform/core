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

func (s *DebugServer) Name() string {
	return "debug"
}

func (s *DebugServer) Serve() error {
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

	if s.Path == "" {
		s.Path = defaultDebugServerPath
	}

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
	defer s.Logger.Sync()

	s.delegate = expvar.Handler()
	s.base.Handler = s

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
		goto flush
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

flush:
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
