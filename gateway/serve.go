package gateway

import (
	stdctx "context"
	"net/http"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
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
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), s.GracePeriod)
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

	s.base.Handler = s

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

func (s *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// xxx - escape path ?

	method := r.Method
	root := s.trees.get(method)
	if root == nil {
		// xxx - support custom method not allowed handler
		s.Logger.Info("no tree for method", zap.String("method", method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	// xxx - use params
	// xxx - handle tsr
	endpoint, _, _, err := root.getValue(path, false)
	if err != nil {
		// xxx - be better
		s.Logger.Error("internal error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if endpoint == nil {
		if method != http.MethodConnect && path != "/" {
			// xxx - redirect?
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	request := &service.Request{Request: r}

	requestID := ""
	if endpoint.Policy.GenerateRequestID {
		requestID = request.GenerateID()
	}

	parent := stdctx.Background()
	if endpoint.Policy.RequestBudget != 0 {
		// xxx - watch for timeout and kill pipeline, return
		var cancel stdctx.CancelFunc
		parent, cancel = stdctx.WithTimeout(parent, endpoint.Policy.RequestBudget)
		defer cancel()
	}

	// xxx - fetch context from pool
	ctx := context.New(parent)
	ctx.SetRequestID(requestID)
	response := endpoint.Pipeline[0](ctx, request)

	for k, vs := range response.Header() {
		w.Header()[k] = vs
	}
	if endpoint.Policy.GenerateRequestID {
		w.Header().Add("X-Request-ID", requestID)
	}
	// xxx - write trailers

	w.WriteHeader(response.StatusCode())
	// xxx - handle error here
	response.Serialize(w)
}

func isSupportedMethod(method string) bool {
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}

	return false
}
