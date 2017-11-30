package gateway

import (
	"testing"
	"time"

	"github.com/percolate/shisa/server"
	"go.uber.org/zap"
)

func TestAuxillaryServer(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("unexpected logger error: %v", err)
	}
	defer logger.Sync()
	expectedGracePeriod := 2 * time.Second
	gw := &Gateway{
		Name:        "test",
		Address:     ":9001", // it's over 9000!
		GracePeriod: expectedGracePeriod,
		Logger:      logger,
	}
	fake := &server.FakeServer{
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
	gw.RegisterAuxillary(fake)

	timer := time.AfterFunc(50*time.Millisecond, func() { gw.Shutdown() })
	defer timer.Stop()
	err = gw.Serve()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fake.AssertServeCalledOnce(t)
	fake.AssertShutdownCalledOnceWith(t, expectedGracePeriod)
}
