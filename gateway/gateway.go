package gateway

import (
	"context"
	"crypto/tls"
	"errors"
	"expvar"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansel1/merry"
	"github.com/percolate/shisa/server"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var (
	stats               = expvar.NewMap("gateway")
	AlreadyStartedError = errors.New("server already started, cannot register auxillary server.")
)

type Gateway struct {
	Name             string        // The name of the Gateway for in logging
	Address          string        // TCP address to listen on, ":http" if empty
	Trace            bool          // Should trace-level logging be enabled?
	HandleInterrupt  bool          // Should SIGINT and SIGTERM interrupts be handled?
	DisableKeepAlive bool          // Should TCP keep alive be disabled?
	GracePeriod      time.Duration // Timeout for graceful shutdown of open connections
	TLSConfig        *tls.Config   // optional TLS config, used by ServeTLS and ListenAndServeTLS

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

	Logger  *zap.Logger
	base    http.Server
	aux     []server.Server
	started bool
}

func (s *Gateway) RegisterAuxillary(aux server.Server) error {
	if s.started {
		return AlreadyStartedError
	}

	s.aux = append(s.aux, aux)

	return nil
}

func (s *Gateway) Serve() (err error) {
	err = s.serve(false)
	return
}

func (s *Gateway) ServeTLS() (err error) {
	err = s.serve(true)
	return
}

func (s *Gateway) Shutdown() error {
	s.Logger.Info("shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), s.GracePeriod)
	defer cancel()

	err := merry.Wrap(s.base.Shutdown(ctx))

	for _, aux := range s.aux {
		e := multierr.Append(err, merry.Wrap(aux.Shutdown(s.GracePeriod)))
		err = merry.Wrap(e)
	}

	s.started = false
	return err
}

func (s *Gateway) init() {
	s.started = true
	stats = stats.Init()
	s.base.Addr = s.Address
	s.base.TLSConfig = s.TLSConfig
	s.base.ReadTimeout = s.ReadTimeout
	s.base.ReadHeaderTimeout = s.ReadHeaderTimeout
	s.base.WriteTimeout = s.WriteTimeout
	s.base.IdleTimeout = s.IdleTimeout
	s.base.MaxHeaderBytes = s.MaxHeaderBytes
	s.base.TLSNextProto = s.TLSNextProto
	s.base.ConnState = connstate

	if s.DisableKeepAlive {
		s.base.SetKeepAlivesEnabled(false)
	}

	if s.HandleInterrupt {
		interrupt := make(chan os.Signal, 1)
		go s.handleInterrupt(interrupt)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	}

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
}

func (s *Gateway) serve(tls bool) (err error) {
	defer func() {
		if err != nil {
			s.Logger.Fatal("fatal error serving gateway", zap.Error(err))
		}
	}()

	s.init()
	defer s.Logger.Sync()

	s.base.Handler = http.HandlerFunc(dummy)
	s.Logger.Info("starting gateway...", zap.String("addr", s.Address))

	ach := make(chan error, len(s.aux))
	gch := make(chan error, 1)
	for _, aux := range s.aux {
		go func() {
			ach <- aux.Serve()
		}()
	}

	go func() {
		if tls {
			gch <- s.base.ListenAndServeTLS("", "")
		} else {
			gch <- s.base.ListenAndServe()
		}
	}()

	aerrs := make([]error, len(s.aux))
	for {
		select {
		case aerr := <-ach:
			if aerr == http.ErrServerClosed {
				s.Logger.Info("auxillary service closed")
			} else if err != nil {
				s.Logger.Error("error in auxillary service", zap.Error(aerr))
				aerrs = append(aerrs, merry.Wrap(aerr))
			}
		case gerr := <-gch:
			err = multierr.Combine(aerrs...)
			if gerr == http.ErrServerClosed {
				s.Logger.Info("gateway service closed")
			} else if err != nil {
				s.Logger.Fatal("error in gateway service", zap.Error(gerr))
				err = multierr.Append(err, merry.Wrap(gerr))
			}
			return
		}
	}
}

func connstate(con net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		stats.Add("total_connections", 1)
		stats.Add("connected", 1)
	case http.StateClosed, http.StateHijacked:
		stats.Add("connected", -1)
	}
}

func (s *Gateway) handleInterrupt(interrupt chan os.Signal) {
	select {
	case <-interrupt:
		s.Logger.Info("interrupt received!")
		s.Shutdown()
	}
}

func dummy(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, world"))
}
