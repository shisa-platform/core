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

func mustMakeGenericAuthenticator(extractor httpx.StringExtractor, idp IdentityProvider) Authenticator {
	authenticator, err := NewAuthenticator(extractor, idp, "Zalgo", "chaos")
	if err != nil {
		panic(err)
	}

	return authenticator
}

func TestGenericAuthenticatorTokenExtractorError(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Bearer he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	extractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		return "", merry.New("he waits behind the wall")
	}
	authn := mustMakeGenericAuthenticator(extractor, NewFakeIdentityProviderDefaultFatal(t))

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestGenericAuthenticatorIdPError(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Zalgo slithy")
	ctx := context.NewFakeContextDefaultFatal(t)

	extractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		return "slithy", nil
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "slithy", token)
			return nil, merry.New("the <center> cannot hold")
		},
	}
	authn := mustMakeGenericAuthenticator(extractor, idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.Error(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestGenericAuthenticator(t *testing.T) {
	request := &httpx.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Zalgo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	extractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		return "he:comes", nil
	}
	expectedUser := &models.FakeUser{
		IDHook: func() string { return "1" },
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(_ context.Context, token string) (models.User, merry.Error) {
			assert.Equal(t, "he:comes", token)
			return expectedUser, nil
		},
	}
	authn := mustMakeGenericAuthenticator(extractor, idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.NoError(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestGenericAuthenticatorChallenge(t *testing.T) {
	extractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		t.Fatal("unexpected call to extractor")
		return "", nil
	}
	authn := mustMakeGenericAuthenticator(extractor, NewFakeIdentityProviderDefaultFatal(t))

	challenge := authn.Challenge()
	assert.Equal(t, "Zalgo realm=\"chaos\"", challenge)
}

func TestGenericAuthenticatorConstructorNilExtractor(t *testing.T) {
	authenticator, err := NewAuthenticator(nil, NewFakeIdentityProviderDefaultFatal(t), "Foo", "bar")
	assert.Nil(t, authenticator)
	assert.Error(t, err)
}

func TestGenericAuthenticatorConstructorNilIdp(t *testing.T) {
	extractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		t.Fatal("unexpected call to extractor")
		return "", nil
	}
	authenticator, err := NewAuthenticator(extractor, nil, "Foo", "bar")
	assert.Nil(t, authenticator)
	assert.Error(t, err)
}
