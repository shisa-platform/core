package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/ratelimit"
	"github.com/percolate/shisa/service"
	"time"
)

type clientThrottlerCase struct {
	Throttler          *ClientThrottler
	FakeLimiter        *ratelimit.FakeProvider
	ExpectShortCircuit bool
	StatusCode         int
}

func checkClientThrottlerCase(t *testing.T, c clientThrottlerCase) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &service.Request{
		Request: httpReq,
	}

	res := c.Throttler.Service(ctx, req)

	if res != nil {
		assert.True(t, c.ExpectShortCircuit)
		assert.Equal(t, c.StatusCode, res.StatusCode())
	} else {
		assert.False(t, c.ExpectShortCircuit)
	}

	c.FakeLimiter.AssertAllowCalledOnceWith(t, req.ClientIP(), req.Method, req.URL.Path)
}

func TestClientThrottlerServiceErrorHandlerHook(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("found a teapot")
		},
	}

	ct := &ClientThrottler{
		Limiter: fl,
		ErrorHandler: func(c context.Context, r *service.Request, err merry.Error) service.Response {
			return service.NewEmpty(http.StatusTeapot)
		},
	}

	c := clientThrottlerCase{
		Throttler:          ct,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTeapot,
	}

	checkClientThrottlerCase(t, c)
}

func TestClientThrottlerServiceRateLimitHandlerHook(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	ut := &UserThrottler{
		Limiter: fl,
		RateLimitHandler: func(c context.Context, r *service.Request, cd time.Duration) service.Response {
			return service.NewEmpty(http.StatusTeapot)
		},
	}

	c := userThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTeapot,
	}

	checkUserThrottlerCase(t, c)
}

func TestClientThrottlerServiceAllowError(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("something broke")
		},
	}

	ct := &ClientThrottler{
		Limiter: fl,
	}

	c := clientThrottlerCase{
		Throttler:          ct,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusInternalServerError,
	}

	checkClientThrottlerCase(t, c)
}

func TestClientThrottlerServiceThrottled(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	ct := &ClientThrottler{
		Limiter: fl,
	}

	c := clientThrottlerCase{
		Throttler:          ct,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTooManyRequests,
	}

	checkClientThrottlerCase(t, c)
}

func TestClientThrottlerServicePass(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	ct := &ClientThrottler{
		Limiter: fl,
	}

	c := clientThrottlerCase{
		Throttler:          ct,
		FakeLimiter:        fl,
		ExpectShortCircuit: false,
	}

	checkClientThrottlerCase(t, c)
}

type userThrottlerCase struct {
	Throttler          *UserThrottler
	FakeLimiter        *ratelimit.FakeProvider
	ExpectShortCircuit bool
	StatusCode         int
}

func checkUserThrottlerCase(t *testing.T, u userThrottlerCase) {
	ctx := context.New(nil).WithActor(&models.FakeUser{
		IDHook: func() string {
			return "5"
		},
	})

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &service.Request{
		Request: httpReq,
	}

	res := u.Throttler.Service(ctx, req)

	if res != nil {
		assert.True(t, u.ExpectShortCircuit)
		assert.Equal(t, u.StatusCode, res.StatusCode())
	} else {
		assert.False(t, u.ExpectShortCircuit)
	}

	u.FakeLimiter.AssertAllowCalledOnceWith(t, ctx.Actor().ID(), req.Method, req.URL.Path)
}

func TestUserThrottlerServiceErrorHandlerHook(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("found a teapot")
		},
	}

	ut := &UserThrottler{
		Limiter: fl,
		ErrorHandler: func(c context.Context, r *service.Request, err merry.Error) service.Response {
			return service.NewEmpty(http.StatusTeapot)
		},
	}

	c := userThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTeapot,
	}

	checkUserThrottlerCase(t, c)
}

func TestUserThrottlerServiceRateLimitHandlerHook(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	ut := &UserThrottler{
		Limiter: fl,
		RateLimitHandler: func(c context.Context, r *service.Request, cd time.Duration) service.Response {
			return service.NewEmpty(http.StatusTeapot)
		},
	}

	c := userThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTeapot,
	}

	checkUserThrottlerCase(t, c)
}

func TestUserThrottlerServiceAllowError(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, 0, merry.New("something broke")
		},
	}

	ut := &UserThrottler{
		Limiter: fl,
	}

	c := userThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusInternalServerError,
	}

	checkUserThrottlerCase(t, c)
}

func TestUserThrottlerServiceThrottled(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return false, time.Hour, nil
		},
	}

	ut := &UserThrottler{
		Limiter: fl,
	}

	c := userThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTooManyRequests,
	}

	checkUserThrottlerCase(t, c)
}

func TestUserThrottlerServicePass(t *testing.T) {
	fl := &ratelimit.FakeProvider{
		AllowHook: func(actor, action, path string) (bool, time.Duration, merry.Error) {
			return true, 0, nil
		},
	}

	ut := &UserThrottler{
		Limiter: fl,
	}

	c := userThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: false,
	}

	checkUserThrottlerCase(t, c)
}
