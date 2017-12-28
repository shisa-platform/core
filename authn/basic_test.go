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

func mustMakeBasicAuthenticator(idp IdentityProvider) Authenticator {
	authenticator, err := NewBasicAuthenticator(idp, "test")
	if err != nil {
		panic(err)
	}

	return authenticator
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

func TestBasicAuthenticatorBadToken(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic x===")
	ctx := context.NewFakeContextDefaultFatal(t)

	authn := mustMakeBasicAuthenticator(NewFakeIdentityProviderDefaultFatal(t))

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
}

func TestBasicAuthenticatorUnknownToken(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return nil, nil
		},
	}
	authn := mustMakeBasicAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthenticatorIdPError(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	request.Header.Set(authHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return nil, merry.New("i blewed up!")
		},
	}
	authn := mustMakeBasicAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthenticator(t *testing.T) {
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
	authn := mustMakeBasicAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthenticatorChallenge(t *testing.T) {
	authn := mustMakeBasicAuthenticator(NewFakeIdentityProviderDefaultFatal(t))

	challenge := authn.Challenge()
	assert.Equal(t, "Basic realm=\"test\"", challenge)
}

func TestBasicProviderConstructorNilIdp(t *testing.T) {
	provider, err := NewBasicAuthenticator(nil, "bar")
	assert.Nil(t, provider)
	assert.NotNil(t, err)
}
