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

type endpoint struct {
	service.Endpoint
	serviceName       string
	badQueryHandler service.Handler
	notAllowedHandler service.Handler
	redirectHandler   service.Handler
	iseHandler        service.ErrorHandler
}

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
	for _, svc := range services {
		if svc.Name() == "" {
			return merry.New("service name cannot be empty")
		}
		if len(svc.Endpoints()) == 0 {
			return merry.New("service endpoints cannot be empty").WithValue("service", svc.Name())
		}

		g.Logger.Info("installing service", zap.String("name", svc.Name()))
		for i, endp := range svc.Endpoints() {
			if endp.Route == "" {
				return merry.New("endpoint route cannot be emtpy").WithValue("service", svc.Name()).WithValue("index", i)
			}
			if endp.Route[0] != '/' {
				return merry.New("endpoint route must begin with '/'").WithValue("service", svc.Name()).WithValue("index", i)
			}

			e := endpoint{
				Endpoint: service.Endpoint{
					Route: endp.Route,
				},
				serviceName:       svc.Name(),
				badQueryHandler: svc.MalformedRequestHandler(),
				notAllowedHandler: svc.MethodNotAllowedHandler(),
				redirectHandler:   svc.RedirectHandler(),
				iseHandler:        svc.InternalServerErrorHandler(),
			}

			if e.badQueryHandler == nil {
				e.badQueryHandler = defaultMalformedRequestHandler
			}
			if e.notAllowedHandler == nil {
				e.notAllowedHandler = defaultMethodNotAlowedHandler
			}
			if e.redirectHandler == nil {
				e.redirectHandler = defaultRedirectHandler
			}
			if e.iseHandler == nil {
				e.iseHandler = defaultInternalServerErrorHandler
			}

			foundMethod := false
			if endp.Head != nil {
				foundMethod = true
				e.Head = &service.Pipeline{
					Policy:   endp.Head.Policy,
					Handlers: append(svc.Handlers(), endp.Head.Handlers...),
				}
			}
			if endp.Get != nil {
				foundMethod = true
				e.Get = &service.Pipeline{
					Policy:   endp.Get.Policy,
					Handlers: append(svc.Handlers(), endp.Get.Handlers...),
				}
			}
			if endp.Put != nil {
				foundMethod = true
				e.Put = &service.Pipeline{
					Policy:   endp.Put.Policy,
					Handlers: append(svc.Handlers(), endp.Put.Handlers...),
				}
			}
			if endp.Post != nil {
				foundMethod = true
				e.Post = &service.Pipeline{
					Policy:   endp.Post.Policy,
					Handlers: append(svc.Handlers(), endp.Post.Handlers...),
				}
			}
			if endp.Patch != nil {
				foundMethod = true
				e.Patch = &service.Pipeline{
					Policy:   endp.Patch.Policy,
					Handlers: append(svc.Handlers(), endp.Patch.Handlers...),
				}
			}
			if endp.Delete != nil {
				foundMethod = true
				e.Delete = &service.Pipeline{
					Policy:   endp.Delete.Policy,
					Handlers: append(svc.Handlers(), endp.Delete.Handlers...),
				}
			}
			if endp.Connect != nil {
				foundMethod = true
				e.Connect = &service.Pipeline{
					Policy:   endp.Connect.Policy,
					Handlers: append(svc.Handlers(), endp.Connect.Handlers...),
				}
			}
			if endp.Options != nil {
				foundMethod = true
				e.Options = &service.Pipeline{
					Policy:   endp.Options.Policy,
					Handlers: append(svc.Handlers(), endp.Options.Handlers...),
				}
			}
			if endp.Trace != nil {
				foundMethod = true
				e.Trace = &service.Pipeline{
					Policy:   endp.Trace.Policy,
					Handlers: append(svc.Handlers(), endp.Trace.Handlers...),
				}
			}

			if !foundMethod {
				return merry.New("endpoint requires least one method").WithValue("service", svc.Name()).WithValue("index", i)
			}

			g.Logger.Debug("adding endpoint", zap.String("route", endp.Route))
			if err := g.tree.addRoute(endp.Route, &e); err != nil {
				return err
			}
		}
	}

	return nil
}
