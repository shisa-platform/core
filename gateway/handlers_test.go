package gateway

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func TestDefaultNotFoundHandler(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultNotFoundHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.StatusCode())
	assert.Empty(t, response.Headers())
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultMethodNotAllowedHandler(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultMethodNotAlowedHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode())
	assert.Empty(t, response.Headers())
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultMalformedRequestHandler(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultMalformedRequestHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode())
	assert.Empty(t, response.Headers())
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultInternalServerErrorHandler(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	iseErr := merry.New("i blewed up")
	response := defaultInternalServerErrorHandler(ctx, request, iseErr)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
	assert.Empty(t, response.Headers())
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultRedirectHandlerTrailingSlashGet(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", expectedRoute+"/", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultRedirectHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusSeeOther, response.StatusCode())
	assert.NotEmpty(t, response.Headers())
	assert.Equal(t, expectedRoute, response.Headers().Get(service.LocationHeaderKey))
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultRedirectHandlerTrailingSlashNonGet(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("PUT", expectedRoute+"/", nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultRedirectHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTemporaryRedirect, response.StatusCode())
	assert.NotEmpty(t, response.Headers())
	assert.Equal(t, expectedRoute, response.Headers().Get(service.LocationHeaderKey))
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultRedirectHandlerNoSlashGet(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("GET", expectedRoute, nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultRedirectHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusSeeOther, response.StatusCode())
	assert.NotEmpty(t, response.Headers())
	assert.Equal(t, expectedRoute+"/", response.Headers().Get(service.LocationHeaderKey))
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefaultRedirectHandlerNoSlashNonGet(t *testing.T) {
	request := &service.Request{Request: httptest.NewRequest("PUT", expectedRoute, nil)}
	ctx := context.NewFakeContextDefaultFatal(t)

	response := defaultRedirectHandler(ctx, request)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTemporaryRedirect, response.StatusCode())
	assert.NotEmpty(t, response.Headers())
	assert.Equal(t, expectedRoute+"/", response.Headers().Get(service.LocationHeaderKey))
	assert.Empty(t, response.Trailers())

	var buf bytes.Buffer
	size, err := response.Serialize(&buf)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	assert.Equal(t, 0, buf.Len())
}

func TestDefautlRequestIDGenerator(t *testing.T) {
	request := &service.Request{Request: fakeRequest}
	ctx := context.NewFakeContextDefaultFatal(t)

	rid, err := defaultRequestIDGenerator(ctx, request)
	assert.NoError(t, err)
	assert.NotEmpty(t, rid)
}
