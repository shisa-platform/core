package authn

import (
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

func TestBearerProviderBadScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Foo zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	authn := &BearerAuthnProvider{
		IdP:   NewFakeIdentityProviderDefaultFatal(t),
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
}

func TestBearerProviderUnknownToken(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Bearer zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "zalgo.he:comes", token)
			return nil, nil
		},
	}
	authn := &BearerAuthnProvider{
		IdP:   idp,
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBearerProviderIdPError(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Bearer zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "zalgo.he:comes", token)
			return nil, merry.New("i blewed up!")
		},
	}
	authn := &BearerAuthnProvider{
		IdP:   idp,
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBearerProvider(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Bearer zalgo.he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	expectedUser := &models.FakeUser{
		IDHook: func() string { return "1" },
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "zalgo.he:comes", token)
			return expectedUser, nil
		},
	}
	authn := &BearerAuthnProvider{
		IdP:   idp,
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBearerProviderChallenge(t *testing.T) {
	authn := &BearerAuthnProvider{
		IdP:   NewFakeIdentityProviderDefaultFatal(t),
		Realm: "test",
	}

	challenge := authn.Challenge()
	assert.Equal(t, "Bearer realm=\"test\"", challenge)
}
