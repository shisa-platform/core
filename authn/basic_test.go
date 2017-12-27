package authn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

var (
	fakeRequest = httptest.NewRequest(http.MethodGet, "/", nil)
)

func mustMakeBasicProvder(idp IdentityProvider) Provider {
	provider, err := NewBasicAuthenticationProvider(idp, "test")
	if err != nil {
		panic(err)
	}

	return provider
}

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

	authn := mustMakeBasicProvder(NewFakeIdentityProviderDefaultFatal(t))

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
	authn := mustMakeBasicProvder(idp)

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
	authn := mustMakeBasicProvder(idp)

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
	authn := mustMakeBasicProvder(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthProviderChallenge(t *testing.T) {
	authn := mustMakeBasicProvder(NewFakeIdentityProviderDefaultFatal(t))

	challenge := authn.Challenge()
	assert.Equal(t, "Basic realm=\"test\"", challenge)
}

func TestBasicProviderConstructorNilIdp(t *testing.T) {
	provider, err := NewBasicAuthenticationProvider(nil, "bar")
	assert.Nil(t, provider)
	assert.NotNil(t, err)
}
