package authn

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestCommonTokenExtractorMissingHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := CommonTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestCommonTokenExtractorEmptyHeader(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := CommonTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestCommonTokenExtractorBadChallenge(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "zalgo he comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := CommonTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestCommonTokenExtractorMissingScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "zalgo")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := CommonTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestCommonTokenExtractorBadScheme(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "Foo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := CommonTokenExtractor(ctx, request, "zalgo")
	assert.Empty(t, token)
	assert.NotNil(t, err)
}

func TestCommonTokenExtractor(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", "/foo", nil)}
	request.Header.Set(authHeaderKey, "zalgo he:comes")
	ctx := context.NewFakeContextDefaultFatal(t)

	token, err := CommonTokenExtractor(ctx, request, "zalgo")
	assert.Equal(t, "he:comes", token)
	assert.Nil(t, err)
}
