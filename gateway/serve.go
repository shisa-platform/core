package gateway

import (
	"context"
	"net/http"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/percolate/shisa/server"
	"github.com/percolate/shisa/service"
)

func (s *Gateway) Serve(services []service.Service, auxiliaries ...server.Server) error {
	return s.serve(false, services, auxiliaries)
}

func (s *Gateway) ServeTLS(services []service.Service, auxiliaries ...server.Server) error {
	return s.serve(true, services, auxiliaries)
}

func (s *Gateway) Shutdown() (err error) {
	s.Logger.Info("shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), s.GracePeriod)
	defer cancel()

	err = merry.Wrap(s.base.Shutdown(ctx))

	for _, aux := range s.auxiliaries {
		err = multierr.Append(err, merry.Wrap(aux.Shutdown(s.GracePeriod)))
	}

	s.started = false
	return
}

func (s *Gateway) serve(tls bool, services []service.Service, auxiliaries []server.Server) (err error) {
	if len(services) == 0 {
		return merry.New("services must not be empty")
	}

	defer func() {
		if err != nil {
			s.Logger.Error("fatal error serving gateway", zap.Error(err))
		}
	}()

	s.init()
	defer s.Logger.Sync()

	s.auxiliaries = auxiliaries

	ach := make(chan error, len(s.auxiliaries))
	for _, aux := range s.auxiliaries {
		go func(server server.Server) {
			ach <- server.Serve()
		}(aux)
	}

	s.Logger.Info("starting gateway...", zap.String("addr", s.Address))
	gch := make(chan error, 1)
	go func() {
		if tls {
			gch <- s.base.ListenAndServeTLS("", "")
		} else {
			gch <- s.base.ListenAndServe()
		}
	}()

	aerrs := make([]error, len(s.auxiliaries))
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
