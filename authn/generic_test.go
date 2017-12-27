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

func mustMakeGenericProvder(extractor service.StringExtractor, idp IdentityProvider) Provider {
	provider, err := NewProvider(extractor, idp, "Zalgo", "chaos")
	if err != nil {
		panic(err)
	}

	return provider
}

func TestGenericProviderTokenExtractorError(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "Bearer he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	extractor := func(context.Context, *service.Request) (string, merry.Error) {
		return "", merry.New("he waits behind the wall")
	}
	authn := mustMakeGenericProvder(extractor, NewFakeIdentityProviderDefaultFatal(t))

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
}

func TestGenericProviderIdPError(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "Zalgo slithy")
	ctx := context.NewFakeContextDefaultFatal(t)

	extractor := func(context.Context, *service.Request) (string, merry.Error) {
		return "slithy", nil
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "slithy", token)
			return nil, merry.New("the <center> cannot hold")
		},
	}
	authn := mustMakeGenericProvder(extractor, idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Nil(t, user)
	assert.NotNil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestGenericProvider(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "Zalgo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	extractor := func(context.Context, *service.Request) (string, merry.Error) {
		return "he:comes", nil
	}
	expectedUser := &models.FakeUser{
		IDHook: func() string { return "1" },
	}
	idp := &FakeIdentityProvider{
		AuthenticateHook: func(token string) (models.User, merry.Error) {
			assert.Equal(t, "he:comes", token)
			return expectedUser, nil
		},
	}
	authn := mustMakeGenericProvder(extractor, idp)

	user, err := authn.Authenticate(ctx, request)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
	idp.AssertAuthenticateCalledOnce(t)
}

func TestGenericProviderChallenge(t *testing.T) {
	extractor := func(context.Context, *service.Request) (string, merry.Error) {
		t.Fatal("unexpected call to extractor")
		return "", nil
	}
	authn := mustMakeGenericProvder(extractor, NewFakeIdentityProviderDefaultFatal(t))

	challenge := authn.Challenge()
	assert.Equal(t, "Zalgo realm=\"chaos\"", challenge)
}

func TestGenericProviderConstructorNilExtractor(t *testing.T) {
	provider, err := NewProvider(nil, NewFakeIdentityProviderDefaultFatal(t), "Foo", "bar")
	assert.Nil(t, provider)
	assert.NotNil(t, err)
}

func TestGenericProviderConstructorNilIdp(t *testing.T) {
	extractor := func(context.Context, *service.Request) (string, merry.Error) {
		t.Fatal("unexpected call to extractor")
		return "", nil
	}
	provider, err := NewProvider(extractor, nil, "Foo", "bar")
	assert.Nil(t, provider)
	assert.NotNil(t, err)
}
