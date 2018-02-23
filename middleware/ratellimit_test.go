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

type clientThrottlerCase struct {
	Throttler          *ClientThrottler
	FakeLimiter        *ratelimit.FakeProvider
	ExpectShortCircuit bool
	StatusCode         int
	CoolDown           string
}

func checkClientThrottlerCase(t *testing.T, c clientThrottlerCase) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{
		Request: httpReq,
	}

	res := c.Throttler.Service(ctx, req)

	if res != nil {
		assert.True(t, c.ExpectShortCircuit)
		assert.Equal(t, c.StatusCode, res.StatusCode())
		if c.CoolDown != "" {
			assert.Equal(t, c.CoolDown, res.Headers().Get(RetryAfterHeaderKey))
		}
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
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
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

	ut := &ClientThrottler{
		Limiter: fl,
		RateLimitHandler: func(c context.Context, r *httpx.Request, cd time.Duration) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
		},
	}

	c := clientThrottlerCase{
		Throttler:          ut,
		FakeLimiter:        fl,
		ExpectShortCircuit: true,
		StatusCode:         http.StatusTeapot,
	}

	checkClientThrottlerCase(t, c)
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
		CoolDown:           strconv.Itoa(60 * 60),
	}

	checkClientThrottlerCase(t, c)
}

func TestClientThrottlerServiceNilLimiter(t *testing.T) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{Request: httpReq}

	ut := &ClientThrottler{}
	res := ut.Service(ctx, req)

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
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
	CoolDown           string
}

func checkUserThrottlerCase(t *testing.T, u userThrottlerCase) {
	ctx := context.New(nil).WithActor(&models.FakeUser{
		IDHook: func() string {
			return "5"
		},
	})

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{
		Request: httpReq,
	}

	res := u.Throttler.Service(ctx, req)

	if res != nil {
		assert.True(t, u.ExpectShortCircuit)
		assert.Equal(t, u.StatusCode, res.StatusCode())
		if u.CoolDown != "" {
			assert.Equal(t, u.CoolDown, res.Headers().Get(RetryAfterHeaderKey))
		}
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
		ErrorHandler: func(c context.Context, r *httpx.Request, err merry.Error) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
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
		RateLimitHandler: func(c context.Context, r *httpx.Request, cd time.Duration) httpx.Response {
			return httpx.NewEmpty(http.StatusTeapot)
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
		CoolDown:           strconv.Itoa(60 * 60),
	}

	checkUserThrottlerCase(t, c)
}

func TestUserThrottlerServiceNilLimiter(t *testing.T) {
	ctx := context.New(nil)

	httpReq := httptest.NewRequest(http.MethodPost, "http://10.0.0.1/", nil)
	req := &httpx.Request{Request: httpReq}

	ut := &UserThrottler{}
	res := ut.Service(ctx, req)

	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
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
