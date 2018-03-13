package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/ratelimit"
)

const (
	RetryAfterHeaderKey = "Retry-After"
)

type RateLimitHandler func(context.Context, *httpx.Request, time.Duration) httpx.Response

func (h RateLimitHandler) InvokeSafely(ctx context.Context, request *httpx.Request, cooldown time.Duration) (response httpx.Response, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			exception = merry.Prepend(err1, "panic in rate limit handler")
			return
		} else {
			exception = merry.New("panic in rate limit handler").WithValue("context", arg)
		}
	}()

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
		return nil, merry.New("rate limit provider must be non-nil")
	}
	limiter := &RateLimiter{
		extractor: func(_ context.Context, r *httpx.Request) (string, merry.Error) {
			return r.ClientIP(), nil
		},
		limiter: provider,
	}

	if err := limiter.applyOptions(opts); err != nil {
		return nil, err.Prepend("applying client rate limiter middleware options")
	}

	return limiter, nil
}

// NewUser returns a rate-limiting middleware that throttles
// requests from the context's Actor using the given rate limit
// Provider.
func NewUserLimiter(provider ratelimit.Provider, opts ...RateLimiterOption) (*RateLimiter, merry.Error) {
	if provider == nil {
		return nil, merry.New("rate limit provider must be non-nil")
	}
	limiter := &RateLimiter{
		extractor: func(ctx context.Context, _ *httpx.Request) (string, merry.Error) {
			return ctx.Actor().ID(), nil
		},
		limiter: provider,
	}

	if err := limiter.applyOptions(opts); err != nil {
		return nil, err.Prepend("applying user rate limiter middleware options")
	}

	return limiter, nil
}

// New returns a rate-limiting middleware that throttles requests
// from the given extractor's value using the given rate limit
// Provider.
func NewRateLimiter(provider ratelimit.Provider, extractor httpx.StringExtractor, opts ...RateLimiterOption) (*RateLimiter, merry.Error) {
	if provider == nil {
		return nil, merry.New("rate limit provider must be non-nil")
	}
	if extractor == nil {
		return nil, merry.New("string extractor must be non-nilo")
	}
	limiter := &RateLimiter{
		extractor: extractor,
		limiter:   provider,
	}

	if err := limiter.applyOptions(opts); err != nil {
		return nil, err.Prepend("applying rate limiter middleware options")
	}

	return limiter, nil
}

func (m *RateLimiter) applyOptions(opts []RateLimiterOption) (err merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic in rate limit middleware option")
		} else {
			err = merry.New("panic in rate limit middleware option").WithValue("context", arg)
		}
	}()

	for _, opt := range opts {
		opt(m)
	}

	return nil
}

func (m *RateLimiter) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	if m.limiter == nil {
		err := merry.New("rate limit middleware limiter is nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.handleError(ctx, request, err)
	}
	if m.extractor == nil {
		err := merry.New("rate limit middleware extrator is nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.handleError(ctx, request, err)
	}

	ok, cooldown, err := m.throttle(ctx, request)
	if err != nil {
		return m.handleError(ctx, request, err)
	}
	if !ok {
		return m.handleRateLimit(ctx, request, cooldown)
	}

	return nil
}

func (m *RateLimiter) throttle(ctx context.Context, request *httpx.Request) (ok bool, cd time.Duration, err merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic in rate limit middleware limiter")
		} else {
			err = merry.New("panic in rate limit middleware limiter").WithValue("context", arg)
		}

		err = err.WithHTTPCode(http.StatusInternalServerError)
	}()

	var actor string
	actor, err = m.extractor.InvokeSafely(ctx, request)
	if err != nil {
		err = err.Prepend("running rate limit middleware extractor")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return
	}

	return m.limiter.Allow(actor, request.Method, request.URL.Path)
}

func (m *RateLimiter) handleRateLimit(ctx context.Context, request *httpx.Request, cooldown time.Duration) (response httpx.Response) {
	value := strconv.Itoa(int(cooldown / time.Second))
	if m.Handler == nil {
		response = httpx.NewEmpty(http.StatusTooManyRequests)
		response.Headers().Set(RetryAfterHeaderKey, value)
		return
	}

	var exception merry.Error
	response, exception = m.Handler.InvokeSafely(ctx, request, cooldown)
	if exception != nil {
		exception = exception.Prepend("running rate limit middleware rate limit handler")
		exception = exception.WithHTTPCode(http.StatusTooManyRequests)
		response = m.handleError(ctx, request, exception)
		if response.StatusCode() == http.StatusTooManyRequests {
			response.Headers().Set(RetryAfterHeaderKey, value)
		}
	}

	return
}

func (m *RateLimiter) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	var exception merry.Error
	response := m.ErrorHandler.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		exception = exception.Prepend("running rate limit middleware ErrorHandler")
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}
