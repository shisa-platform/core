package middleware

import (
	stdctx "context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/authn"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
)

var (
	expectedUser      = &models.FakeUser{IDHook: func() string { return "123" }}
	fakeRequest       = httptest.NewRequest(http.MethodGet, "/", nil)
	expectedChallenge = "Brute realm=\"Outer Space\""
)

func TestAuthenticationNilAuthenticator(t *testing.T) {
	cut := &Authentication{}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, "", response.Headers().Get(WWWAuthenticateHeaderKey))
}

func TestAuthenticationAuthenticatorError(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	tracer := mocktracer.New()
	span := tracer.StartSpan("test")
	ctx = ctx.WithSpan(span)

	opentracing.SetGlobalTracer(tracer)

	response := cut.Service(ctx, request)

	opentracing.SetGlobalTracer(opentracing.NoopTracer{})

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(WWWAuthenticateHeaderKey))

	authn.AssertAuthenticateCalledOnce(t)
}

func TestAuthenticationAuthenticatorPanic(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			panic(merry.New("I blewed up!"))
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationAuthenticatorPanicString(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			panic("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationOK(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return expectedUser, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.Nil(t, response)
	assert.Nil(t, cut.ErrorHandler)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationUnauthorized(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(WWWAuthenticateHeaderKey))

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomHandler(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var handlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		UnauthorizedHandler: func(c context.Context, r *httpx.Request) httpx.Response {
			handlerInvoked = true
			assert.Equal(t, ctx, c)
			assert.Equal(t, request, r)

			response := httpx.NewEmpty(http.StatusForbidden)
			return response
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
	assert.Empty(t, response.Headers().Get(WWWAuthenticateHeaderKey))
	assert.True(t, handlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomHandlerWithoutSettingHeader(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var handlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		UnauthorizedHandler: func(c context.Context, r *httpx.Request) httpx.Response {
			handlerInvoked = true
			assert.Equal(t, ctx, c)
			assert.Equal(t, request, r)

			response := httpx.NewEmpty(http.StatusUnauthorized)
			return response
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(WWWAuthenticateHeaderKey))
	assert.True(t, handlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomHandlerPanic(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var handlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		UnauthorizedHandler: func(c context.Context, r *httpx.Request) httpx.Response {
			handlerInvoked = true
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.True(t, handlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomHandlerPanicString(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var handlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		UnauthorizedHandler: func(c context.Context, r *httpx.Request) httpx.Response {
			handlerInvoked = true
			panic("i blewed up!")
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.True(t, handlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomErrorHandler(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	challenge := "Custom realm=\"secrets, inc\""
	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var errorHandlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			errorHandlerInvoked = true
			assert.Equal(t, ctx, c)
			assert.Equal(t, request, r)
			assert.Error(t, err)
			assert.Equal(t, http.StatusUnauthorized, merry.HTTPCode(err))

			response := httpx.NewEmpty(http.StatusForbidden)
			response.Headers().Set(WWWAuthenticateHeaderKey, challenge)
			return response
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
	assert.Equal(t, challenge, response.Headers().Get(WWWAuthenticateHeaderKey))
	assert.True(t, errorHandlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomErrorHandlerPanic(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var errorHandlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			errorHandlerInvoked = true
			panic(merry.New("i blewed up!"))
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(WWWAuthenticateHeaderKey))
	assert.True(t, errorHandlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomErrorHandlerPanicString(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	var errorHandlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			errorHandlerInvoked = true
			panic("i blewed up!")
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(WWWAuthenticateHeaderKey))
	assert.True(t, errorHandlerInvoked, "custom error handler not invoked")

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationNilAuthenticator(t *testing.T) {
	cut := &PassiveAuthentication{}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, "", response.Headers().Get(WWWAuthenticateHeaderKey))
}

func TestPassiveAuthenticationAuthenticatorError(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
	}

	cut := &PassiveAuthentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	tracer := mocktracer.New()
	span := tracer.StartSpan("test")
	ctx = ctx.WithSpan(span)

	opentracing.SetGlobalTracer(tracer)

	response := cut.Service(ctx, request)

	opentracing.SetGlobalTracer(opentracing.NoopTracer{})

	assert.Nil(t, response)

	authn.AssertAuthenticateCalledOnce(t)
}

func TestPassiveAuthenticationAuthenticatorPanic(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			panic(merry.New("i blewed up!"))
		},
	}

	cut := &PassiveAuthentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, "", response.Headers().Get(WWWAuthenticateHeaderKey))

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationAuthenticatorPanicString(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			panic("i blewed up!")
		},
	}

	cut := &PassiveAuthentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, "", response.Headers().Get(WWWAuthenticateHeaderKey))

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationUnknownPrincipal(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			t.Fatal("unexpected call to Challenge")
			return ""
		},
	}

	cut := &PassiveAuthentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.Nil(t, response)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationOK(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return expectedUser, nil
		},
		ChallengeHook: func() string {
			t.Fatal("unexpected call to Challenge")
			return ""
		},
	}

	cut := &PassiveAuthentication{
		Authenticator: authn,
	}

	request := &httpx.Request{Request: fakeRequest}
	ctx := context.New(stdctx.Background())

	response := cut.Service(ctx, request)
	assert.Nil(t, response)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}
