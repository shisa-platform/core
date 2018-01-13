package gateway

import (
	"net/http"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestGatewayNoServices(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	err := cut.Serve([]service.Service{})
	assert.Error(t, err)
}

func TestGatewayServiceWithNoName(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	svc := &service.FakeService{
		NameHook: func() string { return "" },
	}
	err := cut.Serve([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayServiceWithNoEndpoints(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	svc := &service.FakeService{
		NameHook:      func() string { return "test" },
		EndpointsHook: func() []service.Endpoint { return nil },
	}

	err := cut.Serve([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayEndpointWithEmptyRoute(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	endpoint := service.GetEndpoint("", dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayEndpointWithRelativeRoute(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	endpoint := service.GetEndpoint("test", dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayEndpointWithNoPipelines(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	endpoint := service.Endpoint{Route: expectedRoute}
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.Serve([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayEndpointRedundantRegistration(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	endpoint1 := service.GetEndpoint(expectedRoute, dummyHandler)
	endpoint2 := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint1, endpoint2})

	err := cut.Serve([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayFieldDefaultMissingName(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9003",
	}

	pipeline := &service.Pipeline{
		Handlers:    []service.Handler{dummyHandler},
		QueryFields: []service.Field{service.Field{Default: "zalgo"}},
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
		err := cut.Serve([]service.Service{svc})
		assert.Error(t, err)
	}
}

func TestGatewayMisconfiguredTLS(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	err := cut.ServeTLS([]service.Service{svc})
	assert.Error(t, err)
}

func TestGatewayFailingAuxillary(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: ":9001",
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	aux := &auxillary.FakeServer{
		AddressHook: func() string {
			return "127.0.0.1:0"
		},
		NameHook: func() string {
			return "aux"
		},
		ServeHook: func() error {
			return merry.New("i blewed up!")
		},
		ShutdownHook: func(gracePeriod time.Duration) error {
			return nil
		},
	}

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()
	err := cut.Serve([]service.Service{svc}, aux)
	assert.Error(t, err)
}

func TestGatewayFullyLoadedEndpoint(t *testing.T) {
	cut := &Gateway{
		Name:    "test",
		Address: "127.0.0.1:0",
	}

	pipline := &service.Pipeline{Handlers: []service.Handler{dummyHandler}}
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
	err := cut.Serve([]service.Service{svc})
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
	assert.Equal(t, svc.Name(), e.serviceName)
	assert.NotNil(t, e.badQueryHandler)
	assert.NotNil(t, e.notAllowedHandler)
	assert.NotNil(t, e.redirectHandler)
	assert.NotNil(t, e.iseHandler)
}

func TestGatewayAuxillaryServer(t *testing.T) {
	expectedGracePeriod := 2 * time.Second
	gw := &Gateway{
		Name:        "test",
		Address:     "127.0.0.1:0",
		GracePeriod: expectedGracePeriod,
	}
	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	aux := &auxillary.FakeServer{
		AddressHook: func() string {
			return "127.0.0.1:0"
		},
		NameHook: func() string {
			return "fake"
		},
		ServeHook: func() error {
			return http.ErrServerClosed
		},
		ShutdownHook: func(gracePeriod time.Duration) error {
			if gracePeriod != expectedGracePeriod {
				t.Errorf("grace period %v != expected %v", gracePeriod, expectedGracePeriod)
			}
			return nil
		},
	}

	timer := time.AfterFunc(50*time.Millisecond, func() { gw.Shutdown() })
	defer timer.Stop()

	if err := gw.Serve([]service.Service{svc}, aux); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	aux.AssertServeCalledOnce(t)
	aux.AssertShutdownCalledOnceWith(t, expectedGracePeriod)
}

func teapotHandler(context.Context, *service.Request) service.Response {
	return service.NewEmpty(http.StatusTeapot)
}

func TestInstallPipelineAppliesServiceHandlers(t *testing.T) {
	pipeline := &service.Pipeline{Handlers: []service.Handler{dummyHandler}}

	augpipe, err := installPipeline([]service.Handler{teapotHandler}, pipeline)

	assert.NoError(t, err)
	assert.Len(t, augpipe.Handlers, 2)
	assert.Equal(t, augpipe.Handlers[0](nil, nil).StatusCode(), 418)
	assert.Equal(t, augpipe.Handlers[1](nil, nil).StatusCode(), 200)
}
