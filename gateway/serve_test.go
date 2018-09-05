package gateway

import (
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/sd"
	"github.com/shisa-platform/core/service"
)

func TestGatewayNoServices(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	err := cut.Serve()
	assert.Error(t, err)
}

func TestGatewayServiceWithNoName(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	svc := &service.Service{}
	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayServiceWithNoEndpoints(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	svc := &service.Service{Name: "test"}

	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayEndpointWithEmptyRoute(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	endpoint := service.GetEndpoint("", dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayEndpointWithRelativeRoute(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	endpoint := service.GetEndpoint("test", dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayEndpointWithNoPipelines(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	endpoint := service.Endpoint{Route: expectedRoute}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayEndpointRedundantRegistration(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	endpoint2 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1, endpoint2})

	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayFieldDefaultMissingName(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	pipeline := &service.Pipeline{
		Handlers:     []httpx.Handler{dummyHandler},
		QuerySchemas: []httpx.ParameterSchema{{Default: "zalgo"}},
	}
	endpoints := []service.Endpoint{
		{Route: "/", Head: pipeline},
		{Route: "/", Get: pipeline},
		{Route: "/", Put: pipeline},
		{Route: "/", Post: pipeline},
		{Route: "/", Patch: pipeline},
		{Route: "/", Delete: pipeline},
		{Route: "/", Connect: pipeline},
		{Route: "/", Options: pipeline},
		{Route: "/", Trace: pipeline},
	}

	for _, endpoint := range endpoints {
		svc := newFakeService([]service.Endpoint{endpoint})
		err := cut.Serve(svc)
		assert.Error(t, err)
	}
}

func TestGatewayMisconfiguredTLS(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":0",
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.ServeTLS(svc)
	assert.Error(t, err)
}

func TestGatewayListenerAddressFailure(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: ":80",
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)
	assert.Error(t, err)
}

func TestGatewayFullyLoadedEndpoint(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: "127.0.0.1:0",
	}

	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route:   expectedRoute,
		Head:    pipline,
		Get:     pipline,
		Put:     pipline,
		Post:    pipline,
		Patch:   pipline,
		Delete:  pipline,
		Connect: pipline,
		Options: pipline,
		Trace:   pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()
	err := cut.Serve(svc)
	assert.NoError(t, err)

	e, _, _, err := cut.tree.getValue(expectedRoute)
	assert.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, expectedRoute, e.Route)
	assert.NotNil(t, e.Head)
	assert.NotNil(t, e.Get)
	assert.NotNil(t, e.Put)
	assert.NotNil(t, e.Post)
	assert.NotNil(t, e.Patch)
	assert.NotNil(t, e.Delete)
	assert.NotNil(t, e.Connect)
	assert.NotNil(t, e.Options)
	assert.NotNil(t, e.Trace)
	assert.Equal(t, svc.Name, e.serviceName)
}

func teapotHandler(context.Context, *httpx.Request) httpx.Response {
	return httpx.NewEmpty(http.StatusTeapot)
}

func TestInstallPipelineAppliesServiceHandlers(t *testing.T) {
	pipeline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}

	augpipe, err := installPipeline([]httpx.Handler{teapotHandler}, pipeline)

	assert.NoError(t, err)
	assert.Len(t, augpipe.Handlers, 2)
	assert.Equal(t, augpipe.Handlers[0](nil, nil).StatusCode(), 418)
	assert.Equal(t, augpipe.Handlers[1](nil, nil).StatusCode(), 200)
}

func TestGatewayServeWithRegistrar(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.NoError(t, err)
	registrar.AssertRegisterCalledOnce(t)
}

func TestGatewayServeWithRegistrarError(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return merry.New("error")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
}

func TestGatewayServeWithRegistrarPanic(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			panic(merry.New("i blewed up!"))
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
}

func TestGatewayServeWithRegistrarPanicString(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			panic("i blewed up!")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
}

func TestGatewayServeWithNilRegistrationHook(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:      "test",
		Addr:      ":0",
		Registrar: registrar,
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.NoError(t, err)
	registrar.AssertRegisterNotCalled(t)
}

func TestGatewayServeWithRegistrationHookPanic(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:      "test",
		Addr:      ":0",
		Registrar: registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) {
			panic("i blew up!")
		},
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterNotCalled(t)
}

func TestGatewayServeWithRegistrationHookError(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:      "test",
		Addr:      ":0",
		Registrar: registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) {
			return nil, merry.New("new error")
		},
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterNotCalled(t)
}

func TestGatewayServeWithDeregisterError(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return merry.New("new error")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertDeregisterCalledOnce(t)
}

func TestGatewayServeWithDeregisterPanic(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			panic(merry.New("i blew up!"))
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertDeregisterCalledOnce(t)
}

func TestGatewayServeWithDeregisterPanicString(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			panic("i blew up!")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
	}
	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1})

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertDeregisterCalledOnce(t)
}

func TestGatewayServeWithCheckURLHookParseError(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			return nil, merry.New("new error")
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckNotCalled(t)
}

func TestGatewayRedundantStart(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: "127.0.0.1:0",
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	timer := time.AfterFunc(250*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		assert.NoError(t, cut.Serve(svc))
	}()

	time.Sleep(time.Millisecond * 200)

	wg.Wait()
	assert.Error(t, cut.Serve(svc))
}

func TestGatewayRedundantStop(t *testing.T) {
	cut := &Gateway{
		Name: "test",
		Addr: "127.0.0.1:0",
	}

	assert.NoError(t, cut.Shutdown())
}

func TestGatewayServeWithCheckURLHookPanic(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			panic(merry.New("i blew up!"))
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckNotCalled(t)
}

func TestGatewayServeWithCheckURLHookPanicString(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			panic("i blew up!")
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckNotCalled(t)
}

func TestGatewayServeWithCheckURLHookAddCheckError(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
		AddCheckHook: func(string, *url.URL) merry.Error {
			return merry.New("new error")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			u, err := url.Parse("http://localhost:5000")
			return u, merry.Wrap(err)
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckCalledOnce(t)
}

func TestGatewayServeWithCheckURLHookAddCheckPanic(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
		AddCheckHook: func(string, *url.URL) merry.Error {
			panic(merry.New("i blewed up!"))
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			u, err := url.Parse("http://localhost:5000")
			return u, merry.Wrap(err)
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckCalledOnce(t)
}

func TestGatewayServeWithCheckURLHookAddCheckPanicString(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
		AddCheckHook: func(string, *url.URL) merry.Error {
			panic("i blewed up!")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			u, err := url.Parse("http://localhost:5000")
			return u, merry.Wrap(err)
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckCalledOnce(t)
}

func TestGatewayServeWithCheckURLHookRemoveChecksError(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
		AddCheckHook: func(string, *url.URL) merry.Error {
			return nil
		},
		RemoveChecksHook: func(string) merry.Error {
			return merry.New("new error")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			u, err := url.Parse("http://localhost:5000")
			return u, merry.Wrap(err)
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})
	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckCalledOnce(t)
	registrar.AssertDeregisterCalledOnce(t)
	registrar.AssertRemoveChecksCalledOnce(t)
}

func TestGatewayServeWithCheckURLHookRemoveChecksPanic(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
		AddCheckHook: func(string, *url.URL) merry.Error {
			return nil
		},
		RemoveChecksHook: func(string) merry.Error {
			panic(merry.New("i blewed up"))
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			u, err := url.Parse("http://localhost:5000")
			return u, merry.Wrap(err)
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})
	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckCalledOnce(t)
	registrar.AssertDeregisterCalledOnce(t)
	registrar.AssertRemoveChecksCalledOnce(t)
}

func TestGatewayServeWithCheckURLHookRemoveChecksPanicString(t *testing.T) {
	registrar := &sd.FakeRegistrar{
		RegisterHook: func(string, *url.URL) merry.Error {
			return nil
		},
		DeregisterHook: func(string) merry.Error {
			return nil
		},
		AddCheckHook: func(string, *url.URL) merry.Error {
			return nil
		},
		RemoveChecksHook: func(string) merry.Error {
			panic("i blewed up")
		},
	}
	cut := &Gateway{
		Name:                "test",
		Addr:                ":0",
		Registrar:           registrar,
		RegistrationURLHook: func() (u *url.URL, err merry.Error) { return },
		CheckURLHook: func() (*url.URL, merry.Error) {
			u, err := url.Parse("http://localhost:5000")
			return u, merry.Wrap(err)
		},
	}
	pipline := &service.Pipeline{Handlers: []httpx.Handler{dummyHandler}}
	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  pipline,
	}
	svc := newFakeService([]service.Endpoint{endpoint})
	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)

	assert.Error(t, err)
	registrar.AssertRegisterCalledOnce(t)
	registrar.AssertAddCheckCalledOnce(t)
	registrar.AssertDeregisterCalledOnce(t)
	registrar.AssertRemoveChecksCalledOnce(t)
}
