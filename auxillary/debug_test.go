package auxillary

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDebugServerEmpty(t *testing.T) {
	cut := DebugServer{}

	err := cut.Serve()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
	assert.NotEmpty(t, cut.Path)
	assert.NotNil(t, cut.Logger)
}

func TestGatewayMisconfiguredTLS(t *testing.T) {
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Addr:   ":9900",
			UseTLS: true,
		},
	}

	err := cut.Serve()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
}

func TestDebugServer(t *testing.T) {
	logger := zap.NewExample()
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Addr:             ":9999",
			DisableKeepAlive: true,
		},
		Logger: logger,
	}
	assert.Equal(t, "debug", cut.Name())
	assert.Equal(t, ":9999", cut.Address())

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown(0) })
	defer timer.Stop()

	err := cut.Serve()
	assert.Error(t, err)
	assert.True(t, merry.Is(err, http.ErrServerClosed))
	assert.NotEmpty(t, cut.Path)
	assert.Equal(t, logger, cut.Logger)
}

func TestDebugServerServeHTTPBadPath(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/secrets", nil)
	w := httptest.NewRecorder()

	cut := DebugServer{}
	cut.init()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomPath(t *testing.T) {
	cut := DebugServer{
		Path: "/foo/bar",
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTP(t *testing.T) {
	cut := DebugServer{
		Logger: zap.NewExample(),
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}
