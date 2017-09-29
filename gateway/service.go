package gateway

import (
	log "github.com/percolate/shisa/log"
)

// xxx - need configuration object with setter helpers
// gateway.New(name, configuration)

type Service struct {
	Name   string
	logger log.Logger
}

func New(name string) *Service {
	return &Service{
		Name: name,
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
	logger, err := log.New(s.Name)
	if err != nil {
		return err
	}
	s.logger = logger

	return nil
}

func (s *Service) serve() error {
	return nil
}

func (s *Service) close() error {
	return nil
}
