package authn

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestAuthenticationHeaderTokenExtractorMissingHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestAuthenticationHeaderTokenExtractorEmptyHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestAuthenticationHeaderTokenExtractorBadChallenge(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Zalgo he comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestAuthenticationHeaderTokenExtractorMissingScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Zalgo")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestAuthenticationHeaderTokenExtractorBadScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Foo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestAuthenticationHeaderTokenExtractor(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(AuthnHeaderKey, "Zalgo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Equal(t, "he:comes", token)
	assert.NoError(t, err)
}

func TestURLTokenExtractorMissingUserInfo(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := URLTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestURLTokenExtractor(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.URL.User = url.UserPassword("he", "comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := URLTokenExtractor(ctx, request)
	assert.Equal(t, "he:comes", token)
	assert.NoError(t, err)
}
