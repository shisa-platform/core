package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/ratelimit"
	"github.com/percolate/shisa/service"
	"time"
)

// ClientThrottler is a a rate-limiting middleware that
// throttles requests from a given ClientIP using its Limiter
type ClientThrottler struct {
	// Limiter is a ratelimit.Provider that determines whether
	// the request should be throttled, based on the request's ClientIP
	Limiter ratelimit.Provider

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

	ip := r.ClientIP()
	ok, cd, err := throttle(m.Limiter, ip, r)

	if !ok && err == nil {
		err = merry.New("request throttled, rate limit exceeded").WithHTTPCode(http.StatusTooManyRequests)
		err = err.WithValue("cooldown", cd).WithValue("clientip", ip)
		err = err.WithValue("method", r.Method).WithValue("path", r.URL.Path)
	}

	if err != nil {
		return m.ErrorHandler(ctx, r, err)
	}

	return nil
}

func (m *ClientThrottler) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

// UserThrottler is a a rate-limiting middleware that
// throttles requests from a given User, via the
// request Context's Actor) using its Limiter
type UserThrottler struct {
	// Limiter is a ratelimit.Provider that determines whether
	// the request should be throttled, based on the request
	// context Actor's ID method
	Limiter ratelimit.Provider

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

	user := ctx.Actor().ID()
	ok, cd, err := throttle(m.Limiter, user, r)

	if !ok && err == nil {
		err = merry.New("request throttled, rate limit exceeded").WithHTTPCode(http.StatusTooManyRequests)
		err = err.WithValue("cooldown", cd).WithValue("userid", user)
		err = err.WithValue("method", r.Method).WithValue("path", r.URL.Path)
	}

	if err != nil {
		return m.ErrorHandler(ctx, r, err)
	}

	return nil
}

func (m *UserThrottler) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func throttle(limiter ratelimit.Provider, actor string, r *service.Request) (ok bool, cd time.Duration, err merry.Error) {
	action, path := r.Method, r.URL.Path
	ok, cd, err = limiter.Allow(actor, action, path)
	return
}
