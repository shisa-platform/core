package authn

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestAuthenticationHeaderTokenExtractorMissingHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorEmptyHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorBadChallenge(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Zalgo he comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorMissingScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Zalgo")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorBadScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Slithy he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractor(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Zalgo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "Zalgo")
	assert.Equal(t, "he:comes", token)
	assert.Nil(t, err)
}

func TestURLTokenExtractorMissingUserInfo(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := URLTokenExtractor(ctx, request)
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestURLTokenExtractor(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.URL.User = url.UserPassword("he", "comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := URLTokenExtractor(ctx, request)
	assert.Equal(t, "he:comes", token)
	assert.Nil(t, err)
}
