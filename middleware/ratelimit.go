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
		}

		exception = merry.Errorf("panic in rate limit handler: \"%v\"", arg)
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
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic in option")
			return
		}

		err = merry.Errorf("panic in option: \"%v\"", arg)
	}()

	for _, opt := range opts {
		opt(m)
	}

	return nil
}

func (m *RateLimiter) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	if m.limiter == nil {
		err := merry.New("rate limit middleware: check invariants: provider is nil")
		return m.handleError(ctx, request, err)
	}
	if m.extractor == nil {
		err := merry.New("rate limit middleware: check invariants: extractor is nil")
		return m.handleError(ctx, request, err)
	}

	ok, cooldown, err := m.throttle(ctx, request)
	if err != nil {
		return m.handleError(ctx, request, err)
	} else if !ok {
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
			err = merry.Prepend(err1, "proxy middleware: run provider: panic in provider")
			return
		}

		err = merry.Errorf("proxy middleware: run provider: panic in provider: \"%v\"", arg)
	}()

	actor, err, exception := m.extractor.InvokeSafely(ctx, request)
	if exception != nil {
		err = exception.Prepend("proxy middleware: run extractor")
		return
	} else if err != nil {
		err = merry.Prepend(err, "proxy middleware: run extractor")
		return
	}

	return m.limiter.Allow(actor, request.Method, request.URL.Path)
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
	if m.ErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	response, exception := m.ErrorHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		exception = exception.Prepend("proxy middleware: run ErrorHandler")
		exception = exception.Append("original error").Append(err.Error())
		response = httpx.NewEmptyError(merry.HTTPCode(err), exception)
	}

	return response
}
