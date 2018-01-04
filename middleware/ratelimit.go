package middleware

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/ratelimit"
	"github.com/percolate/shisa/service"
)

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
	err := throttle(m.Limiter, ip, r)
	if err != nil {
		return m.ErrorHandler(ctx, r, err)
	}
	return nil
}

func (m *ClientThrottler) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

type UserThrottler struct {
	// Limiter is a ratelimit.Provider that determines whether
	// the request should be throttled, based on the request
	// context Actor's ID receiver
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

	actor := ctx.Actor().ID()
	err := throttle(m.Limiter, actor, r)
	if err != nil {
		return m.ErrorHandler(ctx, r, err)
	}
	return nil
}

func (m *UserThrottler) defaultErrorHandler(ctx context.Context, r *service.Request, err merry.Error) service.Response {
	return service.NewEmpty(merry.HTTPCode(err))
}

func throttle(limiter ratelimit.Provider, actor string, r *service.Request) merry.Error {
	action, path := r.Method, r.URL.Path
	ok, err := limiter.Allow(actor, action, path)

	if err != nil {
		return err
	} else if !ok {
		return merry.Errorf("throttle %q to %q on %q", actor, action, path).WithHTTPCode(http.StatusTooManyRequests)
	}
	return nil
}
