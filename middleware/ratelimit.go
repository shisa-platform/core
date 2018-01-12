package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/ratelimit"
	"github.com/percolate/shisa/service"
)

const (
	RetryAfterHeaderKey = "Retry-After"
)

type RateLimitHandler func(context.Context, *service.Request, time.Duration) service.Response

// ClientThrottler is a a rate-limiting middleware that
// throttles requests from a given ClientIP using its Limiter
type ClientThrottler struct {
	// Limiter determines whether the request should be throttled
	// based on the request's ClientIP.  This must be non-nil or
	// an InternalServiceError status response will be returned.
	Limiter ratelimit.Provider

	// RateLimitHandler optionally customizes the response for a
	// throttled request. The default handler will return
	// a 429 Too Many Requests response code, an empty body, and
	// the cooldown in seconds in the `Retry-After` header.
	RateLimitHandler RateLimitHandler

	// ErrorHandler optionally customizes the response for an
	// error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code.
	// The default handler will return the recommended status
	// code and an empty body.
	ErrorHandler service.ErrorHandler
}

func (m *ClientThrottler) Service(ctx context.Context, r *service.Request) service.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	if m.Limiter == nil {
		err := merry.New("limiter is nil")
		err = err.WithUserMessage("middleware.ClientThrottler.Limiter must be non-nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.ErrorHandler(ctx, r, err)
	}

	if m.RateLimitHandler == nil {
		m.RateLimitHandler = m.defaultRateLimitHandler
	}

	ip := r.ClientIP()
	ok, cd, err := throttle(m.Limiter, ip, r)

	if err != nil {
		return m.ErrorHandler(ctx, r, err)
	}

	if !ok {
		return m.RateLimitHandler(ctx, r, cd)
	}

	return nil
}

func (m *ClientThrottler) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func (m *ClientThrottler) defaultRateLimitHandler(ctx context.Context, r *service.Request, cd time.Duration) service.Response {
	response := service.NewEmpty(http.StatusTooManyRequests)
	response.Headers().Set(RetryAfterHeaderKey, strconv.Itoa(int(cd/time.Second)))

	return response
}

// UserThrottler is a a rate-limiting middleware that
// throttles requests from a given User, via the
// request Context's Actor) using its Limiter
type UserThrottler struct {
	// Limiter determines whether the request should be throttled
	// based on the context Actor's `ID` method.  This must be
	// non-nil or an InternalServiceError status response will be
	// returned.
	Limiter ratelimit.Provider

	// RateLimitHandler optionally customizes the response for a
	// throttled request. The default handler will return
	// a 429 Too Many Requests response code, an empty body, and
	// the cooldown in seconds in the `Retry-After` header.
	RateLimitHandler RateLimitHandler

	// ErrorHandler optionally customizes the response for an
	// error. The `err` parameter passed to the handler will
	// have a recommended HTTP status code.
	// The default handler will return the recommended status
	// code and an empty body.
	ErrorHandler service.ErrorHandler
}

func (m *UserThrottler) Service(ctx context.Context, r *service.Request) service.Response {
	if m.ErrorHandler == nil {
		m.ErrorHandler = m.defaultErrorHandler
	}

	if m.Limiter == nil {
		err := merry.New("limiter is nil")
		err = err.WithUserMessage("middleware.UserThrottler.Limiter must be non-nil")
		err = err.WithHTTPCode(http.StatusInternalServerError)
		return m.ErrorHandler(ctx, r, err)
	}

	if m.RateLimitHandler == nil {
		m.RateLimitHandler = m.defaultRateLimitHandler
	}

	user := ctx.Actor().ID()
	ok, cd, err := throttle(m.Limiter, user, r)

	if err != nil {
		return m.ErrorHandler(ctx, r, err)
	}

	if !ok {
		return m.RateLimitHandler(ctx, r, cd)
	}

	return nil
}

func (m *UserThrottler) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func (m *UserThrottler) defaultRateLimitHandler(ctx context.Context, r *service.Request, cd time.Duration) service.Response {
	response := service.NewEmpty(http.StatusTooManyRequests)
	response.Headers().Set(RetryAfterHeaderKey, strconv.Itoa(int(cd/time.Second)))

	return response
}

func throttle(limiter ratelimit.Provider, actor string, r *service.Request) (bool, time.Duration, merry.Error) {
	action, path := r.Method, r.URL.Path
	return limiter.Allow(actor, action, path)
}
