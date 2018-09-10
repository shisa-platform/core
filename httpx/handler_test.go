package httpx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
)

func TestHandlerPanic(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var handler Handler
	handler = func(context.Context, *Request) Response {
		panic(merry.New("i blewed up!"))
	}

	response, exception := handler.InvokeSafely(ctx, request)
	assert.Error(t, exception)
	assert.Nil(t, response)
}

func TestHandlerPanicString(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var handler Handler
	handler = func(context.Context, *Request) Response {
		panic("i blewed up!")
	}

	response, exception := handler.InvokeSafely(ctx, request)
	assert.Error(t, exception)
	assert.Nil(t, response)
}

func TestHandlerOK(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var handler Handler
	handler = func(context.Context, *Request) Response {
		return NewEmpty(http.StatusTeapot)
	}

	response, exception := handler.InvokeSafely(ctx, request)
	assert.NoError(t, exception)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTeapot, response.StatusCode())
}

func TestErrorHandlerPanic(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var handler ErrorHandler
	handler = func(context.Context, *Request, merry.Error) Response {
		panic(merry.New("i blewed up!"))
	}

	response, exception := handler.InvokeSafely(ctx, request, err)
	assert.Error(t, exception)
	assert.Nil(t, response)
}

func TestErrorHandlerPanicString(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var handler ErrorHandler
	handler = func(context.Context, *Request, merry.Error) Response {
		panic("i blewed up!")
	}

	response, exception := handler.InvokeSafely(ctx, request, err)
	assert.Error(t, exception)
	assert.Nil(t, response)
}

func TestErrorHandlerOK(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var handler ErrorHandler
	handler = func(context.Context, *Request, merry.Error) Response {
		return NewEmpty(http.StatusTeapot)
	}

	response, exception := handler.InvokeSafely(ctx, request, err)
	assert.NoError(t, exception)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusTeapot, response.StatusCode())
}

func TestAdaptStdHandlerNoOutput(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	handler := &FakeHandler{
		ServeHTTPHook: func(w http.ResponseWriter, r *http.Request) {
			assert.NotNil(t, w)
			assert.Equal(t, ctx, r.Context())
			assert.NotEqual(t, request.Request, r)
			return
		},
	}
	cut := AdaptStandardHandler(handler)

	response := cut(ctx, request)
	assert.Nil(t, response)
	handler.AssertServeHTTPCalledOnce(t)
}

func TestAdaptStdHandlerWithResponse(t *testing.T) {
	status := http.StatusTeapot
	payload := "here is my handle"
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	handler := &FakeHandler{
		ServeHTTPHook: func(w http.ResponseWriter, r *http.Request) {
			assert.NotNil(t, w)
			assert.Equal(t, ctx, r.Context())
			assert.NotEqual(t, request.Request, r)
			w.Header().Set("x-zalgo", "he comes")
			w.WriteHeader(status)
			w.Write([]byte(payload))
			return
		},
	}
	cut := AdaptStandardHandler(handler)

	response := cut(ctx, request)
	assert.NotNil(t, response)
	handler.AssertServeHTTPCalledOnce(t)
	assert.Equal(t, status, response.StatusCode())
	assert.Nil(t, response.Err())
	assert.Equal(t, "he comes", response.Headers().Get("x-zalgo"))
	assert.Nil(t, response.Trailers())

	var bs bytes.Buffer
	err := response.Serialize(&bs)
	assert.NoError(t, err)
	assert.Equal(t, payload, bs.String())
}
