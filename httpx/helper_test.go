package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
)

func TestStringExtractorPanic(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper StringExtractor
	helper = func(context.Context, *Request) (string, merry.Error) {
		panic(merry.New("i blewed up!"))
	}

	str, err, exception := helper.InvokeSafely(ctx, request)
	assert.Error(t, exception)
	assert.NoError(t, err)
	assert.Empty(t, str)
}

func TestStringExtractorPanicString(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper StringExtractor
	helper = func(context.Context, *Request) (string, merry.Error) {
		panic("i blewed up!")
	}

	str, err, exception := helper.InvokeSafely(ctx, request)
	assert.Error(t, exception)
	assert.NoError(t, err)
	assert.Empty(t, str)
}

func TestStringExtractorError(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper StringExtractor
	helper = func(context.Context, *Request) (string, merry.Error) {
		return "", merry.New("i blewed up!")
	}

	str, err, exception := helper.InvokeSafely(ctx, request)
	assert.NoError(t, exception)
	assert.Error(t, err)
	assert.Empty(t, str)
}

func TestStringExtractorOK(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper StringExtractor
	helper = func(context.Context, *Request) (string, merry.Error) {
		return "lol", nil
	}

	str, err, exception := helper.InvokeSafely(ctx, request)
	assert.NoError(t, exception)
	assert.NoError(t, err)
	assert.Equal(t, "lol", str)
}

func TestRequestPredicatePanic(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper RequestPredicate
	helper = func(context.Context, *Request) bool {
		panic(merry.New("i blewed up!"))
	}

	ok, exception := helper.InvokeSafely(ctx, request)
	assert.Error(t, exception)
	assert.False(t, ok)
}

func TestRequestPredicatePanicString(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper RequestPredicate
	helper = func(context.Context, *Request) bool {
		panic("i blewed up!")
	}

	ok, exception := helper.InvokeSafely(ctx, request)
	assert.Error(t, exception)
	assert.False(t, ok)
}

func TestRequestPredicateOK(t *testing.T) {
	ctx := context.NewFakeContextDefaultFatal(t)
	request := &Request{Request: httptest.NewRequest(http.MethodGet, "/test", nil)}

	var helper RequestPredicate
	helper = func(context.Context, *Request) bool {
		return true
	}

	ok, exception := helper.InvokeSafely(ctx, request)
	assert.NoError(t, exception)
	assert.True(t, ok)
}
