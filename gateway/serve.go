package gateway

import (
	stdctx "context"
	"expvar"
	"net/http"
	"sort"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

type byName []httpx.Field

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (g *Gateway) Address() string {
	if g.listener != nil {
		return g.listener.Addr().String()
	}

	return g.Addr
}

func (g *Gateway) Serve(services ...service.Service) merry.Error {
	return g.serve(services, false)
}

func (g *Gateway) ServeTLS(services ...service.Service) merry.Error {
	return g.serve(services, true)
}

func (g *Gateway) Shutdown() (err error) {
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), g.GracePeriod)
	defer cancel()

	err = merry.Prepend(g.base.Shutdown(ctx), "gateway: shutdown")

	g.started = false
	return
}

func (g *Gateway) serve(services []service.Service, tls bool) (err merry.Error) {
	if len(services) == 0 {
		return merry.New("gateway: check invariants: services empty")
	}

	g.init()

	if err := g.installServices(services); err != nil {
		return err
	}

	g.listener, err = httpx.HTTPListenerForAddress(g.Addr)
	if err != nil {
		return merry.Prepend(err, "gateway: serve")
	}

	if err1 := g.registerSafely(g.listener.Addr().String()); err1 != nil {
		g.listener.Close()
		return merry.Prepend(err1, "gateway: serve: register gateway")
	}
	defer func() {
		if err1 := g.deregisterSafely(); err1 != nil && err == nil {
			err = merry.Prepend(err1, "gateway: serve: deregister gateway")
		}
	}()

	if err1 := g.addHealthcheckSafely(); err1 != nil {
		g.listener.Close()
		return merry.Prepend(err1, "gateway: serve: add healthcheck")
	}
	defer func() {
		if err1 := g.removeHealthcheckSafely(); err1 != nil && err == nil {
			err = merry.Prepend(err1, "gateway: serve: remove healthcheck")
		}
	}()

	var err1 error
	if tls {
		err1 = g.base.ServeTLS(g.listener, "", "")
	} else {
		err1 = g.base.Serve(g.listener)
	}

	if merry.Is(err1, http.ErrServerClosed) {
		return nil
	}

	return merry.Prepend(err1, "gateway: serve: abnormal termination")
}

func (g *Gateway) installServices(services []service.Service) merry.Error {
	servicesExpvar := new(expvar.Map)
	gatewayExpvar.Set("services", servicesExpvar)
	for _, svc := range services {
		if svc.Name() == "" {
			return merry.New("gateway: check invariants: service name empty")
		}
		if len(svc.Endpoints()) == 0 {
			return merry.New("gateway: check invariants: service endpoints empty").WithValue("service", svc.Name())
		}

		serviceVar := new(expvar.Map)
		servicesExpvar.Set(svc.Name(), serviceVar)

		for i, endp := range svc.Endpoints() {
			if endp.Route == "" {
				return merry.New("gateway: check invariants: endpoint route emtpy").WithValue("service", svc.Name()).WithValue("index", i)
			}
			if endp.Route[0] != '/' {
				return merry.New("gateway: check invariants: endpoint route must begin with '/'").WithValue("service", svc.Name()).WithValue("route", endp.Route)
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
				return merry.New("gateway: check invariants: endpoint requires least one method").WithValue("service", svc.Name()).WithValue("index", i)
			}

			if err := g.tree.addRoute(endp.Route, &e); err != nil {
				return err
			}

			serviceVar.Set(e.Route, e)
		}
	}

	return nil
}

func installPipeline(handlers []httpx.Handler, pipeline *service.Pipeline) (*service.Pipeline, merry.Error) {
	for _, field := range pipeline.QueryFields {
		if field.Default != "" && field.Name == "" {
			return nil, merry.New("gateway: check invariants: field default requires name")
		}
	}

	result := &service.Pipeline{
		Policy:      pipeline.Policy,
		Handlers:    append(handlers, pipeline.Handlers...),
		QueryFields: append([]httpx.Field(nil), pipeline.QueryFields...),
	}
	sort.Sort(byName(result.QueryFields))

	return result, nil
}

func (g *Gateway) registerSafely(addr string) (err merry.Error) {
	if g.Registrar == nil {
		return
	}

	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic registering service")
			return
		}

		err = merry.Errorf("panic registering service: \"%v\"", arg)
	}()

	return g.Registrar.Register(g.Name, addr)
}

func (g *Gateway) deregisterSafely() (err merry.Error) {
	if g.Registrar == nil {
		return
	}

	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic deregistering service")
			return
		}

		err = merry.Errorf("panic deregistering service: \"%v\"", arg)
	}()

	return g.Registrar.Deregister(g.Name)
}

func (g *Gateway) addHealthcheckSafely() (err merry.Error) {
	if g.CheckURLHook == nil {
		return
	}

	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic adding healthcheck")
			return
		}

		err = merry.Errorf("panic adding healthcheck: \"%v\"", arg)
	}()

	url, err, exception := g.CheckURLHook.InvokeSafely()
	if exception != nil {
		return merry.Prepend(exception, "run CheckURLHook")
	} else if err != nil {
		return merry.Prepend(err, "run CheckURLHook")
	}

	return g.Registrar.AddCheck(g.Name, url)
}

func (g *Gateway) removeHealthcheckSafely() (err merry.Error) {
	if g.CheckURLHook == nil {
		return
	}

	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic removing healthcheck")
			return
		}

		err = merry.Errorf("panic removing healthcheck: \"%v\"", arg)
	}()

	return g.Registrar.RemoveChecks(g.Name)
}
