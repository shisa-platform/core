package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
)

func TestErrorHookNil(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var eh ErrorHook
	exception := eh.InvokeSafely(ctx, request, err)
	assert.Nil(t, exception)
}

func TestErrorHookPanic(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var eh ErrorHook
	eh = func(context.Context, *Request, merry.Error) {
		panic(merry.New("i blewed up!"))
	}

	exception := eh.InvokeSafely(ctx, request, err)
	assert.Error(t, exception)
}

func TestErrorHookPanicString(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var eh ErrorHook
	eh = func(context.Context, *Request, merry.Error) {
		panic("i blewed up!")
	}

	exception := eh.InvokeSafely(ctx, request, err)
	assert.Error(t, exception)
}

func TestErrorHookOK(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	err := merry.New("something bad")

	var eh ErrorHook
	eh = func(context.Context, *Request, merry.Error) {
	}

	exception := eh.InvokeSafely(ctx, request, err)
	assert.NoError(t, exception)
}

func TestCompletionHookNil(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	snapshot := ResponseSnapshot{}

	var ch CompletionHook

	exception := ch.InvokeSafely(ctx, request, snapshot)
	assert.Nil(t, exception)
}

func TestCompletionHookPanic(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	snapshot := ResponseSnapshot{}

	var ch CompletionHook
	ch = func(context.Context, *Request, ResponseSnapshot) {
		panic(merry.New("i blewed up!"))
	}

	exception := ch.InvokeSafely(ctx, request, snapshot)
	assert.Error(t, exception)
}

func TestCompletionHookPanicString(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	snapshot := ResponseSnapshot{}

	var ch CompletionHook
	ch = func(context.Context, *Request, ResponseSnapshot) {
		panic("i blewed up!")
	}

	exception := ch.InvokeSafely(ctx, request, snapshot)
	assert.Error(t, exception)
}

func TestCompletionHookOK(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}
	snapshot := ResponseSnapshot{}

	var ch CompletionHook
	ch = func(context.Context, *Request, ResponseSnapshot) {
	}

	exception := ch.InvokeSafely(ctx, request, snapshot)
	assert.NoError(t, exception)
}
