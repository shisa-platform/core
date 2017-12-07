package gateway

import (
	stdctx "context"
	"net/http"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/service"
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

	g.init()
	defer g.Logger.Sync()

	if err := g.installServices(services); err != nil {
		return err
	}

	g.auxiliaries = auxiliaries

	defer func() {
		if err != nil {
			// xxx - add values from merry to zap here
			g.Logger.Error("fatal error serving gateway", zap.Error(err))
		}
	}()

	ach := make(chan error, len(g.auxiliaries))
	for _, aux := range g.auxiliaries {
		g.Logger.Info("starting auxillary server", zap.String("name", aux.Name()), zap.String("address", aux.Address()))
		go func(server auxillary.Server) {
			ach <- server.Serve()
		}(aux)
	}

	g.Logger.Info("starting gateway", zap.String("addr", g.Address))
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
				aerrs = append(aerrs, merry.Wrap(aerr))
			}
		case gerr := <-gch:
			err = multierr.Combine(aerrs...)
			if gerr == http.ErrServerClosed {
				g.Logger.Info("gateway service closed")
			} else if err != nil {
				err = multierr.Append(err, merry.Wrap(gerr))
			}
			return
		}
	}
}

func (g *Gateway) installServices(services []service.Service) merry.Error {
	for _, service := range services {
		if service.Name() == "" {
			return merry.New("service name cannot be empty")
		}
		if len(service.Endpoints()) == 0 {
			return merry.New("service endpoints cannot be empty").WithValue("service", service.Name())
		}

		g.Logger.Info("installing service", zap.String("name", service.Name()))
		for i, endpoint := range service.Endpoints() {
			if endpoint.Service == nil {
				return merry.New("endpoint service pointer cannot be nil").WithValue("service", service.Name()).WithValue("index", i)
			}
			if endpoint.Route == "" {
				return merry.New("endpoint route cannot be emtpy").WithValue("service", service.Name()).WithValue("index", i)
			}
			if endpoint.Route[0] != '/' {
				return merry.New("endpoint route must begin with '/'").WithValue("service", service.Name()).WithValue("index", i)
			}
			switch {
			case endpoint.Head != nil:
				endpoint.Head.Handlers = append(service.Handlers(), endpoint.Head.Handlers...)
			case endpoint.Get != nil:
				endpoint.Get.Handlers = append(service.Handlers(), endpoint.Get.Handlers...)
			case endpoint.Put != nil:
				endpoint.Put.Handlers = append(service.Handlers(), endpoint.Put.Handlers...)
			case endpoint.Post != nil:
				endpoint.Post.Handlers = append(service.Handlers(), endpoint.Post.Handlers...)
			case endpoint.Patch != nil:
				endpoint.Patch.Handlers = append(service.Handlers(), endpoint.Patch.Handlers...)
			case endpoint.Delete != nil:
				endpoint.Delete.Handlers = append(service.Handlers(), endpoint.Delete.Handlers...)
			case endpoint.Connect != nil:
				endpoint.Connect.Handlers = append(service.Handlers(), endpoint.Connect.Handlers...)
			case endpoint.Options != nil:
				endpoint.Options.Handlers = append(service.Handlers(), endpoint.Options.Handlers...)
			case endpoint.Trace != nil:
				endpoint.Trace.Handlers = append(service.Handlers(), endpoint.Trace.Handlers...)
			default:
				return merry.New("endpoint requires least one method").WithValue("service", service.Name()).WithValue("index", i)
			}

			g.Logger.Debug("adding endpoint", zap.String("route", endpoint.Route))
			if err := g.tree.addRoute(endpoint.Route, &endpoint); err != nil {
				return err
			}
		}
	}

	return nil
}
