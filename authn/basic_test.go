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
	request := &httpx.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestBasicAuthTokenExtractorEmptyHeader(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestBasicAuthTokenExtractorBadChallenge(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Foo Zm9vCg== YmFyCg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestBasicAuthTokenExtractorMissingScheme(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestBasicAuthTokenExtractorBadScheme(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Foo Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestBasicAuthTokenExtractorCorruptCredentials(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Basic x===")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestBasicAuthTokenExtractor(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := BasicAuthTokenExtractor(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, "foo:bar", token)
}

func TestBasicAuthenticatorBadToken(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Basic x===")
	ctx := context.NewFakeContextDefaultFatal(t)

	authn := mustMakeBasicAuthenticator(NewFakeIdentityProviderDefaultFatal(t))

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestBasicAuthenticatorUnknownToken(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return nil, nil
		},
	}
	authn := mustMakeBasicAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NoError(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthenticatorIdPError(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return nil, merry.New("i blewed up!")
		},
	}
	authn := mustMakeBasicAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Error(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestBasicAuthenticator(t *testing.T) {
	request := &httpx.Request{Request: fakeRequest}
	request.Header.Set(AuthnHeaderKey, "Basic Zm9vOmJhcg==")
	ctx := context.NewFakeContextDefaultFatal(t)

	expectedUser := &models.FakeUser{
		IDHook: func() string { return "1" },
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "foo:bar", token)
			return expectedUser, nil
		},
	}
	authn := mustMakeBasicAuthenticator(idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.NoError(t, err)
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
	assert.Error(t, err)
}
