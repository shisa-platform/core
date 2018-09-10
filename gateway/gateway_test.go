package gateway

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/sd"
	"github.com/shisa-platform/core/service"
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
		err := cut.Serve(svc)
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

func TestGatewaySignalNotFired(t *testing.T) {
	cut := &Gateway{
		Addr:            ":0",
		HandleInterrupt: true,
	}

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	timer := time.AfterFunc(250*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	err := cut.Serve(svc)
	assert.NoError(t, err)

	timer2 := time.AfterFunc(time.Second*2, func() {
		t.Fatal("interrupt channel not closed!")
	})
	defer timer2.Stop()

	select {
	case <-cut.interrupt:
	}
}

func TestGatewayAddress(t *testing.T) {
	cut := &Gateway{
		Addr: ":0",
	}
	assert.Equal(t, ":0", cut.Address())

	endpoint := service.GetEndpoint(expectedRoute, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})

	timer := time.AfterFunc(200*time.Millisecond, func() { cut.Shutdown() })
	defer timer.Stop()

	go cut.Serve(svc)

	time.Sleep(time.Millisecond * 100)

	assert.NotEqual(t, ":0", cut.Address())
}

func TestGatewayRepr(t *testing.T) {
	config := tls.Config{}
	nextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){
		"lolwut": func(*http.Server, *tls.Conn, http.Handler) {},
	}
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
		RequestIDGenerator: func(context.Context, *httpx.Request) (string, merry.Error) {
			return "", nil
		},
		InternalServerErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			return nil
		},
		NotFoundHandler: func(context.Context, *httpx.Request) httpx.Response {
			return nil
		},
		Registrar: sd.NewFakeRegistrarDefaultFatal(t),
		CheckURLHook: func() (*url.URL, merry.Error) {
			return nil, nil
		},
		ErrorHook:      func(context.Context, *httpx.Request, merry.Error) {},
		CompletionHook: func(context.Context, *httpx.Request, httpx.ResponseSnapshot) {},
	}
	cut.init()

	repr := gatewayExpvar.String()
	assert.NotEmpty(t, repr)

	var expvars map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(repr), &expvars))
	assert.NotEmpty(t, expvars)

	assert.Contains(t, expvars, "start-time")
	assert.Contains(t, expvars, "uptime")
	assert.Contains(t, expvars, "auxiliary")
	assert.Contains(t, expvars, "settings")

	settings, ok := expvars["settings"].(map[string]interface{})
	assert.True(t, ok)

	expectedSettings := map[string]interface{}{
		"Name":                       defaultName,
		"Addr":                       ":9001",
		"HandleInterrupt":            false,
		"DisableKeepAlive":           true,
		"GracePeriod":                "0s",
		"ReadTimeout":                "5ms",
		"ReadHeaderTimeout":          "10ms",
		"WriteTimeout":               "15ms",
		"IdleTimeout":                "20ms",
		"MaxHeaderBytes":             float64(1024),
		"Handlers":                   float64(0),
		"HandlersTimeout":            "0s",
		"RequestIDHeaderName":        defaultRequestIDResponseHeader,
		"TLSConfig":                  "configured",
		"TLSNextProto":               "configured",
		"RequestIDGenerator":         "configured",
		"InternalServerErrorHandler": "configured",
		"NotFoundHandler":            "configured",
		"Registrar":                  "configured",
		"CheckURLHook":               "configured",
		"ErrorHook":                  "configured",
		"CompletionHook":             "configured",
	}
	assert.Equal(t, expectedSettings, settings)
}

func TestGatewayReprEmpty(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	repr := cut.String()
	assert.NotEmpty(t, repr)

	var settings map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(repr), &settings))
	assert.NotEmpty(t, settings)

	expectedSettings := map[string]interface{}{
		"Name":                       defaultName,
		"Addr":                       "",
		"HandleInterrupt":            false,
		"DisableKeepAlive":           false,
		"GracePeriod":                "0s",
		"ReadTimeout":                "0s",
		"ReadHeaderTimeout":          "0s",
		"WriteTimeout":               "0s",
		"IdleTimeout":                "0s",
		"MaxHeaderBytes":             float64(0),
		"Handlers":                   float64(0),
		"HandlersTimeout":            "0s",
		"RequestIDHeaderName":        defaultRequestIDResponseHeader,
		"TLSConfig":                  "unset",
		"TLSNextProto":               "unset",
		"RequestIDGenerator":         "unset",
		"InternalServerErrorHandler": "unset",
		"NotFoundHandler":            "unset",
		"Registrar":                  "unset",
		"CheckURLHook":               "unset",
		"ErrorHook":                  "unset",
		"CompletionHook":             "unset",
	}
	assert.Equal(t, expectedSettings, settings)
}
