package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	//"github.com/percolate/shisa/router"
	"github.com/percolate/shisa/server"
	"github.com/percolate/shisa/service"
)

func (s *Gateway) Serve(services []service.Service, auxillary ...server.Server) error {
	return s.serve(false, services, auxillary)
}

func (s *Gateway) ServeTLS(services []service.Service, auxillary ...server.Server) error {
	return s.serve(true, services, auxillary)
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

func (s *Gateway) serve(tls bool, services []service.Service, auxillary []server.Server) (err error) {
	s.init()
	if len(services) == 0 {
		return errors.New("services must not be empty")
	}

	s.aux = auxillary

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

	//s.base.Handler = http.HandlerFunc(dummy)

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
