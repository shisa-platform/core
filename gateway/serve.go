package service

import (
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	connStats     = expvar.NewMap("connstats")
	connected     = new(expvar.Int)
	gracePeriod   = time.Second * 2
	headerTimeout = time.Millisecond * 500
)

func init() {
	connStats.Set("connected", connected)
}

func (s *Service) serve() error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	go s.handleInterrupt(interrupt)

	mux := http.NewServeMux()
	mux.Handle("/debug/vars", expvar.Handler())
	mux.HandleFunc("/healthcheck", s.healthcheckHandler)

	s.logger.Stdout().Printf("starting debug on port: %v", s.debugPort)
	go http.ListenAndServe(fmt.Sprintf(":%d", s.debugPort), mux)

	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.provision(),
		ReadHeaderTimeout: headerTimeout,
		ErrorLog:          s.logger.Stderr(),
		ConnState:         trackConnections,
	}

	s.logger.Stdout().Printf("starting service on port: %v", s.port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Service) handleInterrupt(interrupt chan os.Signal) {
	select {
	case <-interrupt:
		s.server.ConnState = nil // stop listening for connection state changes
		s.shutdown(context.Background())
	}
}

func (s *Service) shutdown(c context.Context) {
	s.logger.Stdout().Println("shutting down service...")
	s.logger.Stdout().Printf("starting %v connection grace period...", gracePeriod)
	timeoutCtx, cancel := context.WithTimeout(c, gracePeriod)
	defer cancel()
	if err := s.server.Shutdown(timeoutCtx); err != nil {
		s.logger.Stdout().Println("killing open connections.")
		if err := s.server.Close(); err != nil {
			s.logger.Stderr().Printf("Error shutting down: %s", err)
		}
	}
	s.close()
}

func trackConnections(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		connected.Add(1)
	case http.StateClosed:
		connected.Add(-1)
	}
}
