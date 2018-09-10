package auxiliary

import (
	"io"
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

func TestDebugServerEmpty(t *testing.T) {
	cut := DebugServer{}

	err := cut.Listen()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
}

func TestDebugServerAddress(t *testing.T) {
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Addr: ":0",
		},
	}

	err := cut.Listen()
	assert.NoError(t, err)
	assert.NotEqual(t, ":0", cut.Address())

	cut.listener.Close()
}

func TestDebugServerMisconfiguredTLS(t *testing.T) {
	cut := DebugServer{
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

func TestDebugServerServeBeforeListen(t *testing.T) {
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Addr: ":0",
		},
	}

	err := cut.Serve()
	assert.Error(t, err)
	assert.False(t, merry.Is(err, http.ErrServerClosed))
}

func TestDebugServer(t *testing.T) {
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Addr:             "127.0.0.1:0",
			DisableKeepAlive: true,
		},
	}
	assert.Equal(t, "debug", cut.Name())
	assert.Equal(t, "127.0.0.1:0", cut.Address())

	timer := time.AfterFunc(50*time.Millisecond, func() { cut.Shutdown(0) })
	defer timer.Stop()

	err := cut.Listen()
	assert.NoError(t, err)
	err = cut.Serve()
	assert.NoError(t, err)
	assert.NotEmpty(t, cut.Path)
}

func TestDebugServerServeHTTPBadPath(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/secrets", nil)
	w := httptest.NewRecorder()

	cut := DebugServer{}
	cut.HTTPServer.init()
	cut.init()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomPath(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHook.Handle,
		},
		Path: "/foo/bar",
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomIDGeneratorFail(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
				return "", merry.New("i blewed up!")
			},
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomIDGeneratorEmptyValue(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
				return "", nil
			},
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomIDGeneratorCustomHeader(t *testing.T) {
	requestID := "abc123"
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
				return requestID, nil
			},
			RequestIDHeaderName: "x-zalgo",
			ErrorHook:           errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, requestID, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthenticationFail(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthenticationWriteFail(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
				UnauthorizedHandler: func(context.Context, *httpx.Request) httpx.Response {
					return unserializableResponse()
				},
			},
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthenticationCustomResponseTrailers(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHook := new(mockErrorHook)
	cut := DebugServer{
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
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.Equal(t, "he comes", w.HeaderMap.Get("x-zalgo"))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthentication(t *testing.T) {
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
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthorizationError(t *testing.T) {
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
	errHook := new(mockErrorHook)
	cut := HealthcheckServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
			ErrorHook:  errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthorizationFail(t *testing.T) {
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
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
			ErrorHook:  errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthorization(t *testing.T) {
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
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
			ErrorHook:  errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTP(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := DebugServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHook.Handle,
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomCompletionHook(t *testing.T) {
	errHook := new(mockErrorHook)
	var handlerCalled bool
	cut := DebugServer{
		HTTPServer: HTTPServer{
			ErrorHook: errHook.Handle,
			CompletionHook: func(context.Context, *httpx.Request, httpx.ResponseSnapshot) {
				handlerCalled = true
			},
		},
	}
	cut.HTTPServer.init()
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertNotCalled(t)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

type failingWriter struct{}

func (w failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.EOF
}

func TestDebugServerResponseWithBadWriter(t *testing.T) {
	cut := expvarResponse{}

	w := failingWriter{}
	err := cut.Serialize(w)
	assert.Error(t, err)
}
