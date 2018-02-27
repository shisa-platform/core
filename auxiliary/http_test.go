package auxiliary

import (
	stdctx "context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

func unserializableResponse() httpx.Response {
	return &httpx.FakeResponse{
		StatusCodeHook: func() int {
			return http.StatusInternalServerError
		},
		HeadersHook: func() http.Header {
			return nil
		},
		TrailersHook: func() http.Header {
			return nil
		},
		ErrHook: func() error {
			return nil
		},
		SerializeHook: func(io.Writer) merry.Error {
			return merry.New("i blewed up!")
		},
	}
}

type stubAuthorizer struct {
	ok  bool
	err merry.Error
}

func (a stubAuthorizer) Authorize(context.Context, *httpx.Request) (bool, merry.Error) {
	return a.ok, a.err
}

type mockErrorHook struct {
	calls int
}

func (m *mockErrorHook) Handle(context.Context, *httpx.Request, merry.Error) {
	m.calls++
}

func (m *mockErrorHook) assertNotCalled(t *testing.T) {
	t.Helper()
	assert.Equal(t, 0, m.calls, "unexpected error handler calls")
}

func (m *mockErrorHook) assertCalled(t *testing.T) {
	t.Helper()
	assert.NotEqual(t, 0, m.calls, "error handler not called")
}

func (m *mockErrorHook) assertCalledN(t *testing.T, expected int) {
	t.Helper()
	assert.Equalf(t, expected, m.calls, "error handler called %d times, expected %d", m.calls, expected)
}

func TestHTTPServerRouterDefault(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &HTTPServer{
		ErrorHook: errHook.Handle,
	}
	ctx := context.New(stdctx.Background())
	fakeRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	request := &httpx.Request{Request: fakeRequest}

	response := cut.route(ctx, request)

	errHook.assertCalledN(t, 1)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.StatusCode())
}

func TestHTTPServerRouterPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &HTTPServer{
		Router: func(context.Context, *httpx.Request) httpx.Handler {
			panic(merry.New("i blewed up!"))
		},
		ErrorHook: errHook.Handle,
	}
	ctx := context.New(stdctx.Background())
	fakeRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	request := &httpx.Request{Request: fakeRequest}

	response := cut.route(ctx, request)

	errHook.assertCalledN(t, 1)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestHTTPServerRouterPanicString(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &HTTPServer{
		Router: func(context.Context, *httpx.Request) httpx.Handler {
			panic("i blewed up!")
		},
		ErrorHook: errHook.Handle,
	}
	ctx := context.New(stdctx.Background())
	fakeRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	request := &httpx.Request{Request: fakeRequest}

	response := cut.route(ctx, request)

	errHook.assertCalledN(t, 1)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestHTTPServerRouterHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &HTTPServer{
		Router: func(context.Context, *httpx.Request) httpx.Handler {
			return func(context.Context, *httpx.Request) httpx.Response {
				panic(merry.New("i blewed up!"))
			}
		},
		ErrorHook: errHook.Handle,
	}
	ctx := context.New(stdctx.Background())
	fakeRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	request := &httpx.Request{Request: fakeRequest}

	response := cut.route(ctx, request)

	errHook.assertCalledN(t, 1)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestHTTPServerRouterHandlerPanicString(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &HTTPServer{
		Router: func(context.Context, *httpx.Request) httpx.Handler {
			return func(context.Context, *httpx.Request) httpx.Response {
				panic("i blewed up!")
			}
		},
		ErrorHook: errHook.Handle,
	}
	ctx := context.New(stdctx.Background())
	fakeRequest := httptest.NewRequest(http.MethodGet, "/test", nil)
	request := &httpx.Request{Request: fakeRequest}

	response := cut.route(ctx, request)

	errHook.assertCalledN(t, 1)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode())
}

func TestHTTPServerRequestIDGeneratorPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := HTTPServer{
		RequestIDGenerator: func(c context.Context, r *httpx.Request) (string, merry.Error) {
			panic(merry.New("i blewed up!"))
		},
		Router: func(context.Context, *httpx.Request) httpx.Handler {
			return func(context.Context, *httpx.Request) httpx.Response {
				return httpx.NewEmpty(http.StatusOK)
			}
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}

func TestHTTPServerCompletionHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	var hookCalled bool
	cut := HTTPServer{
		Router: func(context.Context, *httpx.Request) httpx.Handler {
			return func(context.Context, *httpx.Request) httpx.Response {
				return httpx.NewEmpty(http.StatusOK)
			}
		},
		ErrorHook: errHook.Handle,
		CompletionHook: func(context.Context, *httpx.Request, httpx.ResponseSnapshot) {
			hookCalled = true
			panic("i blewed up!")
		},
	}
	cut.init()

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	cut.ServeHTTP(w, r)

	errHook.assertCalledN(t, 1)
	assert.True(t, hookCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
	assert.True(t, w.Flushed)
}
