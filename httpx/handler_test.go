package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
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
