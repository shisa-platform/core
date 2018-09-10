package authn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
)

func mustMakeBearerAuthenticator(idp IdentityProvider) Authenticator {
	authenticator, err := NewBearerAuthenticator(idp, "test")
	if err != nil {
		panic(err)
	}

	return authenticator
}

func TestBearerAuthenticatorBadScheme(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Foo zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	authn := mustMakeBearerAuthenticator(NewFakeIdentityProviderDefaultFatal(t))

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestBearerAuthenticatorUnknownToken(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Bearer zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "zalgo.he:comes", token)
			return nil, nil
		},
	}
	authn := mustMakeBearerAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NoError(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBearerAuthenticatorIdPError(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Bearer zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "zalgo.he:comes", token)
			return nil, merry.New("i blewed up!")
		},
	}
	authn := mustMakeBearerAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Error(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBearerAuthenticator(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Bearer zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	expectedUser := &models.FakeUser{
		IDHook: func() string { return "1" },
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "zalgo.he:comes", token)
			return expectedUser, nil
		},
	}
	authn := mustMakeBearerAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.NoError(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBearerAuthenticatorChallenge(t *testing.T) {
	authn := mustMakeBearerAuthenticator(NewFakeIdentityProviderDefaultFatal(t))

	challenge := authn.Challenge()
	assert.Equal(t, "Bearer realm=\"test\"", challenge)
}

func TestBearerAuthenticatorConstructorNilIdp(t *testing.T) {
	authenticator, err := NewBearerAuthenticator(nil, "bar")
	assert.Nil(t, authenticator)
	assert.Error(t, err)
}
