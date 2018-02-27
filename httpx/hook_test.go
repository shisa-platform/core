package httpx

import (
	stdctx "context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
)

var (
	fakeRequest = httptest.NewRequest(http.MethodGet, "/test", nil)
)

func TestErrorHookNil(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &Request{Request: fakeRequest}
	err := merry.New("something bad")
	var exception merry.Error

	var eh ErrorHook
	eh.InvokeSafely(ctx, request, err, &exception)
	assert.Nil(t, exception)
}

func TestErrorHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &Request{Request: fakeRequest}
	err := merry.New("something bad")
	var exception merry.Error

	var eh ErrorHook
	eh = func(context.Context, *Request, merry.Error) {
		panic(merry.New("i blewed up!"))
	}

	eh.InvokeSafely(ctx, request, err, &exception)
	assert.Error(t, exception)
}

func TestErrorHookPanicString(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &Request{Request: fakeRequest}
	err := merry.New("something bad")
	var exception merry.Error

	var eh ErrorHook
	eh = func(context.Context, *Request, merry.Error) {
		panic("i blewed up!")
	}

	eh.InvokeSafely(ctx, request, err, &exception)
	assert.Error(t, exception)
}

func TestCompletionHookNil(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &Request{Request: fakeRequest}
	var exception merry.Error

	var ch CompletionHook
	snapshot := ResponseSnapshot{}
	ch.InvokeSafely(ctx, request, snapshot, &exception)
	assert.Nil(t, exception)
}

func TestCompletionHookPanic(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &Request{Request: fakeRequest}
	var exception merry.Error

	var ch CompletionHook
	snapshot := ResponseSnapshot{}
	ch = func(context.Context, *Request, ResponseSnapshot) {
		panic(merry.New("i blewed up!"))
	}

	ch.InvokeSafely(ctx, request, snapshot, &exception)
	assert.Error(t, exception)
}

func TestCompletionHookPanicString(t *testing.T) {
	ctx := context.New(stdctx.Background())
	request := &Request{Request: fakeRequest}
	var exception merry.Error

	var ch CompletionHook
	snapshot := ResponseSnapshot{}
	ch = func(context.Context, *Request, ResponseSnapshot) {
		panic("i blewed up!")
	}

	ch.InvokeSafely(ctx, request, snapshot, &exception)
	assert.Error(t, exception)
}
