package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/ratelimit"
)

const (
	RetryAfterHeaderKey = "Retry-After"
)

type RateLimitHandler func(context.Context, *httpx.Request, time.Duration) httpx.Response

func (h RateLimitHandler) InvokeSafely(ctx context.Context, request *httpx.Request, cooldown time.Duration) (response httpx.Response, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in rate limit handler")

	return h(ctx, request, cooldown), nil
}

type RateLimiterOption func(*RateLimiter)

func RateLimiterHandler(handler RateLimitHandler) func(*RateLimiter) {
	return func(t *RateLimiter) {
		t.Handler = handler
	}
}

func RateLimiterErrorHandler(handler httpx.ErrorHandler) func(*RateLimiter) {
	return func(t *RateLimiter) {
		t.ErrorHandler = handler
	}
}

type RateLimiter struct {
	// RateLimitHandler optionally customizes the response for a
	// throttled request. The default handler will return
	// a 429 Too Many Requests response code, an empty body, and
	// the cooldown in seconds in the `Retry-After` header.
	Handler RateLimitHandler

	// ErrorHandler optionally customizes the response for an
	// error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code.
	// The default handler will return the recommended status
	// code and an empty body.
	ErrorHandler httpx.ErrorHandler

	extractor httpx.StringExtractor
	limiter   ratelimit.Provider
}

// NewClient returns a rate-limiting middleware that throttles
// requests from the request's client IP address using the given
// rate limit Provider.
func NewClientLimiter(provider ratelimit.Provider, opts ...RateLimiterOption) (*RateLimiter, merry.Error) {
	if provider == nil {
		return nil, merry.New("rate limit middleware: check invariants: provider is nil")
	}
	limiter := &RateLimiter{
		extractor: func(_ context.Context, r *httpx.Request) (string, merry.Error) {
			return r.ClientIP(), nil
		},
		limiter: provider,
	}

	if exception := limiter.applyOptions(opts); exception != nil {
		return nil, exception.Prepend("rate limit middleware: apply option")
	}

	return limiter, nil
}

// NewUser returns a rate-limiting middleware that throttles
// requests from the context's Actor using the given rate limit
// Provider.
func NewUserLimiter(provider ratelimit.Provider, opts ...RateLimiterOption) (*RateLimiter, merry.Error) {
	if provider == nil {
		return nil, merry.New("rate limit middleware: check invariants: provider is nil")
	}
	limiter := &RateLimiter{
		extractor: func(c context.Context, _ *httpx.Request) (string, merry.Error) {
			actor := c.Actor()
			if actor == nil {
				return "", merry.New("rate limit middleware: extract actor: context actor is nil")
			}
			return actor.ID(), nil
		},
		limiter: provider,
	}

	if exception := limiter.applyOptions(opts); exception != nil {
		return nil, exception.Prepend("rate limit middleware: apply option")
	}

	return limiter, nil
}

// New returns a rate-limiting middleware that throttles requests
// from the given extractor's value using the given rate limit
// Provider.
func NewRateLimiter(provider ratelimit.Provider, extractor httpx.StringExtractor, opts ...RateLimiterOption) (*RateLimiter, merry.Error) {
	if provider == nil {
		return nil, merry.New("rate limit middleware: check invariants: provider is nil")
	}
	if extractor == nil {
		return nil, merry.New("rate limit middleware: check invariants: extractor is nil")
	}
	limiter := &RateLimiter{
		extractor: extractor,
		limiter:   provider,
	}

	if exception := limiter.applyOptions(opts); exception != nil {
		return nil, exception.Prepend("rate limit middleware: apply option")
	}

	return limiter, nil
}

func (m *RateLimiter) applyOptions(opts []RateLimiterOption) (err merry.Error) {
	defer errorx.CapturePanic(&err, "panic in option")

	for _, opt := range opts {
		opt(m)
	}

	return nil
}

func (m *RateLimiter) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	subCtx := ctx
	if ctx.Span() != nil {
		var span opentracing.Span
		span, subCtx = context.StartSpan(ctx, "RateLimit")
		defer span.Finish()
		ext.Component.Set(span, "middleware")
	}

	if m.limiter == nil {
		err := merry.New("rate limit middleware: check invariants: provider is nil")
		return m.handleError(subCtx, request, err)
	}
	if m.extractor == nil {
		err := merry.New("rate limit middleware: check invariants: extractor is nil")
		return m.handleError(subCtx, request, err)
	}

	ok, cooldown, err := m.throttle(subCtx, request)
	if err != nil {
		return m.handleError(subCtx, request, err)
	} else if !ok {
		return m.handleRateLimit(subCtx, request, cooldown)
	}

	return nil
}

func (m *RateLimiter) throttle(ctx context.Context, request *httpx.Request) (ok bool, cd time.Duration, err merry.Error) {
	defer errorx.CapturePanic(&err, "proxy middleware: run provider: panic in provider")

	actor, err, exception := m.extractor.InvokeSafely(ctx, request)
	if exception != nil {
		err = exception.Prepend("proxy middleware: run extractor")
		return
	} else if err != nil {
		err = err.Prepend("proxy middleware: run extractor")
		return
	}

	return m.limiter.Allow(ctx, actor, request.Method, request.URL.Path)
}

func (m *RateLimiter) handleRateLimit(ctx context.Context, request *httpx.Request, cooldown time.Duration) httpx.Response {
	value := strconv.Itoa(int(cooldown / time.Second))
	if m.Handler == nil {
		response := httpx.NewEmpty(http.StatusTooManyRequests)
		response.Headers().Set(RetryAfterHeaderKey, value)
		return response
	}

	response, exception := m.Handler.InvokeSafely(ctx, request, cooldown)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run Handler")
		exception = exception.WithHTTPCode(http.StatusTooManyRequests)
		response = m.handleError(ctx, request, exception)
	}

	if response.StatusCode() == http.StatusTooManyRequests && response.Headers().Get(RetryAfterHeaderKey) == "" {
		response.Headers().Set(RetryAfterHeaderKey, value)
	}

	return response
}

func (m *RateLimiter) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	span := noopSpan
	if ctxSpan := ctx.Span(); ctxSpan != nil {
		span = ctxSpan
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
	}

	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	response, exception := m.ErrorHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run ErrorHandler")
		span.LogFields(otlog.String("exception", exception.Error()))
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}
