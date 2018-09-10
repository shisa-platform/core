package gateway

import (
	stdctx "context"
	"expvar"
	"net/http"
	"os/signal"
	"sort"
	"strconv"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/service"
)

type byName []httpx.ParameterSchema

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (g *Gateway) Address() string {
	if g.listener != nil {
		return g.listener.Addr().String()
	}

	return g.Addr
}

func (g *Gateway) Serve(services ...*service.Service) merry.Error {
	return g.serve(services, false)
}

func (g *Gateway) ServeTLS(services ...*service.Service) merry.Error {
	return g.serve(services, true)
}

func (g *Gateway) Shutdown() (err error) {
	if !g.started {
		return
	}

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), g.GracePeriod)
	defer cancel()

	err = merry.Prepend(g.base.Shutdown(ctx), "gateway: shutdown")

	if g.HandleInterrupt && g.interrupt != nil {
		signal.Stop(g.interrupt)
		close(g.interrupt)
	}
	g.started = false

	return
}

func (g *Gateway) serve(services []*service.Service, tls bool) (err merry.Error) {
	if len(services) == 0 {
		return merry.New("gateway: check invariants: services empty")
	}
	if g.started {
		return merry.New("gateway: check invariants: already running")
	}

	g.init()

	if err := g.installServices(services); err != nil {
		return err
	}

	g.listener, err = httpx.HTTPListenerForAddress(g.Addr)
	if err != nil {
		return err.Prepend("gateway: serve")
	}

	if err1 := g.registerSafely(); err1 != nil {
		g.listener.Close()
		return err1.Prepend("gateway: serve: register gateway")
	}
	defer func() {
		if err1 := g.deregisterSafely(); err1 != nil && err == nil {
			err = err1.Prepend("gateway: serve: deregister gateway")
		}
	}()

	if err1 := g.addHealthcheckSafely(); err1 != nil {
		g.listener.Close()
		return err1.Prepend("gateway: serve: add healthcheck")
	}
	defer func() {
		if err1 := g.removeHealthcheckSafely(); err1 != nil && err == nil {
			err = err1.Prepend("gateway: serve: remove healthcheck")
		}
	}()

	g.started = true
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

func (g *Gateway) installServices(services []*service.Service) merry.Error {
	servicesExpvar := new(expvar.Map)
	gatewayExpvar.Set("services", servicesExpvar)
	for _, svc := range services {
		if svc.Name == "" {
			return merry.New("gateway: check invariants: service name empty")
		}
		if len(svc.Endpoints) == 0 {
			return merry.New("gateway: check invariants: service endpoints empty").Append(svc.Name)
		}

		serviceVar := new(expvar.Map)
		servicesExpvar.Set(svc.Name, serviceVar)

		for i, endp := range svc.Endpoints {
			if endp.Route == "" {
				return merry.New("gateway: check invariants: endpoint route emtpy").Append(svc.Name).Append(strconv.Itoa(i))
			}
			if endp.Route[0] != '/' {
				return merry.New("gateway: check invariants: endpoint route must begin with '/'").Append(svc.Name).Append(endp.Route)
			}

			e := endpoint{
				Endpoint: service.Endpoint{
					Route: endp.Route,
				},
				serviceName:       svc.Name,
				badQueryHandler:   svc.MalformedRequestHandler,
				notAllowedHandler: svc.MethodNotAllowedHandler,
				redirectHandler:   svc.RedirectHandler,
				iseHandler:        svc.InternalServerErrorHandler,
			}

			foundMethod := false
			if endp.Head != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Head)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodHead)
				}
				e.Head = pipeline
			}
			if endp.Get != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Get)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodGet)
				}
				e.Get = pipeline
			}
			if endp.Put != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Put)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodPut)
				}
				e.Put = pipeline
			}
			if endp.Post != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Post)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodPost)
				}
				e.Post = pipeline
			}
			if endp.Patch != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Patch)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodPatch)
				}
				e.Patch = pipeline
			}
			if endp.Delete != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Delete)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodDelete)
				}
				e.Delete = pipeline
			}
			if endp.Connect != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Connect)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodConnect)
				}
				e.Connect = pipeline
			}
			if endp.Options != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Options)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodOptions)
				}
				e.Options = pipeline
			}
			if endp.Trace != nil {
				foundMethod = true
				pipeline, err := installPipeline(svc.Handlers, endp.Trace)
				if err != nil {
					return err.Append(svc.Name).Append(endp.Route).Append(http.MethodTrace)
				}
				e.Trace = pipeline
			}

			if !foundMethod {
				return merry.New("gateway: check invariants: endpoint requires least one method").Append(svc.Name).Append(strconv.Itoa(i))
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
	for _, field := range pipeline.QuerySchemas {
		if field.Default != "" && field.Name == "" {
			return nil, merry.New("gateway: check invariants: field default requires name")
		}
	}

	result := &service.Pipeline{
		Policy:       pipeline.Policy,
		Handlers:     append(handlers, pipeline.Handlers...),
		QuerySchemas: append([]httpx.ParameterSchema(nil), pipeline.QuerySchemas...),
	}
	sort.Sort(byName(result.QuerySchemas))

	return result, nil
}

func (g *Gateway) registerSafely() (err merry.Error) {
	if g.Registrar == nil {
		return
	}

	if g.RegistrationURLHook == nil {
		return
	}

	defer errorx.CapturePanic(&err, "panic registering service")

	u, err, exception := g.RegistrationURLHook.InvokeSafely()
	if exception != nil {
		return merry.Prepend(exception, "run RegistrationURLHook")
	} else if err != nil {
		return merry.Prepend(err, "run RegistrationURLHook")
	}

	return g.Registrar.Register(g.Name, u)
}

func (g *Gateway) deregisterSafely() (err merry.Error) {
	if g.Registrar == nil {
		return
	}

	defer errorx.CapturePanic(&err, "panic deregistering service")

	return g.Registrar.Deregister(g.Name)
}

func (g *Gateway) addHealthcheckSafely() (err merry.Error) {
	if g.CheckURLHook == nil {
		return
	}

	defer errorx.CapturePanic(&err, "panic adding healthcheck")

	url, err, exception := g.CheckURLHook.InvokeSafely()
	if exception != nil {
		return exception.Prepend("run CheckURLHook")
	} else if err != nil {
		return err.Prepend("run CheckURLHook")
	}

	return g.Registrar.AddCheck(g.Name, url)
}

func (g *Gateway) removeHealthcheckSafely() (err merry.Error) {
	if g.CheckURLHook == nil {
		return
	}

	defer errorx.CapturePanic(&err, "panic removing healthcheck")

	return g.Registrar.RemoveChecks(g.Name)
}
