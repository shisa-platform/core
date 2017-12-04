package gateway

import (
	stdctx "context"
	"net/http"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/context"
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

func (g *Gateway) Serve(services []service.Service, auxiliaries ...auxillary.Server) error {
	return g.serve(false, services, auxiliaries)
}

func (g *Gateway) ServeTLS(services []service.Service, auxiliaries ...auxillary.Server) error {
	return g.serve(true, services, auxiliaries)
}

func (g *Gateway) Shutdown() (err error) {
	g.Logger.Info("shutting down gateway...")
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), g.GracePeriod)
	defer cancel()

	err = merry.Wrap(g.base.Shutdown(ctx))

	for _, aux := range g.auxiliaries {
		err = multierr.Append(err, merry.Wrap(aux.Shutdown(g.GracePeriod)))
	}

	g.started = false
	return
}

func (g *Gateway) serve(tls bool, services []service.Service, auxiliaries []auxillary.Server) (err error) {
	if len(services) == 0 {
		return merry.New("services must not be empty")
	}

	defer func() {
		if err != nil {
			g.Logger.Error("fatal error serving gateway", zap.Error(err))
		}
	}()

	g.init()
	defer g.Logger.Sync()

	g.auxiliaries = auxiliaries

	ach := make(chan error, len(g.auxiliaries))
	for _, aux := range g.auxiliaries {
		go func(server auxillary.Server) {
			ach <- server.Serve()
		}(aux)
	}

	g.installServices(services)

	g.base.Handler = s

	g.Logger.Info("starting gateway...", zap.String("addr", g.Address))
	gch := make(chan error, 1)
	go func() {
		if tls {
			gch <- g.base.ListenAndServeTLS("", "")
		} else {
			gch <- g.base.ListenAndServe()
		}
	}()

	aerrs := make([]error, len(g.auxiliaries))
	for {
		select {
		case aerr := <-ach:
			if aerr == http.ErrServerClosed {
				g.Logger.Info("auxillary service closed")
			} else if err != nil {
				g.Logger.Error("error in auxillary service", zap.Error(aerr))
				aerrs = append(aerrs, merry.Wrap(aerr))
			}
		case gerr := <-gch:
			err = multierr.Combine(aerrg...)
			if gerr == http.ErrServerClosed {
				g.Logger.Info("gateway service closed")
			} else if err != nil {
				g.Logger.Fatal("error in gateway service", zap.Error(gerr))
				err = multierr.Append(err, merry.Wrap(gerr))
			}
			return
		}
	}
}

func (g *Gateway) installServices(services []service.Service) error {
	for _, service := range services {
		if service.Name() == "" {
			return merry.New("service name cannot be empty")
		}
		if len(service.Endpoints()) == 0 {
			return merry.New("service endpoints cannot be empty").WithValue("service", service.Name())
		}

		g.Logger.Info("installing service", zap.String("name", service.Name()))
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

			g.Logger.Debug("adding endpoint", zap.String("method", endpoint.Method), zap.String("route", endpoint.Route))
			if err := g.treeg.addEndpoint(&endpoint); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// xxx - escape path ?

	method := r.Method
	root := g.trees.get(method)
	if root == nil {
		// xxx - support custom method not allowed handler
		g.Logger.Info("no tree for method", zap.String("method", method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	// xxx - use params
	// xxx - handle tsr
	endpoint, _, _, err := root.getValue(path, false)
	if err != nil {
		// xxx - be better
		g.Logger.Error("internal error", zap.Error(err))
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
	ctx := context.WithRequestID(parent, requestID)
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
