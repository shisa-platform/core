package authn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestAuthenticationHeaderTokenExtractorMissingHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorEmptyHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorBadChallenge(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "zalgo he comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorMissingScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "zalgo")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractorBadScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "Foo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestAuthenticationHeaderTokenExtractor(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest(http.MethodGet, "/", nil)}
	request.Header.Set(authHeaderKey, "zalgo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := AuthenticationHeaderTokenExtractor(ctx, request, "zalgo")
	assert.Equal(t, "he:comes", token)
	assert.Nil(t, err)
}
