package middleware

import (
	stdctx "context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/models"
	"github.com/shisa-platform/core/ratelimit"
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
	ExtractorExcep bool
}

func (c rateLimitTestCase) check(t *testing.T) {
	t.Helper()
	request := &httpx.Request{
		Request: httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil),
	}
	ctx := context.New(request.Context())
	ctx = ctx.WithActor(&models.FakeUser{IDHook: func() string { return "123" }})

	tracer := mocktracer.New()
	span := tracer.StartSpan("test")
	ctx = ctx.WithSpan(span)

	opentracing.SetGlobalTracer(tracer)

	response := c.RateLimiter.Service(ctx, request)

	opentracing.SetGlobalTracer(opentracing.NoopTracer{})

	if response != nil {
		assert.True(t, c.ExpectResponse)
		assert.Equal(t, c.StatusCode, response.StatusCode())
		if c.CoolDown != "" {
			assert.Equal(t, c.CoolDown, response.Headers().Get(RetryAfterHeaderKey))
		}
	} else {
		assert.False(t, c.ExpectResponse)
	}

	actor, err, exception := c.RateLimiter.extractor.InvokeSafely(ctx, request)
	if c.ExtractorExcep {
		assert.Empty(t, actor)
		assert.NoError(t, err)
		assert.Error(t, exception)
	} else if c.ExtractorError {
		assert.Empty(t, actor)
		assert.NoError(t, exception)
		assert.Error(t, err)
	} else {
		assert.NotEmpty(t, actor)
		assert.NoError(t, err)
		assert.NoError(t, exception)
		c.Provider.AssertAllowCalledOnce(t)
	}
}

func TestRateLimiterServiceCustomErrorHandler(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("found a teapot")
		},
	}

	handler := func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		return httpx.NewEmpty(http.StatusTeapot)
	}

	limiter, err := NewRateLimiter(provider, lolExtractor, RateLimiterErrorHandler(handler))
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		return httpx.NewEmpty(http.StatusTeapot)
	}

	limiter, err := NewRateLimiter(provider, lolExtractor, RateLimiterHandler(handler))
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic(merry.New("i blew up!"))
	}

	limiter, err := NewRateLimiter(provider, lolExtractor, RateLimiterHandler(handler))
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic(merry.New("i blew up!"))
	}
	errHandler := func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		return httpx.NewEmpty(http.StatusTeapot)
	}

	limiter, err := NewRateLimiter(provider, lolExtractor)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic(merry.New("i blew up!"))
	}
	errHandler := func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		panic(merry.New("i blewed up too!"))
	}

	limiter, err := NewRateLimiter(provider, lolExtractor)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	handler := func(context.Context, *httpx.Request, time.Duration) httpx.Response {
		panic("i blew up!")
	}

	limiter, err := NewRateLimiter(provider, lolExtractor, RateLimiterHandler(handler))
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("something broke")
		},
	}

	limiter, err := NewClientLimiter(provider)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	limiter, err := NewClientLimiter(provider)
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

func TestUserRateLimiterServiceMissingActor(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	cut, err := NewUserLimiter(provider)
	assert.NoError(t, err)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{Request: httpReq}
	ctx := context.New(request.Context())

	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestRateLimiterServiceRateLimitProviderPanic(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			panic(merry.New("i blewed up!"))
		},
	}

	limiter, err := NewRateLimiter(provider, lolExtractor)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			panic("i blewed up!")
		},
	}

	limiter, err := NewRateLimiter(provider, lolExtractor)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}
	limiter, err := NewClientLimiter(provider)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}
	limiter, err := NewUserLimiter(provider)
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
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}
	limiter, err := NewRateLimiter(provider, lolExtractor)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: false,
	}

	c.check(t)
}

func TestRateLimiterConstructor(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	limiter, err := NewClientLimiter(provider)
	assert.NotNil(t, limiter)
	assert.NoError(t, err)
	assert.Nil(t, limiter.Handler)
	assert.Nil(t, limiter.ErrorHandler)

	limiter, err = NewUserLimiter(provider)
	assert.NotNil(t, limiter)
	assert.NoError(t, err)
	assert.Nil(t, limiter.Handler)
	assert.Nil(t, limiter.ErrorHandler)

	limiter, err = NewRateLimiter(provider, lolExtractor)
	assert.NotNil(t, limiter)
	assert.NoError(t, err)
	assert.Nil(t, limiter.Handler)
	assert.Nil(t, limiter.ErrorHandler)
}

func TestCustomRateLimiterExtractorPanic(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	panicExtractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		panic(merry.New("i blewed up!"))
	}

	limiter, err := NewRateLimiter(provider, panicExtractor)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
		ExtractorExcep: true,
	}

	c.check(t)
}

func TestCustomRateLimiterExtractorPanicString(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	panicExtractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		panic("i blewed up!")
	}

	limiter, err := NewRateLimiter(provider, panicExtractor)
	assert.NoError(t, err)

	c := rateLimitTestCase{
		RateLimiter:    limiter,
		Provider:       provider,
		ExpectResponse: true,
		StatusCode:     http.StatusInternalServerError,
		ExtractorExcep: true,
	}

	c.check(t)
}

func TestCustomRateLimiterExtractorError(t *testing.T) {
	provider := &ratelimit.FakeProvider{
		AllowHook: func(context.Context, string, string, string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	extractor := func(context.Context, *httpx.Request) (string, merry.Error) {
		return "", merry.New("i blewed up!")
	}

	limiter, err := NewRateLimiter(provider, extractor)
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
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{Request: httpReq}

	cut := &RateLimiter{}
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestRateLimiterServiceNilLimiter(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{Request: httpReq}

	cut := &RateLimiter{extractor: lolExtractor}
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestRateLimiterServiceNilExtractor(t *testing.T) {
	ctx := context.New(stdctx.Background())

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	request := &httpx.Request{Request: httpReq}

	cut := &RateLimiter{
		limiter: &ratelimit.FakeProvider{},
	}
	response := cut.Service(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestRateLimiterNewClientLimiterNilProvider(t *testing.T) {
	limiter, err := NewClientLimiter(nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewClientLimiterPanicOption(t *testing.T) {
	provider := &ratelimit.FakeProvider{}

	limiter, err := NewClientLimiter(provider, panicOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)

	limiter, err = NewClientLimiter(provider, panicStringOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewUserLimiterNilProvider(t *testing.T) {
	limiter, err := NewUserLimiter(nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewUserLimiterPanicOption(t *testing.T) {
	provider := &ratelimit.FakeProvider{}

	limiter, err := NewUserLimiter(provider, panicOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)

	limiter, err = NewUserLimiter(provider, panicStringOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewrNilProvider(t *testing.T) {
	limiter, err := NewRateLimiter(nil, lolExtractor)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewrNilExtractor(t *testing.T) {
	provider := &ratelimit.FakeProvider{}
	limiter, err := NewRateLimiter(provider, nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewrAllNil(t *testing.T) {
	limiter, err := NewRateLimiter(nil, nil)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}

func TestRateLimiterNewPanicOption(t *testing.T) {
	provider := &ratelimit.FakeProvider{}

	limiter, err := NewRateLimiter(provider, lolExtractor, panicOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)

	limiter, err = NewRateLimiter(provider, lolExtractor, panicStringOption)
	assert.Nil(t, limiter)
	assert.Error(t, err)
}
