package gateway

import (
	stdctx "context"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

func TestErrorHookNil(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &httpx.Request{Request: fakeRequest}
	err := merry.New("something bad")
	var exception merry.Error

	var eh ErrorHook
	eh.InvokeSafely(ctx, request, err, &exception)
	assert.Nil(t, exception)
}

func TestErrorHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &httpx.Request{Request: fakeRequest}
	err := merry.New("something bad")
	var exception merry.Error

	var eh ErrorHook
	eh = func(context.Context, *httpx.Request, merry.Error) {
		panic(merry.New("i blewed up!"))
	}

	eh.InvokeSafely(ctx, request, err, &exception)
	assert.Error(t, exception)
}

func TestErrorHookPanicString(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &httpx.Request{Request: fakeRequest}
	err := merry.New("something bad")
	var exception merry.Error

	var eh ErrorHook
	eh = func(context.Context, *httpx.Request, merry.Error) {
		panic("i blewed up!")
	}

	eh.InvokeSafely(ctx, request, err, &exception)
	assert.Error(t, exception)
}

func TestCompletionHookNil(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &httpx.Request{Request: fakeRequest}
	var exception merry.Error

	var ch CompletionHook
	snapshot := httpx.ResponseSnapshot{}
	ch.InvokeSafely(ctx, request, snapshot, &exception)
	assert.Nil(t, exception)
}

func TestCompletionHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &httpx.Request{Request: fakeRequest}
	var exception merry.Error

	var ch CompletionHook
	snapshot := httpx.ResponseSnapshot{}
	ch = func(context.Context, *httpx.Request, httpx.ResponseSnapshot) {
		panic(merry.New("i blewed up!"))
	}

	ch.InvokeSafely(ctx, request, snapshot, &exception)
	assert.Error(t, exception)
}

func TestCompletionHookPanicString(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &httpx.Request{Request: fakeRequest}
	var exception merry.Error

	var ch CompletionHook
	snapshot := httpx.ResponseSnapshot{}
	ch = func(context.Context, *httpx.Request, httpx.ResponseSnapshot) {
		panic("i blewed up!")
	}

	ch.InvokeSafely(ctx, request, snapshot, &exception)
	assert.Error(t, exception)
}
