package gateway

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/percolate/shisa/auxillary"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestAuxillaryServer(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("unexpected logger error: %v", err)
	}
	defer logger.Sync()
	expectedGracePeriod := 2 * time.Second
	dummyEndpoint := service.Endpoint{
		Method: http.MethodGet,
		Route:  "/",
		Pipeline: []service.Handler{
			func(context.Context, *service.Request) service.Response {
				return service.NewOK(nil)
			},
		},
	}
	gw := &Gateway{
		Name:        "test",
		Address:     ":9001", // it's over 9000!
		GracePeriod: expectedGracePeriod,
		Logger:      logger,
	}
	svc := &service.FakeService{
		NameHook: func() string {
			return "fake"
		},
		EndpointsHook: func() []service.Endpoint {
			return []service.Endpoint{dummyEndpoint}
		},
	}
	aux := &auxillary.FakeServer{
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

	if err := gw.Serve([]service.Service{svc}, aux); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	aux.AssertServeCalledOnce(t)
	aux.AssertShutdownCalledOnceWith(t, expectedGracePeriod)
}
