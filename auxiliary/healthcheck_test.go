package auxiliary

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/authn"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/middleware"
	"github.com/shisa-platform/core/models"
)

type stubHealthchecker struct {
	name string
	err  merry.Error
}

func (h stubHealthchecker) Name() string {
	return h.name
}

func (h stubHealthchecker) Healthcheck(context.Context) merry.Error {
	return h.err
}

type panicHealthchecker struct {
	name string
	arg  interface{}
}

func (h panicHealthchecker) Name() string {
	return h.name
}

func (h panicHealthchecker) Healthcheck(context.Context) merry.Error {
	panic(h.arg)
}

func TestHealthcheckServerEmpty(t *testing.T) {
	cut := HealthcheckServer{}

	err := cut.Listen()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
}

func TestHealthcheckServerAddress(t *testing.T) {
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Addr: ":0",
		},
	}

	err := cut.Listen()
	assert.NoError(t, err)
	assert.NotEqual(t, ":0", cut.Address())

	cut.listener.Close()
}

func TestHealthcheckServerMisconfiguredTLS(t *testing.T) {
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Addr:   ":0",
			UseTLS: true,
		},
	}

	err := cut.Listen()
	assert.NoError(t, err)
	err = cut.Serve()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
}

func TestHealthcheckServerServeBeforeListen(t *testing.T) {
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Addr: ":0",
		},
	}

	err := cut.Serve()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
}

func TestHealthcheckServer(t *testing.T) {
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Addr:             "127.0.0.1:0",
			DisableKeepAlive: true,
		},
	}
	assert.Equal(t, "healthcheck", cut.Name())
	assert.Equal(t, "127.0.0.1:0", cut.Address())

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown(0) })
	defer timer.Stop()

	err := cut.Listen()
	assert.NoError(t, err)
	err = cut.Serve()
	assert.NoError(t, err)
	assert.NotEmpty(t, cut.Path)
}

func TestHealthcheckServerServeHTTPBadPath(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/plonk", nil)
	w := httptest.NewRecorder()

	cut := HealthcheckServer{}
	cut.HTTPServer.init()
	cut.init()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, "text/plain; charset=utf-8", w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPCustomPath(t *testing.T) {
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHandler.Handle,
		},
		Path: "/foo/bar",
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.Equal(t, "application/json", w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPCustomIDGeneratorFail(t *testing.T) {
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
				return "", merry.New("i blewed up!")
			},
			ErrorHook: errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.Equal(t, "application/json", w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPCustomIDGeneratorEmptyValue(t *testing.T) {
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
				return "", nil
			},
			ErrorHook: errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.Equal(t, "application/json", w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPCustomIDGeneratorCustomHeader(t *testing.T) {
	requestID := "abc123"
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
				return requestID, nil
			},
			RequestIDHeaderName: "x-zalgo",
			ErrorHook:           errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.Equal(t, "application/json", w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, requestID, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthenticationFail(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			ErrorHook: errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthenticationWriteFail(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
				UnauthorizedHandler: func(context.Context, *httpx.Request) httpx.Response {
					return unserializableResponse()
				},
			},
			ErrorHook: errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthenticationCustomResponseTrailers(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
				UnauthorizedHandler: func(context.Context, *httpx.Request) httpx.Response {
					response := httpx.NewEmpty(http.StatusUnauthorized)
					response.Headers().Set(middleware.WWWAuthenticateHeaderKey, challenge)
					response.Trailers().Add("x-zalgo", "he comes")
					return response
				},
			},
			ErrorHook: errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.Equal(t, "he comes", w.HeaderMap.Get("x-zalgo"))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthentication(t *testing.T) {
	user := &models.FakeUser{IDHook: func() string { return "123" }}
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			ErrorHook: errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthorizationError(t *testing.T) {
	user := &models.FakeUser{IDHook: func() string { return "123" }}
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	authz := stubAuthorizer{err: merry.New("i blewed up!")}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
			ErrorHook:  errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthorizationFail(t *testing.T) {
	user := &models.FakeUser{IDHook: func() string { return "123" }}
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	authz := stubAuthorizer{ok: false}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
			ErrorHook:  errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTPAuthorization(t *testing.T) {
	user := &models.FakeUser{IDHook: func() string { return "123" }}
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	authz := stubAuthorizer{ok: true}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
			ErrorHook:  errHandler.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestHealthcheckServerServeHTTP(t *testing.T) {
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHandler.Handle,
		},
		Checkers: []Healthchecker{stubHealthchecker{name: "pass"}},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.Equal(t, "application/json", w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)

	expectedJson := `{
  "pass": "OK"
}`
	assert.JSONEq(t, expectedJson, w.Body.String())
}

func TestHealthcheckServerServeHTTPCustomCompletionHook(t *testing.T) {
	errHandler := new(mockErrorHook)
	var handlerCalled bool
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHandler.Handle,
			CompletionHook: func(context.Context, *httpx.Request, httpx.ResponseSnapshot) {
				handlerCalled = true
			},
		},
		Checkers: []Healthchecker{stubHealthchecker{name: "pass"}},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertNotCalled(t)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.Equal(t, "application/json", w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)

	expectedJson := `{
  "pass": "OK"
}`
	assert.JSONEq(t, expectedJson, w.Body.String())
}

func TestHealthcheckServerServeHTTPFailingCheck(t *testing.T) {
	msg := "i blewed up!"
	err := merry.New(msg)
	ng := stubHealthchecker{name: "fail", err: err}
	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHandler.Handle,
		},
		Checkers: []Healthchecker{stubHealthchecker{name: "pass"}, ng},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)

	expectedJson := `{
  "pass": "OK",
  "fail": "i blewed up!"
}`
	assert.JSONEq(t, expectedJson, w.Body.String())
}

func TestHealthcheckServerServeHTTPHealthcheckerPanic(t *testing.T) {
	ng := panicHealthchecker{name: "fail", arg: merry.New("i blewed up!")}

	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHandler.Handle,
		},
		Checkers: []Healthchecker{stubHealthchecker{name: "pass"}, ng},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)

	expectedJson := `{
  "pass": "OK",
  "fail": "panic in healthcheck: i blewed up!"
}`
	assert.JSONEq(t, expectedJson, w.Body.String())
}

func TestHealthcheckServerServeHTTPHealthcheckerPanicString(t *testing.T) {
	ng := panicHealthchecker{name: "fail", arg: "i blewed up!"}

	errHandler := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHandler.Handle,
		},
		Checkers: []Healthchecker{stubHealthchecker{name: "pass"}, ng},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHandler.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)

	expectedJson := `{
  "pass": "OK",
  "fail": "panic in healthcheck: i blewed up!"
}`
	assert.JSONEq(t, expectedJson, w.Body.String())
}
