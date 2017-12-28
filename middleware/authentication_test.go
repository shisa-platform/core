package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

var (
	expectedUser      = &models.FakeUser{IDHook: func() string { return "123" }}
	fakeRequest       = httptest.NewRequest(http.MethodGet, "/", nil)
	expectedChallenge = "Brute realm=\"Outer Space\""
)

func TestAuthenticationNilAuthenticator(t *testing.T) {
	cut := &Authentication{}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, "", response.Headers().Get(wwwAuthenticateHeaderKey))
}

func TestAuthenticationAuthenticatorError(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(wwwAuthenticateHeaderKey))

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationOK(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return expectedUser, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	ctx.WithActorHook = func(value models.User) context.Context { return ctx }

	response := cut.Service(ctx, request)
	assert.Nil(t, response)
	assert.NotNil(t, cut.ErrorHandler)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
	ctx.AssertWithActorCalledOnceWith(t, expectedUser)
}

func TestAuthenticationUnauthorized(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	cut := &Authentication{
		Authenticator: authn,
	}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode())
	assert.Equal(t, expectedChallenge, response.Headers().Get(wwwAuthenticateHeaderKey))

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomHandler(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	challenge := "Custom realm=\"secrets, inc\""
	var handlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		UnauthorizedHandler: func(c context.Context, r *service.Request) service.Response {
			handlerInvoked = true
			assert.Equal(t, ctx, c)
			assert.Equal(t, request, r)

			response := service.NewEmpty(http.StatusForbidden)
			response.Headers().Set(wwwAuthenticateHeaderKey, challenge)
			return response
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
	assert.Equal(t, challenge, response.Headers().Get(wwwAuthenticateHeaderKey))

	if !handlerInvoked {
		t.Fatal("custom error handler not invoked")
	}

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestAuthenticationCustomErrorHandler(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
		ChallengeHook: func() string {
			return expectedChallenge
		},
	}

	challenge := "Custom realm=\"secrets, inc\""
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	var errorHandlerInvoked bool
	cut := &Authentication{
		Authenticator: authn,
		ErrorHandler: func(c context.Context, r *service.Request, err merry.Error) service.Response {
			errorHandlerInvoked = true
			assert.Equal(t, ctx, c)
			assert.Equal(t, request, r)
			assert.NotNil(t, err)
			assert.Equal(t, http.StatusUnauthorized, merry.HTTPCode(err))

			response := service.NewEmpty(http.StatusForbidden)
			response.Headers().Set(wwwAuthenticateHeaderKey, challenge)
			return response
		},
	}

	response := cut.Service(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode())
	assert.Equal(t, challenge, response.Headers().Get(wwwAuthenticateHeaderKey))

	if !errorHandlerInvoked {
		t.Fatal("custom error handler not invoked")
	}

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationNilAuthenticator(t *testing.T) {
	cut := &PassiveAuthentication{}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Equal(t, "", response.Headers().Get(wwwAuthenticateHeaderKey))
}

func TestPassiveAuthenticationAuthenticatorError(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
			return nil, merry.New("I blewed up!")
		},
	}

	cut := &PassiveAuthentication{
		Authenticator: authn,
	}

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := cut.Service(ctx, request)
	assert.Nil(t, response)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationUnknownPrincipal(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
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

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := cut.Service(ctx, request)
	assert.Nil(t, response)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
}

func TestPassiveAuthenticationOK(t *testing.T) {
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *service.Request) (models.User, merry.Error) {
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

	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)
	ctx.WithActorHook = func(value models.User) context.Context { return ctx }

	response := cut.Service(ctx, request)
	assert.Nil(t, response)

	authn.AssertAuthenticateCalledOnceWith(t, ctx, request)
	ctx.AssertWithActorCalledOnceWith(t, expectedUser)
}
