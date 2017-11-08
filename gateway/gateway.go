package gateway

import (
	"context"
	"crypto/tls"
	"expvar"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	stats := expvar.NewMap("gateway")
)

type Gateway struct {
	Name             string        // The name of the Gateway for in logging
	Addr             string        // TCP address to listen on, ":http" if empty
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
	TLSNextProto map[string]func(*Server, *tls.Conn, Handler)

	// xxx - logger, factory?
	base http.Server
}

func (s *Gateway) Serve() error {
	s.init()

	// xxx - log("starting gateway on %s", s.addrOrDefault())
	err := base.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Gateway) ServeTLS() error {
	s.init()

	// xxx - log("starting gateway on %s", s.addrOrDefault())
	err := base.ListenAndServeTLS("", "")
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Gateway) init() {
	base.Addr = s.Addr
	base.TLSConfig = s.TLSConfig
	base.ReadTimeout = s.ReadTimeout
	base.ReadHeaderTimeout = s.ReadHeaderTimeout
	base.WriteTimeout = s.WriteTimeout
	base.IdleTimeout = s.IdleTimeout
	base.MaxHeaderBytes = s.MaxHeaderBytes
	base.TLSNextProto = s.TLSNextProto
	base.ConnState = connstate

	if s.DisableKeepAlive {
		base.SetKeepAlivesEnabled(false)
	}

	if s.HandleInterrupt {
		interrupt := make(chan os.Signal, 1)
		go s.handleInterrupt(interrupt)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	}
}

func (s *Gateway) Shutdown() error {
	// xxx - log("shutting down service...")
	ctx, cancel := context.WithTimeout(context.Background(), s.GracePeriod)
	defer cancel()
	return base.Shutdown(ctx)
}

func (s *Gateway) openDependencies() error {
	// xxx - do something
	// xxx - configure pingables for healthcheck, HealthcheckProvider? RegisterHealthcheck(name, Pinger)
	// xxx - PingerAdapter struct {func() error} has Ping() method?
	// Gateway{Name(), Ping()} -> RegisterHealthcheck(s.Name(), s.Ping()) ?
	// xxx - Pinger{Ping()} are DownstreamProviders{Ping(), Close()}, harvested from pipeline participants

	return nil
}

func connstate(con net.Conn, state ConnState) {
	switch state {
	case http.StateNew:
		stats.Add("total_connections", 1)
		stats.Add("connected", 1)
	case http.StateClosed, http.StateHijacked:
		stats.Add("connected", -1)
	}	
}

func (s *Service) handleInterrupt(interrupt chan os.Signal) {
	select {
	case <-interrupt:
		// xxx - log("interrupt received!")
		s.Shutdown()
	}
}
