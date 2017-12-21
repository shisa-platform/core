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

var (
	fakeRequest = httptest.NewRequest("GET", "/foo", nil)
)

func TestBasicAuthTokenExtractorMissingHeader(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestBasicAuthTokenExtractorEmptyHeader(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestBasicAuthTokenExtractorBadChallenge(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Foo Zm9vCg== YmFyCg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestBasicAuthTokenExtractorMissingScheme(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestBasicAuthTokenExtractorBadScheme(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Foo Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestBasicAuthTokenExtractorCorruptCredentials(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic x===")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestBasicAuthTokenExtractor(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, "foo:bar", token)
}

func TestBasicAuthProviderBadToken(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic x===")
	ctx := context.NewFakeContextDefaultFatal(t)

	authn := &BasicAuthnProvider{
		IdP:   NewFakeIdentityProviderDefaultFatal(t),
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
}

func TestBasicAuthProviderUnknownToken(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return nil, nil
		},
	}
	authn := &BasicAuthnProvider{
		IdP:   idp,
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthProviderIdPError(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return nil, merry.New("i blewed up!")
		},
	}
	authn := &BasicAuthnProvider{
		IdP:   idp,
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthProvider(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	expectedUser := &models.FakeUser{
		IDHook: func() string { return "1" },
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return expectedUser, nil
		},
	}
	authn := &BasicAuthnProvider{
		IdP:   idp,
		Realm: "test",
	}

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthProviderChallenge(t *testing.T) {
	authn := &BasicAuthnProvider{
		IdP:   NewFakeIdentityProviderDefaultFatal(t),
		Realm: "test",
	}

	challenge := authn.Challenge()
	assert.Equal(t, "Basic realm=\"test\"", challenge)
}
