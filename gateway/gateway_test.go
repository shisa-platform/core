package gateway

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/service"
)

func waitSig(t *testing.T, c <-chan os.Signal, sig os.Signal) {
	t.Helper()
	select {
	case s := <-c:
		if s != sig {
			t.Fatalf("signal was %v, want %v", s, sig)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for %v", sig)
	}
}

func TestGatewayInit(t *testing.T) {
	config := tls.Config{}
	nextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	cut := &Gateway{
		Addr:              ":9001",
		DisableKeepAlive:  true,
		TLSConfig:         &config,
		ReadTimeout:       time.Millisecond * 5,
		ReadHeaderTimeout: time.Millisecond * 10,
		WriteTimeout:      time.Millisecond * 15,
		IdleTimeout:       time.Millisecond * 20,
		MaxHeaderBytes:    1024,
		TLSNextProto:      nextProto,
	}
	cut.init()

	assert.Equal(t, cut.Addr, cut.base.Addr)
	assert.Equal(t, cut.TLSConfig, cut.base.TLSConfig)
	assert.Equal(t, cut.ReadTimeout, cut.base.ReadTimeout)
	assert.Equal(t, cut.ReadHeaderTimeout, cut.base.ReadHeaderTimeout)
	assert.Equal(t, cut.WriteTimeout, cut.base.WriteTimeout)
	assert.Equal(t, cut.IdleTimeout, cut.base.IdleTimeout)
	assert.Equal(t, cut.MaxHeaderBytes, cut.base.MaxHeaderBytes)
	assert.Equal(t, cut.TLSNextProto, cut.base.TLSNextProto)
	assert.NotNil(t, cut.base.ConnState)
	assert.Equal(t, cut, cut.base.Handler)
	assert.Equal(t, defaultRequestIDResponseHeader, cut.RequestIDHeaderName)
	assert.NotNil(t, cut.Logger)
}

func TestGatewaySignal(t *testing.T) {
	cut := &Gateway{
		Addr:            ":0",
		HandleInterrupt: true,
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := cut.Serve([]service.Service{svc})
		assert.NoError(t, err)
		wg.Done()
	}()

	time.Sleep(time.Millisecond * 200)

	monitor := make(chan os.Signal, 1)
	signal.Notify(monitor, syscall.SIGINT)
	defer signal.Stop(monitor)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	waitSig(t, monitor, syscall.SIGINT)

	timer := time.AfterFunc(time.Second*2, func() {
		wg.Done()
		t.Fatal("timeout waiting for gateway shutdown")
	})
	defer timer.Stop()

	wg.Wait()
}
