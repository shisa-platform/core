package gateway

import (
	"expvar"
	"fmt"
	"net/http"

	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/x/log"
)

// xxx - need configuration object with setter helpers
// gateway.New(name, configuration)

type Service struct {
	Name         string
	port         int
	debugPort    int
	server       *http.Server
	logger       logx.Logger
	healthchecks map[string]func() error
}

func New(port, debugPort int, trace bool) *Service {
	return &Service{
		port:      port,
		debugPort: debugPort,
	}
}

func (s *Service) Run() error {
	if err := s.openLogging(); err != nil {
		return err
	}
	if err := s.openDependencies(); err != nil {
		return err
	}

	return s.serve()
}

func (s *Service) openDependencies() error {
	// xxx - do something
	// xxx - configure pingables for healthcheck, HealthcheckProvider? RegisterHealthcheck(name, Pinger)
	// xxx - PingerAdapter struct {func() error} has Ping() method?
	// Service{Name(), Ping()} -> RegisterHealthcheck(s.Name(), s.Ping()) ?
	// xxx - Pinger{Ping()} are DownstreamProviders{Ping(), Close()}, harvested from pipeline participants

	return nil
}

func (s *Service) openLogging() error {
	logger, err := logx.New(serviceName)
	if err != nil {
		return err
	}
	s.logger = logger

	return nil
}

func (s *Service) close() {
}
