package gateway

import (
	stdctx "context"
	"net/http"
	"sort"

	"github.com/ansel1/merry"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/service"
)

type fields []service.Field

func (p fields) Len() int           { return len(p) }
func (p fields) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p fields) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type endpoint struct {
	service.Endpoint
	serviceName       string
	badQueryHandler   service.Handler
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
	g.Logger.Info("shutting down gateway")
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), g.GracePeriod)
	defer cancel()

	err = merry.Wrap(g.base.Shutdown(ctx))

	for _, aux := range g.auxiliaries {
		g.Logger.Info("shutting down auxillary", zap.String("name", aux.Name()))
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

	for i := len(g.auxiliaries) + 1; i != 0; i-- {
		select {
		case aerr := <-ach:
			if !merry.Is(aerr, http.ErrServerClosed) {
				err = multierr.Append(err, merry.Wrap(aerr))
			}
		case gerr := <-gch:
			if gerr != http.ErrServerClosed {
				err = multierr.Append(err, merry.Wrap(gerr))
			}
		}
	}

	return
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
				return merry.New("endpoint route must begin with '/'").WithValue("service", svc.Name()).WithValue("route", endp.Route)
			}

			e := endpoint{
				Endpoint: service.Endpoint{
					Route: endp.Route,
				},
				serviceName:       svc.Name(),
				badQueryHandler:   svc.MalformedRequestHandler(),
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
				pipeline, err := installPipeline(svc.Handlers(), endp.Head)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodHead)
				}
				e.Head = pipeline
			}
			if endp.Get != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Get)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodGet)
				}
				e.Get = pipeline
			}
			if endp.Put != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Put)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodPut)
				}
				e.Put = pipeline
			}
			if endp.Post != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Post)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodPost)
				}
				e.Post = pipeline
			}
			if endp.Patch != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Patch)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodPatch)
				}
				e.Patch = pipeline
			}
			if endp.Delete != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Delete)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodDelete)
				}
				e.Delete = pipeline
			}
			if endp.Connect != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Connect)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodConnect)
				}
				e.Connect = pipeline
			}
			if endp.Options != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Options)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodOptions)
				}
				e.Options = pipeline
			}
			if endp.Trace != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers(), endp.Trace)
				if err != nil {
					return err.WithValue("service", svc.Name()).WithValue("route", endp.Route).WithValue("method", http.MethodTrace)
				}
				e.Trace = pipeline
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

func installPipeline(handlers []service.Handler, pipeline *service.Pipeline) (*service.Pipeline, merry.Error) {
	for _, field := range pipeline.QueryFields {
		if field.Default != "" && field.Name == "" {
			return nil, merry.New("Field default requires name")
		}
	}

	result := &service.Pipeline{
		Policy:      pipeline.Policy,
		Handlers:    append(handlers, pipeline.Handlers...),
		QueryFields: append([]service.Field(nil), pipeline.QueryFields...),
	}
	sort.Sort(fields(result.QueryFields))

	return result, nil
}
