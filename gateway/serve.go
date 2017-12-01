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

var (
	supportedMethods = []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
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

	s.installServices(services)

	s.base.Handler = http.HandlerFunc(s.dispatch)

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

func (s *Gateway) installServices(services []service.Service) error {
	for _, service := range services {
		if service.Name() == "" {
			return merry.New("service name cannot be empty")
		}
		if len(service.Endpoints()) == 0 {
			return merry.New("service endpoints cannot be empty").WithValue("service", service.Name())
		}

		s.Logger.Info("installing service", zap.String("name", service.Name()))
		for _, endpoint := range service.Endpoints() {
			if endpoint.Method == "" {
				return merry.New("endpoint method cannot be emtpy").WithValue("service", service.Name())
			}
			if !isSupportedMethod(endpoint.Method) {
				return merry.New("method not supported").WithValue("service", service.Name()).WithValue("method", endpoint.Method)
			}
			if endpoint.Route == "" {
				return merry.New("endpoint route cannot be emtpy").WithValue("service", service.Name())
			}
			if endpoint.Route[0] != '/' {
				return merry.New("endpoint route must begin with '/'").WithValue("service", service.Name())
			}

			s.Logger.Debug("adding endpoint", zap.String("method", endpoint.Method), zap.String("route", endpoint.Route))
			if err := s.trees.addEndpoint(&endpoint); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Gateway) dispatch(w http.ResponseWriter, r *http.Request) {

}

func isSupportedMethod(method string) bool {
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}

	return false
}
