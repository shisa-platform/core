package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/ratelimit"
)

func lolExtractor(context.Context, *httpx.Request) (string, merry.Error) {
	return "lol", nil
}

func panicOption(*RateLimiter) {
	panic(merry.New("i blewed up"))
}

func panicStringOption(*RateLimiter) {
	panic("i blewed up")
}

type rateLimitTestCase struct {
	RateLimiter    *RateLimiter
	Provider       *ratelimit.FakeProvider
	ExpectResponse bool
	StatusCode     int
	CoolDown       string
	ExtractorError bool
}

func (c rateLimitTestCase) check(t *testing.T) {
	t.Helper()
	request := &httpx.Request{
		Request: httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil),
	}
	ctx := context.New(request.Context())
	ctx = ctx.WithActor(&models.FakeUser{IDHook: func() string { return "123" }})

	result := c.RateLimiter.Service(ctx, request)

	if result != nil {
		assert.True(t, c.ExpectResponse)
		assert.Equal(t, c.StatusCode, result.StatusCode())
		if c.CoolDown != "" {
			assert.Equal(t, c.CoolDown, result.Headers().Get(RetryAfterHeaderKey))
		}
	} else {
		assert.False(t, c.ExpectResponse)
	}

	actor, err := c.RateLimiter.extractor.InvokeSafely(ctx, request)
	if c.ExtractorError {
		assert.Empty(t, actor)
		assert.Error(t, err)
	} else {
		assert.NotEmpty(t, actor)
		assert.NoError(t, err)
		c.Provider.AssertAllowCalledOnceWith(t, actor, request.Method, request.URL.Path)
	}
}

func TestRateLimiterServiceCustomErrorHandler(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("found a teapot")
		},
	}

	handler := func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		return httpx.NewEmpty(http.StatusTeapot)
	}

	limiter, err := New(provider, lolExtractor, RateLimiterErrorHandler(handler))
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTeapot,
	}

	c.check(t)
}

func TestRateLimiterServiceCustomRateLimitHandler(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		return httpx.NewEmpty(http.StatusTeapot)
	}

	limiter, err := New(provider, lolExtractor, RateLimiterHandler(handler))
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTeapot,
	}

	c.check(t)
}

func TestRateLimiterServiceCustomRateLimitHandlerPanic(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic(merry.New("i blew up!"))
	}

	limiter, err := New(provider, lolExtractor, RateLimiterHandler(handler))
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTooManyRequests,
		CoolDown:       strconv.Itoa(60 * 60),
	}

	c.check(t)
}

func TestRateLimiterServiceCustomRateLimitHandlerPanicErrorHandler(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic(merry.New("i blew up!"))
	}
	errHandler := func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		return httpx.NewEmpty(http.StatusTeapot)
	}

	limiter, err := New(provider, lolExtractor)
	assert.NoError(t, err)

	limiter.Handler = handler
	limiter.ErrorHandler = errHandler

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTeapot,
	}

	c.check(t)
}

func TestRateLimiterServiceCustomRateLimitHandlerAndErrorHandlerBoom(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic(merry.New("i blew up!"))
	}
	errHandler := func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		panic(merry.New("i blewed up too!"))
	}

	limiter, err := New(provider, lolExtractor)
	assert.NoError(t, err)

	limiter.Handler = handler
	limiter.ErrorHandler = errHandler

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTooManyRequests,
		CoolDown:       strconv.Itoa(60 * 60),
	}

	c.check(t)
}

func TestRateLimiterServiceCustomRateLimitHandlerPanicString(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic("i blew up!")
	}

	limiter, err := New(provider, lolExtractor, RateLimiterHandler(handler))
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTooManyRequests,
		CoolDown:       strconv.Itoa(60 * 60),
	}

	c.check(t)
}

func TestClientRateLimiterServiceAllowError(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("something broke")
		},
	}

	limiter, err := NewClient(provider)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
	}

	c.check(t)
}

func TestClientRateLimiterServiceThrottled(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	limiter, err := NewClient(provider)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusTooManyRequests,
		CoolDown:       strconv.Itoa(60 * 60),
	}

	c.check(t)
}

func TestRateLimiterServiceRateLimitProviderPanic(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			panic(merry.New("i blewed up!"))
		},
	}

	limiter, err := NewClient(provider)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
	}

	c.check(t)
}

func TestRateLimiterServiceRateLimitProviderPanicString(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			panic("i blewed up!")
		},
	}

	limiter, err := NewClient(provider)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
	}

	c.check(t)
}

func TestClientRateLimiterOK(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}
	limiter, err := NewClient(provider)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: false,
	}

	c.check(t)
}

func TestUserRateLimiterOK(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}
	limiter, err := NewUser(provider)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: false,
	}

	c.check(t)
}

func TestCustomRateLimiterOK(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}
	limiter, err := New(provider, lolExtractor)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: false,
	}

	c.check(t)
}

func TestRateLimiterConstrctor(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	limiter, err := NewClient(provider)
	assert.NotNil(t, limiter)
	assert.NoError(t, err)
	assert.Nil(t, limiter.Handler)
	assert.Nil(t, limiter.ErrorHandler)

	limiter, err = NewUser(provider)
	assert.NotNil(t, limiter)
	assert.NoError(t, err)
	assert.Nil(t, limiter.Handler)
	assert.Nil(t, limiter.ErrorHandler)

	limiter, err = New(provider, lolExtractor)
	assert.NotNil(t, limiter)
	assert.NoError(t, err)
	assert.Nil(t, limiter.Handler)
	assert.Nil(t, limiter.ErrorHandler)
}

func TestCustomRateLimiterExtractorPanic(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	panicExtractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		panic(merry.New("i blewed up!"))
	}

	limiter, err := New(provider, panicExtractor)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
		ExtractorError: true,
	}

	c.check(t)
}

func TestCustomRateLimiterExtractorPanicString(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	panicExtractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		panic("i blewed up!")
	}

	limiter, err := New(provider, panicExtractor)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
		ExtractorError: true,
	}

	c.check(t)
}

func TestRateLimiterServiceEmpty(t *testing.T) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{Request: httpReq}

	ut := &RateLimiter{}
	res := ut.Service(ctx, req)

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
}

func TestRateLimiterServiceNilLimiter(t *testing.T) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{Request: httpReq}

	ut := &RateLimiter{extractor: lolExtractor}
	res := ut.Service(ctx, req)

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
}

func TestRateLimiterServiceNilExtractor(t *testing.T) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{Request: httpReq}

	ut := &RateLimiter{
		limiter: &ratelimit.FakeProvider{},
	}
	res := ut.Service(ctx, req)

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
}

func TestRateLimiterNewClientNilProvider(t *testing.T) {
	limiter, err := NewClient(nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewClientPanicOption(t *testing.T) {
	provider := &ratelimit.FakeProvider{}

	limiter, err := NewClient(provider, panicOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)

	limiter, err = NewClient(provider, panicStringOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewUserNilProvider(t *testing.T) {
	limiter, err := NewUser(nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewUserPanicOption(t *testing.T) {
	provider := &ratelimit.FakeProvider{}

	limiter, err := NewUser(provider, panicOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)

	limiter, err = NewUser(provider, panicStringOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewrNilProvider(t *testing.T) {
	limiter, err := New(nil, lolExtractor)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewrNilExtractor(t *testing.T) {
	provider := &ratelimit.FakeProvider{}
	limiter, err := New(provider, nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewrAllNil(t *testing.T) {
	limiter, err := New(nil, nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewPanicOption(t *testing.T) {
	provider := &ratelimit.FakeProvider{}

	limiter, err := New(provider, lolExtractor, panicOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)

	limiter, err = New(provider, lolExtractor, panicStringOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}
