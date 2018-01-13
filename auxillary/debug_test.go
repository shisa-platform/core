package auxillary

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

type unserializableResponse struct{}

func (r unserializableResponse) StatusCode() int {
	return 500
}

func (r unserializableResponse) Headers() http.Header {
	return nil
}

func (r unserializableResponse) Trailers() http.Header {
	return nil
}

func (r unserializableResponse) Serialize(io.Writer) (int, error) {
	return 0, merry.New("i blewed up")
}

type stubAuthorizer struct {
	Err merry.Error
}

func (a stubAuthorizer) Authorize(context.Context, *service.Request) merry.Error {
	return a.Err
}

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
			Addr:             "127.0.0.1:0",
			DisableKeepAlive: true,
		},
		Logger: logger,
	}
	assert.Equal(t, "debug", cut.Name())
	assert.Equal(t, "127.0.0.1:0", cut.Address())

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

func TestDebugServerServeHTTPCustomIDGeneratorFail(t *testing.T) {
	cut := DebugServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *service.Request) (string, merry.Error) {
				return "", merry.New("i blewed up!")
			},
		},
		Logger: zap.NewExample(),
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomIDGeneratorEmptyValue(t *testing.T) {
	cut := DebugServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *service.Request) (string, merry.Error) {
				return "", nil
			},
		},
		Logger: zap.NewExample(),
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPCustomIDGeneratorCustomHeader(t *testing.T) {
	requestID := "abc123"
	cut := DebugServer{
		HTTPServer: HTTPServer{
			RequestIDGenerator: func(c context.Context, r *service.Request) (string, merry.Error) {
				return requestID, nil
			},
			RequestIDHeaderName: "x-zalgo",
		},
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get("Content-Type"))
	assert.Equal(t, requestID, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthenticationFail(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
		},
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthenticationWriteFail(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
				UnauthorizedHandler: func(context.Context, *service.Request) service.Response {
					return unserializableResponse{}
				},
			},
		},
		Logger: zap.NewExample(),
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthenticationCustomResponseTrailers(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
				UnauthorizedHandler: func(context.Context, *service.Request) service.Response {
					response := service.NewEmpty(http.StatusUnauthorized)
					response.Headers().Set(middleware.WWWAuthenticateHeaderKey, challenge)
					response.Trailers().Add("x-zalgo", "he comes")
					return response
				},
			},
		},
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

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
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
		},
		Logger: zap.NewExample(),
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
	assert.True(t, w.Flushed)
}

func TestDebugServerServeHTTPAuthorizationFail(t *testing.T) {
	user := &models.FakeUser{IDHook: func() string { return "123" }}
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	authz := stubAuthorizer{merry.New("i blewed up!")}
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
		},
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

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
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	authz := stubAuthorizer{}
	cut := DebugServer{
		HTTPServer: HTTPServer{
			Authentication: &middleware.Authentication{
				Authenticator: authn,
			},
			Authorizer: authz,
		},
		Logger: zap.NewExample(),
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, cut.Path, nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
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
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}
