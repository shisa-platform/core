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

	// Logger optionally specifies the logger to use by the
	// Gateway and all of its services.  Leave this as nil to
	// disable all logging.
	Logger *zap.Logger
}

func (s *DebugServer) Name() string {
	return "debug"
}

func (s *DebugServer) Serve() error {
	s.base.Addr = s.Address
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

	mux := http.NewServeMux()
	path := s.Path
	if path == "" {
		path = defaultDebugServerPath
	}

	mux.Handle(path, expvar.Handler())
	s.base.Handler = mux

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}

	s.Logger.Info("starting debug server...", zap.String("addr", s.Address))

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
