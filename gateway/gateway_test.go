package gateway

import (
	"net/http"
	"testing"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
	"github.com/percolate/shisa/server"
)

func TestAuxillaryServer(t *testing.T) {
	expectedGracePeriod := 2 * time.Second
	dummyEndpoint := service.Endpoint{
		Method: http.MethodGet,
		Route: "/",
		Handler: &service.FakeHandler{
			HandleHook: func(context.Context, *service.Request) service.Response {
				return service.NewOK(nil)
			},
		},
	}
	gw := &Gateway{
		Name:        "test",
		Address:     ":9001", // it's over 9000!
		GracePeriod: expectedGracePeriod,
	}
	svc := &service.FakeService{
		NameHook: func() string {
			return "fake"
		},
		EndpointsHook: func() []service.Endpoint {
			return []service.Endpoint{dummyEndpoint}
		},
	}
	aux := &server.FakeServer{
		ServeHook: func() error {
			return nil
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
	err := gw.Serve([]service.Service{svc}, aux)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	aux.AssertServeCalledOnce(t)
	aux.AssertShutdownCalledOnceWith(t, expectedGracePeriod)
}
