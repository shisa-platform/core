package gateway

import (
	"context"
	"crypto/tls"
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/percolate/shisa/server"
)

var (
	stats               = expvar.NewMap("gateway")
	AlreadyStartedError = errors.New("Server already started, cannot register auxillary server.")
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

	// xxx - logger, factory?
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

func (s *Gateway) Serve() error {
	return s.serve(false)
}

func (s *Gateway) ServeTLS() error {
	return s.serve(true)
}

func (s *Gateway) Shutdown() error {
	// xxx - log("shutting down service...")
	ctx, cancel := context.WithTimeout(context.Background(), s.GracePeriod)
	defer cancel()
	err := s.base.Shutdown(ctx)

	errs := make([]error, len(s.aux))
	for i, aux := range s.aux {
		errs[i] = aux.Shutdown(s.GracePeriod)
	}
	// xxx - gather all shutdown errors into a compound error

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
}

func (s *Gateway) serve(tls bool) (err error) {
	s.init()

	errs := make([]error, len(s.aux))
	for i, aux := range s.aux {
		y, a := i, aux
		go func() {
			errs[y] = a.Serve()
		}()
	}

	for i, auxErr := range errs {
		// xxx - make this a compound error of all sub-errors
		if auxErr != nil {
			err = fmt.Errorf("service %q failed to start: %v", s.aux[i].Name(), auxErr)
			return
		}
	}

	s.base.Handler = http.HandlerFunc(dummy)

	// xxx - log("starting gateway on %s", s.addrOrDefault())
	if tls {
		err = s.base.ListenAndServeTLS("", "")
	} else {
		err = s.base.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		err = nil
	}

	return
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
		// xxx - log("interrupt received!")
		s.Shutdown()
	}
}

func dummy(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, world"))
}
