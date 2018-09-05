package gateway

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ansel1/merry"
	"github.com/stretchr/testify/assert"

	"github.com/shisa-platform/core/authn"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/middleware"
	"github.com/shisa-platform/core/models"
	"github.com/shisa-platform/core/service"
)

func failingResponse(status int) httpx.Response {
	return &httpx.FakeResponse{
		StatusCodeHook: func() int {
			return status
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

func installService(t *testing.T, g *Gateway, svc *service.Service) {
	if err := g.installServices([]*service.Service{svc}); err != nil {
		t.Fatalf("install services failed: %v", err)
	}
}

func installEndpoints(t *testing.T, g *Gateway, es []service.Endpoint) {
	installService(t, g, newFakeService(es))
}

func newEndpoints(h ...httpx.Handler) []service.Endpoint {
	return []service.Endpoint{service.GetEndpoint(expectedRoute, h...)}
}

func installHandler(t *testing.T, g *Gateway, h httpx.Handler) {
	installEndpoints(t, g, newEndpoints(h))
}

func newEndpointsWithPolicy(p service.Policy, h ...httpx.Handler) []service.Endpoint {
	return []service.Endpoint{service.GetEndpointWithPolicy(expectedRoute, p, h...)}
}

func installHandlersWithPolicy(t *testing.T, g *Gateway, p service.Policy, h ...httpx.Handler) {
	installEndpoints(t, g, newEndpointsWithPolicy(p, h...))
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

type closingResponseWriter struct {
	http.ResponseWriter
	ch               chan bool
	writeCalls       int
	writeHeaderCalls int
}

func (w *closingResponseWriter) Close() {
	close(w.ch)
}

func (w *closingResponseWriter) CloseNotify() <-chan bool {
	return w.ch
}

func (w *closingResponseWriter) Write(bs []byte) (int, error) {
	w.writeCalls++
	return w.ResponseWriter.Write(bs)
}

func (w *closingResponseWriter) WriteHeader(statusCode int) {
	w.writeHeaderCalls++
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *closingResponseWriter) AssertWriteCalled(t *testing.T) {
	assert.NotEqual(t, 0, w.writeCalls, "Write method not called")
}

func (w *closingResponseWriter) AssertWriteNotCalled(t *testing.T) {
	assert.Equalf(t, 0, w.writeCalls, "Write method called %d times", w.writeCalls)
}

func (w *closingResponseWriter) AssertWriteHeaderCalled(t *testing.T) {
	assert.NotEqual(t, 0, w.writeHeaderCalls, "WriteHeader called")
}

func (w *closingResponseWriter) AssertWriteHeaderNotCalled(t *testing.T) {
	assert.Equalf(t, 0, w.writeHeaderCalls, "WriteHeader called %d times", w.writeHeaderCalls)
}

func TestRouterCustomRequestIDGenerator(t *testing.T) {
	expectedRequestID := "zalgo-he-comes"
	var generatorCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *httpx.Request) (string, merry.Error) {
			generatorCalled = true
			return expectedRequestID, nil
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, expectedRequestID, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDGeneratorError(t *testing.T) {
	var generatorCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *httpx.Request) (string, merry.Error) {
			generatorCalled = true
			return "", merry.New("i blewed up!")
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return httpx.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDGeneratorPanic(t *testing.T) {
	var generatorCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *httpx.Request) (string, merry.Error) {
			generatorCalled = true
			panic(merry.New("i blewed up!"))
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return httpx.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDGeneratorEmptyResult(t *testing.T) {
	var generatorCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *httpx.Request) (string, merry.Error) {
			generatorCalled = true
			return "", nil
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return httpx.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterDefaultRequestIDGenerator(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return httpx.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDHeaderKey(t *testing.T) {
	headerKey := "x-zalgo"
	errHook := new(mockErrorHook)
	cut := &Gateway{
		RequestIDHeaderName: headerKey,
		ErrorHook:           errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(headerKey))
}

func TestRouterHandlersPanic(t *testing.T) {
	var handlerCalled bool
	handler := func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(merry.New("i blewed up!"))
	}
	errHook := new(mockErrorHook)
	var completionHookCalled bool
	cut := &Gateway{
		Handlers:  []httpx.Handler{handler},
		ErrorHook: errHook.Handle,
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusInternalServerError, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, completionHookCalled, "completion hook not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersPanicWithErrorHandlerPanic(t *testing.T) {
	var handlerCalled bool
	handler := func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(merry.New("i blewed up!"))
	}
	var errorHandlerCalled bool
	var errorHookCalled bool
	var completionHookCalled bool
	cut := &Gateway{
		Handlers: []httpx.Handler{handler},
		InternalServerErrorHandler: func(context.Context, *httpx.Request, merry.Error) httpx.Response {
			errorHandlerCalled = true
			panic(merry.New("error handler blewed up!"))
		},
		ErrorHook: func(context.Context, *httpx.Request, merry.Error) {
			errorHookCalled = true
			panic(merry.New("error hook blewed up!"))
		},
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusInternalServerError, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, errorHandlerCalled, "error handler not called")
	assert.True(t, errorHookCalled, "error hook not called")
	assert.True(t, completionHookCalled, "completion hook not called")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersAuthentictionNGResponse(t *testing.T) {
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return nil, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHook := new(mockErrorHook)
	var completionHookCalled bool
	cut := &Gateway{
		Handlers:  []httpx.Handler{(&middleware.Authentication{Authenticator: authn}).Service},
		ErrorHook: errHook.Handle,
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusUnauthorized, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, completionHookCalled, "completion hook not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, challenge, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
}

func TestRouterHandlersAuthentictionOKResponse(t *testing.T) {
	user := &models.FakeUser{IDHook: func() string { return "123" }}
	challenge := "Test realm=\"test\""
	authn := &authn.FakeAuthenticator{
		AuthenticateHook: func(context.Context, *httpx.Request) (models.User, merry.Error) {
			return user, nil
		},
		ChallengeHook: func() string {
			return challenge
		},
	}
	errHook := new(mockErrorHook)
	var completionHookCalled bool
	cut := &Gateway{
		Handlers:  []httpx.Handler{(&middleware.Authentication{Authenticator: authn}).Service},
		ErrorHook: errHook.Handle,
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusOK, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		assert.Equal(t, user, ctx.Actor())
		return httpx.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, completionHookCalled, "completion hook not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Empty(t, w.HeaderMap.Get(middleware.WWWAuthenticateHeaderKey))
}

func TestRouterTreeFailure(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	endpoints := []service.Endpoint{
		service.GetEndpoint("/", dummyHandler),
		service.GetEndpoint("/:thing", dummyHandler),
	}
	installEndpoints(t, cut, endpoints)

	cut.tree.children[0].nType = 42

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterBadRoute(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo", nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadRouteCustomHandler(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *httpx.Request) httpx.Response {
			handlerCalled = true
			return httpx.NewEmpty(http.StatusForbidden)
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadRouteCustomHandlerWithPanic(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *httpx.Request) httpx.Response {
			handlerCalled = true
			panic(merry.New("not found handler blewed up!"))
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHeadMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  &service.Pipeline{Handlers: []httpx.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodHead, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterGetMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPutMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.PutEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPostMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.PostEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPatchMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.PatchEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterDeleteMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.DeleteEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterConnectMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route:   expectedRoute,
		Connect: &service.Pipeline{Handlers: []httpx.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodConnect, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterOptionsMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route:   expectedRoute,
		Options: &service.Pipeline{Handlers: []httpx.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterTraceMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route: expectedRoute,
		Trace: &service.Pipeline{Handlers: []httpx.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodTrace, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	svc := newFakeService(newEndpoints(dummyHandler))
	svc.MethodNotAllowedHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusForbidden)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodCustomHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	svc := newFakeService(newEndpoints(dummyHandler))
	svc.MethodNotAllowedHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(merry.New("new allowed handler blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodRedirect(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodPut, route, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodRedirectCustomHandler(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(context.Context, *httpx.Request) httpx.Response {
			handlerCalled = true
			return httpx.NewEmpty(http.StatusForbidden)
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodPut, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodRedirectCustomHandlerPanic(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(context.Context, *httpx.Request) httpx.Response {
			handlerCalled = true
			panic(merry.New("not found handler blewed up!"))
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodPut, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectForbidden(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectForbiddenCustomNotFoundHandler(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(context.Context, *httpx.Request) httpx.Response {
			handlerCalled = true
			return httpx.NewEmpty(http.StatusForbidden)
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirecCustomNotFoundHandlerPanic(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(context.Context, *httpx.Request) httpx.Response {
			handlerCalled = true
			panic("not found handler blewed up!")
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectAllowed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	policy := service.Policy{AllowTrailingSlashRedirects: true}
	installHandlersWithPolicy(t, cut, policy, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, expectedRoute, w.HeaderMap.Get(httpx.LocationHeaderKey))
}

func TestRouterExtraSlashRedirectForbiddenCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	svc := newFakeService(newEndpoints(dummyHandler))
	svc.RedirectHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusForbidden)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to redirect handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectAllowedCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	policy := service.Policy{AllowTrailingSlashRedirects: true}
	svc := newFakeService(newEndpointsWithPolicy(policy, dummyHandler))
	svc.RedirectHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusForbidden)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectAllowedCustomHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	policy := service.Policy{AllowTrailingSlashRedirects: true}
	svc := newFakeService(newEndpointsWithPolicy(policy, dummyHandler))
	svc.RedirectHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(merry.New("redirect handler blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, expectedRoute, w.HeaderMap.Get(httpx.LocationHeaderKey))
}

func TestRouterMissingSlashRedirectForbidden(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	endpoint := service.GetEndpoint(expectedRoute+"/", dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectForbiddenCustomNotFoundHandler(t *testing.T) {
	var handlerCalled bool
	errHook := new(mockErrorHook)
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *httpx.Request) httpx.Response {
			handlerCalled = true
			return httpx.NewEmpty(http.StatusForbidden)
		},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	endpoint := service.GetEndpoint(expectedRoute+"/", dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectAllowed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	route := expectedRoute + "/"

	policy := service.Policy{AllowTrailingSlashRedirects: true}
	endpoint := service.GetEndpointWithPolicy(route, policy, dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, route, w.HeaderMap.Get(httpx.LocationHeaderKey))
}

func TestRouterMissingSlashRedirectAllowedForPutMethod(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	route := expectedRoute + "/"

	policy := service.Policy{AllowTrailingSlashRedirects: true}
	endpoint := service.PutEndpointWithPolicy(route, policy, dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, route, w.HeaderMap.Get(httpx.LocationHeaderKey))
}

func TestRouterMissingSlashRedirectForbiddenCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	endpoint := service.GetEndpoint(expectedRoute+"/", dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.RedirectHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusForbidden)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to redirect handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectAllowedCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	policy := service.Policy{AllowTrailingSlashRedirects: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute+"/", policy, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.RedirectHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusForbidden)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectAllowedCustomHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	policy := service.Policy{AllowTrailingSlashRedirects: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute+"/", policy, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.RedirectHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(merry.New("redirect handler blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPathParamters(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.PathParams, 2)
		assert.Equal(t, "outer", r.PathParams[0].Name)
		assert.Equal(t, "zalgo", r.PathParams[0].Value)
		assert.Equal(t, "inner", r.PathParams[1].Name)
		assert.Equal(t, "he comes", r.PathParams[1].Value)

		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint("/:outer/:inner", handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo/he%20comes", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPathParamtersPreserveEscaping(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.PathParams, 2)
		assert.Equal(t, "outer", r.PathParams[0].Name)
		assert.Equal(t, "zalgo", r.PathParams[0].Value)
		assert.Equal(t, "inner", r.PathParams[1].Name)
		assert.Equal(t, "he%20comes", r.PathParams[1].Value)

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{PreserveEscapedPathParameters: true}
	endpoint := service.GetEndpointWithPolicy("/:outer/:inner", policy, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo/he%20comes", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersForbidMalformed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true

		return httpx.NewEmpty(http.StatusOK)
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?name=foo%zzbar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersForbidMalformedCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var routeHandlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		routeHandlerCalled = true

		return httpx.NewEmpty(http.StatusOK)
	}

	var handlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.MalformedRequestHandler = func(context.Context, *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusForbidden)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?name=foo%zzbar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, routeHandlerCalled, "unexpected call to route handler")
	assert.True(t, handlerCalled, "malformed query handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersForbidMalformedCustomHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var routeHandlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		routeHandlerCalled = true

		return httpx.NewEmpty(http.StatusOK)
	}

	var handlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.MalformedRequestHandler = func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(merry.New("malformed request handler blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?name=foo%zzbar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, routeHandlerCalled, "unexpected call to route handler")
	assert.True(t, handlerCalled, "malformed query handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersAllowMalformed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "bad", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "foo%zzbar", p.Values[0])
				assert.Error(t, p.Err)
				assert.True(t, merry.Is(p.Err, httpx.InvalidParameterValueEscape))
			case 1:
				assert.Equal(t, "good", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "foobar", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusPaymentRequired)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.Policy = service.Policy{AllowMalformedQueryParameters: true}
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "good"}, {Name: "bad"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?bad=foo%zzbar&good=foobar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithoutFields(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he:comes", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithRequiredFieldMissing(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Required: true},
		{Name: "waits"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersRequiredFieldMissingAllowMalformed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "zalgo", p.Name)
				assert.Empty(t, p.Values)
				assert.Error(t, p.Err)
				assert.True(t, merry.Is(p.Err, httpx.MissingQueryParamter))
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{AllowMalformedQueryParameters: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute, policy, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Required: true},
		{Name: "waits"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldMalformedQuery(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he%zzcomes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldMalformedQueryCustomHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}

	var queryHandlerCalled bool
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.MalformedRequestHandler = func(context.Context, *httpx.Request) httpx.Response {
		queryHandlerCalled = true
		return httpx.NewEmpty(http.StatusPaymentRequired)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he%zzcomes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, queryHandlerCalled, "malformed query handler not called")
	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldMalformedQueryAllowMalformed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he%zzcomes", p.Values[0])
				assert.Error(t, p.Err)
				assert.True(t, merry.Is(p.Err, httpx.InvalidParameterValueEscape))
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{AllowMalformedQueryParameters: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute, policy, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he%zzcomes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithRequiredFieldPresent(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he:comes", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Required: true},
		{Name: "waits"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFields(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he:comes", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersFieldValidationFails(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	validator := func(p httpx.QueryParameter) merry.Error {
		if p.Values[0] != "he:comes" {
			return merry.New("can't you feel it?")
		}

		return nil
	}
	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Validator: validator},
		{Name: "waits"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=foobar&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersFieldValidationPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	validator := func(httpx.QueryParameter) merry.Error {
		panic(merry.New("i blewed up!"))
	}
	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Validator: validator},
		{Name: "waits"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=foobar&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersFieldValidationPanicErrorHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	validator := func(httpx.QueryParameter) merry.Error {
		panic(merry.New("i blewed up!"))
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Validator: validator},
		{Name: "waits"},
	}

	var iseHandlerCalled bool
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.InternalServerErrorHandler = func(context.Context, *httpx.Request, merry.Error) httpx.Response {
		iseHandlerCalled = true
		panic(merry.New("i blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=foobar&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	errHook.assertCalledN(t, 2)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersFieldValidationFailsAllowMalformed(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "foobar", p.Values[0])
				assert.Error(t, p.Err)
				assert.True(t, merry.Is(p.Err, httpx.MalformedQueryParamter))
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{AllowMalformedQueryParameters: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute, policy, handler)
	validator := func(p httpx.QueryParameter) merry.Error {
		if p.Values[0] != "he:comes" {
			return merry.New("can't you feel it?")
		}

		return nil
	}
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo", Validator: validator},
		{Name: "waits"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=foobar&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldUnknownParameterForbid(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=foobar&waits=behind%20the%20walls&foo=bar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldUnknownParameterAllow(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 3)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he:comes", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			case 2:
				assert.Equal(t, "foo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "bar", p.Values[0])
				assert.Error(t, p.Err)
				assert.True(t, merry.Is(p.Err, httpx.UnknownQueryParamter))
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{AllowUnknownQueryParameters: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute, policy, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes&waits=behind%20the%20walls&foo=bar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldUnknownInvalidParameterAllow(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 3)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he:comes", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			case 2:
				assert.Equal(t, "x", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "foo%zzbar", p.Values[0])
				assert.Error(t, p.Err)
				assert.True(t, merry.Is(p.Err, httpx.InvalidParameterValueEscape))
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{
		AllowMalformedQueryParameters: true,
		AllowUnknownQueryParameters:   true,
	}
	endpoint := service.GetEndpointWithPolicy(expectedRoute, policy, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "zalgo"}, {Name: "waits"}}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes&waits=behind%20the%20walls&x=foo%zzbar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParametersWithFieldDefault(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		assert.Len(t, r.QueryParams, 2)
		for i, p := range r.QueryParams {
			switch i {
			case 0:
				assert.Equal(t, "zalgo", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "he:comes", p.Values[0])
				assert.NoError(t, p.Err)
			case 1:
				assert.Equal(t, "waits", p.Name)
				assert.Len(t, p.Values, 1)
				assert.Equal(t, "behind the walls", p.Values[0])
				assert.NoError(t, p.Err)
			}
		}

		return httpx.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	endpoint.Get.QuerySchemas = []httpx.ParameterSchema{
		{Name: "zalgo"},
		{Name: "waits", Default: "behind the walls"},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterContextDeadlineSet(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		_, ok := ctx.Deadline()
		assert.True(t, ok, "no deadline set")

		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{TimeBudget: time.Millisecond * 30}
	installHandlersWithPolicy(t, cut, policy, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlerPanic(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(errors.New("i blewed up!"))
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlerPanicCustomISEHandle(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	explosion := errors.New("i blewed up!")
	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(explosion)
	}

	var iseHandlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.InternalServerErrorHandler = func(_ context.Context, _ *httpx.Request, err merry.Error) httpx.Response {
		iseHandlerCalled = true
		assert.True(t, merry.Is(err, explosion))
		return httpx.NewEmpty(http.StatusServiceUnavailable)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlerPanicCustomISEHandleWithPanic(t *testing.T) {
	var errorHookCalled bool
	var completionHookCalled bool
	cut := &Gateway{
		ErrorHook: func(context.Context, *httpx.Request, merry.Error) {
			errorHookCalled = true
			panic(merry.New("error hook blewed up!"))
		},
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusInternalServerError, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	explosion := errors.New("handler blewed up!")
	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		panic(explosion)
	}

	var iseHandlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.InternalServerErrorHandler = func(_ context.Context, _ *httpx.Request, err merry.Error) httpx.Response {
		iseHandlerCalled = true
		assert.True(t, merry.Is(err, explosion))

		panic(merry.New("ise handler blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	assert.True(t, errorHookCalled, "error hook not called")
	assert.True(t, completionHookCalled, "completion hook not called")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersNoResult(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		return nil
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersNoResultCustomerErrorHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		return nil
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersNoResultCustomISEHandler(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		return nil
	}

	var iseHandlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.InternalServerErrorHandler = func(ctx context.Context, r *httpx.Request, err merry.Error) httpx.Response {
		iseHandlerCalled = true
		return httpx.NewEmpty(http.StatusServiceUnavailable)
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersNoResultCustomISEHandlerWithPanic(t *testing.T) {
	var errorHookCalled bool
	var completionHookCalled bool
	cut := &Gateway{
		ErrorHook: func(context.Context, *httpx.Request, merry.Error) {
			errorHookCalled = true
			panic(merry.New("error hook blewed up!"))
		},
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusInternalServerError, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		return nil
	}

	var iseHandlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.InternalServerErrorHandler = func(ctx context.Context, r *httpx.Request, err merry.Error) httpx.Response {
		iseHandlerCalled = true
		panic(merry.New("ise handler blewed up!"))
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	assert.True(t, errorHookCalled, "error hook not called")
	assert.True(t, completionHookCalled, "completion hook not called")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMultipleHandlersEarlyExit(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handler1Called bool
	handler1 := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handler1Called = true

		return httpx.NewEmpty(http.StatusPaymentRequired)
	}
	var handler2Called bool
	handler2 := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handler2Called = true

		return nil
	}
	var handler3Called bool
	handler3 := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handler3Called = true

		return httpx.NewEmpty(http.StatusOK)
	}

	installEndpoints(t, cut, newEndpoints(handler1, handler2, handler3))
	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handler1Called, "handler one not called")
	assert.False(t, handler2Called, "unexpected handler two call")
	assert.False(t, handler3Called, "unexpected handler three call")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterResponseHeaders(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		response := httpx.NewEmpty(http.StatusOK)
		response.Headers().Add("x-zalgo", "he comes")
		return response
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, "he comes", w.HeaderMap.Get("x-zalgo"))
}

func TestRouterResponseTrailers(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		response := httpx.NewEmpty(http.StatusOK)
		response.Trailers().Add("x-zalgo", "he comes")
		return response
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, "he comes", w.HeaderMap.Get("x-zalgo"))
}

func TestRouterSerializationError(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		response := failingResponse(http.StatusOK)
		return response
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouter(t *testing.T) {
	errHook := new(mockErrorHook)
	var completionHookCalled bool
	cut := &Gateway{
		ErrorHook: errHook.Handle,
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusOK, s.StatusCode)
			assert.Equal(t, 0, s.Size)
		},
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, completionHookCalled, "completion hook not called")
	errHook.assertNotCalled(t)
	assert.True(t, w.Flushed)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterNoCompletionHook(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()
	assert.Nil(t, cut.CompletionHook)

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	errHook.assertNotCalled(t)
	assert.True(t, w.Flushed)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterPanicCompletionHook(t *testing.T) {
	errHook := new(mockErrorHook)
	var completionHookCalled bool
	cut := &Gateway{
		ErrorHook: errHook.Handle,
		CompletionHook: func(_ context.Context, _ *httpx.Request, s httpx.ResponseSnapshot) {
			completionHookCalled = true
			assert.Equal(t, http.StatusOK, s.StatusCode)
			assert.Equal(t, 0, s.Size)
			panic(merry.New("completion hook blewed up!"))
		},
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, completionHookCalled, "completion hook not called")
	errHook.assertCalledN(t, 1)
	assert.True(t, w.Flushed)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterPipelineHandlerResponseWithError(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		err := merry.New("lol wut")
		return httpx.NewEmptyError(http.StatusServiceUnavailable, err)
	}

	installEndpoints(t, cut, newEndpoints(handler))
	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterEndpointHandlerTimeout(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		time.Sleep(time.Second * 2)
		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{TimeBudget: time.Millisecond * 30}
	installHandlersWithPolicy(t, cut, policy, handler)
	w := httptest.NewRecorder()

	start := time.Now().UTC()
	cut.ServeHTTP(w, fakeRequest)
	end := time.Now().UTC()

	assert.WithinDuration(t, start, end, time.Millisecond*100)
	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterSlowGatewayHandlerDisruptsEndpointPipeline(t *testing.T) {
	var gatewayHandlerCalled bool
	gatewayHandler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		gatewayHandlerCalled = true
		time.Sleep(time.Millisecond * 500)
		return nil
	}

	errHook := new(mockErrorHook)
	cut := &Gateway{
		Handlers:  []httpx.Handler{gatewayHandler},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	policy := service.Policy{TimeBudget: time.Millisecond * 30}

	handler1 := func(ctx context.Context, r *httpx.Request) httpx.Response {
		return nil
	}
	handler2 := func(ctx context.Context, r *httpx.Request) httpx.Response {
		return httpx.NewEmpty(http.StatusOK)
	}

	installHandlersWithPolicy(t, cut, policy, handler1, handler2)
	w := httptest.NewRecorder()

	start := time.Now().UTC()
	cut.ServeHTTP(w, fakeRequest)
	end := time.Now().UTC()

	assert.WithinDuration(t, start, end, time.Millisecond*1250)
	assert.True(t, gatewayHandlerCalled, "gateway handler not called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterSlowGatewayHandlerWithTimeout(t *testing.T) {
	var gatewayHandlerCalled bool
	gatewayHandler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		gatewayHandlerCalled = true
		time.Sleep(time.Second * 1)
		return nil
	}

	errHook := new(mockErrorHook)
	cut := &Gateway{
		Handlers:        []httpx.Handler{gatewayHandler},
		HandlersTimeout: time.Millisecond * 500,
		ErrorHook:       errHook.Handle,
	}
	cut.init()

	policy := service.Policy{TimeBudget: time.Millisecond * 30}

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		return httpx.NewEmpty(http.StatusOK)
	}

	installHandlersWithPolicy(t, cut, policy, handler)
	w := httptest.NewRecorder()

	start := time.Now().UTC()
	cut.ServeHTTP(w, fakeRequest)
	end := time.Now().UTC()

	assert.WithinDuration(t, start, end, time.Millisecond*600)
	assert.True(t, gatewayHandlerCalled, "gateway handler not called")
	assert.False(t, handlerCalled, "handler called unexpectedly")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterEndpointHandlerClientDisconnect(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true
		time.Sleep(time.Second * 2)
		return httpx.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{TimeBudget: 5 * time.Second}
	installHandlersWithPolicy(t, cut, policy, handler)
	recorder := httptest.NewRecorder()
	w := &closingResponseWriter{
		ResponseWriter: recorder,
		ch:             make(chan bool),
	}
	timer := time.AfterFunc(250*time.Millisecond, func() { w.Close() })
	defer timer.Stop()

	start := time.Now().UTC()
	cut.ServeHTTP(w, fakeRequest)
	end := time.Now().UTC()

	assert.WithinDuration(t, start, end, time.Millisecond*500)
	assert.True(t, handlerCalled, "handler not called")
	errHook.assertCalledN(t, 2)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 0, recorder.Body.Len())
	assert.False(t, recorder.Flushed)
	w.AssertWriteHeaderNotCalled(t)
	w.AssertWriteNotCalled(t)
}

func TestRouterEndpointHandlerCloseNotifierGoroutineNormalExit(t *testing.T) {
	errHook := new(mockErrorHook)
	cut := &Gateway{
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)
	recorder := httptest.NewRecorder()
	w := &closingResponseWriter{
		ResponseWriter: recorder,
		ch:             make(chan bool),
	}

	cut.ServeHTTP(w, fakeRequest)

	errHook.assertNotCalled(t)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 0, recorder.Body.Len())
	assert.False(t, recorder.Flushed)
	w.AssertWriteHeaderCalled(t)
	w.AssertWriteNotCalled(t)
}

func TestRouterGatewayHandlerClientDisconnect(t *testing.T) {
	var gatewayHandlerCalled bool
	gatewayHandler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		gatewayHandlerCalled = true
		time.Sleep(time.Second * 2)
		return httpx.NewEmpty(http.StatusOK)
	}

	errHook := new(mockErrorHook)
	cut := &Gateway{
		Handlers:  []httpx.Handler{gatewayHandler},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)
	recorder := httptest.NewRecorder()
	w := &closingResponseWriter{
		ResponseWriter: recorder,
		ch:             make(chan bool),
	}
	timer := time.AfterFunc(500*time.Millisecond, func() { w.Close() })
	defer timer.Stop()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, gatewayHandlerCalled, "gateway handler not called")
	errHook.assertCalledN(t, 2)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 0, recorder.Body.Len())
	assert.False(t, recorder.Flushed)
	w.AssertWriteHeaderNotCalled(t)
	w.AssertWriteNotCalled(t)
}

func TestGatewayHandlerResponseWithError(t *testing.T) {
	var gwHandlerCalled bool
	gwHandler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		gwHandlerCalled = true

		err := merry.New("lol wut")
		return httpx.NewEmptyError(http.StatusServiceUnavailable, err)
	}
	errHook := new(mockErrorHook)
	cut := &Gateway{
		Handlers:  []httpx.Handler{gwHandler},
		ErrorHook: errHook.Handle,
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *httpx.Request) httpx.Response {
		handlerCalled = true

		return httpx.NewEmpty(http.StatusOK)
	}

	installEndpoints(t, cut, newEndpoints(handler))
	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, gwHandlerCalled, "handler not called")
	assert.False(t, handlerCalled, "handler called")
	errHook.assertCalledN(t, 1)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}
