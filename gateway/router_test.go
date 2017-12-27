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
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

var (
	expectedRoute = "/test"
	fakeRequest   = httptest.NewRequest(http.MethodGet, expectedRoute, nil)
)

type failingResponse struct {
	code int
}

func (r failingResponse) StatusCode() int {
	return r.code
}

func (r failingResponse) Headers() http.Header {
	return nil
}

func (r failingResponse) Trailers() http.Header {
	return nil
}

func (r failingResponse) Serialize(io.Writer) (int, error) {
	return 0, errors.New("i blewed up!")
}

func dummyHandler(context.Context, *service.Request) service.Response {
	return service.NewEmpty(http.StatusOK)
}

func newFakeService(es []service.Endpoint) *service.FakeService {
	return &service.FakeService{
		NameHook:                       func() string { return "test" },
		EndpointsHook:                  func() []service.Endpoint { return es },
		HandlersHook:                   func() []service.Handler { return nil },
		MalformedRequestHandlerHook:    func() service.Handler { return nil },
		MethodNotAllowedHandlerHook:    func() service.Handler { return nil },
		RedirectHandlerHook:            func() service.Handler { return nil },
		InternalServerErrorHandlerHook: func() service.ErrorHandler { return nil },
	}
}

func installService(t *testing.T, g *Gateway, svc service.Service) {
	if err := g.installServices([]service.Service{svc}); err != nil {
		t.Fatalf("install services failed: %v", err)
	}
}

func installEndpoints(t *testing.T, g *Gateway, es []service.Endpoint) {
	installService(t, g, newFakeService(es))
}

func newEndpoints(h ...service.Handler) []service.Endpoint {
	return []service.Endpoint{service.GetEndpoint(expectedRoute, h...)}
}

func installHandler(t *testing.T, g *Gateway, h service.Handler) {
	installEndpoints(t, g, newEndpoints(h))
}

func newEndpointsWithPolicy(h service.Handler, p service.Policy) []service.Endpoint {
	return []service.Endpoint{service.GetEndpointWithPolicy(expectedRoute, p, h)}
}

func installHandlerWithPolicy(t *testing.T, g *Gateway, h service.Handler, p service.Policy) {
	installEndpoints(t, g, newEndpointsWithPolicy(h, p))
}

func TestRouterCustomRequestIDGenerator(t *testing.T) {
	expectedRequestID := "zalgo-he-comes"
	var generatorCalled bool
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *service.Request) (string, merry.Error) {
			generatorCalled = true
			return expectedRequestID, nil
		},
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, expectedRequestID, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDGeneratorError(t *testing.T) {
	var generatorCalled bool
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *service.Request) (string, merry.Error) {
			generatorCalled = true
			return "", merry.New("i blewed up!")
		},
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return service.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDGeneratorEmptyResult(t *testing.T) {
	var generatorCalled bool
	cut := &Gateway{
		RequestIDGenerator: func(context.Context, *service.Request) (string, merry.Error) {
			generatorCalled = true
			return "", nil
		},
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return service.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, generatorCalled, "custom request id generator not called")
	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterDefaultRequestIDGenerator(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.NotEmpty(t, ctx.RequestID())
		return service.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterCustomRequestIDHeaderKey(t *testing.T) {
	headerKey := "x-zalgo"
	cut := &Gateway{
		RequestIDHeaderName: headerKey,
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(headerKey))
}

func TestRouterTreeFailure(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	endpoints := []service.Endpoint{
		service.GetEndpoint("/", dummyHandler),
		service.GetEndpoint("/:thing", dummyHandler),
	}
	installEndpoints(t, cut, endpoints)

	cut.tree.children[0].nType = 42

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}

func TestRouterBadRoute(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo", nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadRouteCustomHandler(t *testing.T) {
	var handlerCalled bool
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		},
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHeadMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route: expectedRoute,
		Head:  &service.Pipeline{Handlers: []service.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodHead, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterGetMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPutMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.PutEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPostMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.PostEndpoint(expectedRoute, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPatchMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
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
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
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
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route:   expectedRoute,
		Connect: &service.Pipeline{Handlers: []service.Handler{handler}},
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
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route:   expectedRoute,
		Options: &service.Pipeline{Handlers: []service.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterTraceMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.Endpoint{
		Route: expectedRoute,
		Trace: &service.Pipeline{Handlers: []service.Handler{handler}},
	}
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodTrace, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethod(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodCustomHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	svc := newFakeService(newEndpoints(dummyHandler))
	var handlerCalled bool
	svc.MethodNotAllowedHandlerHook = func() service.Handler {
		return func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodRedirect(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodPut, route, nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterBadMethodRedirectCustomHandler(t *testing.T) {
	var handlerCalled bool
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		},
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodPut, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectForbidden(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectForbiddenCustomNotFoundHandler(t *testing.T) {
	var handlerCalled bool
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		},
	}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectAllowed(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	policy := service.Policy{AllowTrailingSlashRedirects: true}
	installHandlerWithPolicy(t, cut, dummyHandler, policy)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, expectedRoute, w.HeaderMap.Get(service.LocationHeaderKey))
}

func TestRouterExtraSlashRedirectForbiddenCustomHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	svc := newFakeService(newEndpoints(dummyHandler))
	svc.RedirectHandlerHook = func() service.Handler {
		return func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to redirect handler")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterExtraSlashRedirectAllowedCustomHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	policy := service.Policy{AllowTrailingSlashRedirects: true}
	svc := newFakeService(newEndpointsWithPolicy(dummyHandler, policy))
	svc.RedirectHandlerHook = func() service.Handler {
		return func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	route := expectedRoute + "/"
	request := httptest.NewRequest(http.MethodGet, route, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectForbidden(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	endpoint := service.GetEndpoint(expectedRoute+"/", dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectForbiddenCustomNotFoundHandler(t *testing.T) {
	var handlerCalled bool
	cut := &Gateway{
		NotFoundHandler: func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		},
	}
	cut.init()

	endpoint := service.GetEndpoint(expectedRoute+"/", dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectAllowed(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	route := expectedRoute + "/"

	policy := service.Policy{AllowTrailingSlashRedirects: true}
	endpoint := service.GetEndpointWithPolicy(route, policy, dummyHandler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, route, w.HeaderMap.Get(service.LocationHeaderKey))
}

func TestRouterMissingSlashRedirectForbiddenCustomHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	endpoint := service.GetEndpoint(expectedRoute+"/", dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.RedirectHandlerHook = func() service.Handler {
		return func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to redirect handler")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMissingSlashRedirectAllowedCustomHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	policy := service.Policy{AllowTrailingSlashRedirects: true}
	endpoint := service.GetEndpointWithPolicy(expectedRoute+"/", policy, dummyHandler)
	svc := newFakeService([]service.Endpoint{endpoint})
	svc.RedirectHandlerHook = func() service.Handler {
		return func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, expectedRoute, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "not found handler not called")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPathParamters(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.Len(t, r.PathParams, 2)
		assert.Equal(t, "outer", r.PathParams[0].Key)
		assert.Equal(t, "zalgo", r.PathParams[0].Value)
		assert.Equal(t, "inner", r.PathParams[1].Key)
		assert.Equal(t, "he comes", r.PathParams[1].Value)

		return service.NewEmpty(http.StatusOK)
	}

	endpoint := service.GetEndpoint("/:outer/:inner", handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo/he%20comes", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterPathParamtersPreserveEscaping(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.Len(t, r.PathParams, 2)
		assert.Equal(t, "outer", r.PathParams[0].Key)
		assert.Equal(t, "zalgo", r.PathParams[0].Value)
		assert.Equal(t, "inner", r.PathParams[1].Key)
		assert.Equal(t, "he%20comes", r.PathParams[1].Value)

		return service.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{PreserveEscapedPathParameters: true}
	endpoint := service.GetEndpointWithPolicy("/:outer/:inner", policy, handler)
	installEndpoints(t, cut, []service.Endpoint{endpoint})

	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/zalgo/he%20comes", nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParamters(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.NotEmpty(t, r.QueryParams)
		assert.Equal(t, "he:comes", r.QueryParams.Get("zalgo"))
		assert.Equal(t, "behind the walls", r.QueryParams.Get("waits"))

		return service.NewEmpty(http.StatusOK)
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?zalgo=he:comes&waits=behind%20the%20walls"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParamtersForbidMalformed(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true

		return service.NewEmpty(http.StatusOK)
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?name=foo%zzbar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, handlerCalled, "unexpected call to handler")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParamtersForbidMalformedCustomHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var routeHandlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		routeHandlerCalled = true

		return service.NewEmpty(http.StatusOK)
	}

	var handlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.MalformedRequestHandlerHook = func() service.Handler {
		return func(ctx context.Context, r *service.Request) service.Response {
			handlerCalled = true
			return service.NewEmpty(http.StatusForbidden)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?name=foo%zzbar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.False(t, routeHandlerCalled, "unexpected call to route handler")
	assert.True(t, handlerCalled, "malformed query handler not called")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterQueryParamtersAllowMalformed(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		assert.NotEmpty(t, r.QueryParams)
		assert.Equal(t, "foobar", r.QueryParams.Get("good"))
		assert.Empty(t, r.QueryParams.Get("bad"))

		return service.NewEmpty(http.StatusPaymentRequired)
	}

	policy := service.Policy{AllowMalformedQueryParameters: true}
	installHandlerWithPolicy(t, cut, handler, policy)

	w := httptest.NewRecorder()
	uri := expectedRoute + "?bad=foo%zzbar&good=foobar"
	request := httptest.NewRequest(http.MethodGet, uri, nil)
	cut.ServeHTTP(w, request)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterContextDeadlineSet(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		_, ok := ctx.Deadline()
		assert.True(t, ok, "no deadline set")

		return service.NewEmpty(http.StatusOK)
	}

	policy := service.Policy{TimeBudget: 30 * time.Millisecond}
	installHandlerWithPolicy(t, cut, handler, policy)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlerPanic(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		panic(errors.New("i blewed up!"))
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlerPanicNonError(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		panic("i blewed up!")
	}

	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlerPanicCustomISEHandle(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	explosion := errors.New("i blewed up!")
	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true
		panic(explosion)
	}

	var iseHandlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.InternalServerErrorHandlerHook = func() service.ErrorHandler {
		return func(ctx context.Context, r *service.Request, err merry.Error) service.Response {
			iseHandlerCalled = true
			assert.True(t, merry.Is(err, explosion))
			return service.NewEmpty(http.StatusServiceUnavailable)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterHandlersNoResult(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
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

func TestRouterHandlersNoResultCustomISEHandler(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true

		return nil
	}

	var iseHandlerCalled bool
	svc := newFakeService(newEndpoints(handler))
	svc.InternalServerErrorHandlerHook = func() service.ErrorHandler {
		return func(ctx context.Context, r *service.Request, err merry.Error) service.Response {
			iseHandlerCalled = true
			return service.NewEmpty(http.StatusServiceUnavailable)
		}
	}
	installService(t, cut, svc)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.True(t, iseHandlerCalled, "ISE handler not called")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterMultipleHandlersEarlyExit(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handler1Called bool
	handler1 := func(ctx context.Context, r *service.Request) service.Response {
		handler1Called = true

		return service.NewEmpty(http.StatusPaymentRequired)
	}
	var handler2Called bool
	handler2 := func(ctx context.Context, r *service.Request) service.Response {
		handler2Called = true

		return nil
	}
	var handler3Called bool
	handler3 := func(ctx context.Context, r *service.Request) service.Response {
		handler3Called = true

		return service.NewEmpty(http.StatusOK)
	}

	installEndpoints(t, cut, newEndpoints(handler1, handler2, handler3))
	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handler1Called, "handler one not called")
	assert.False(t, handler2Called, "unexpected handler two call")
	assert.False(t, handler3Called, "unexpected handler three call")
	assert.Equal(t, http.StatusPaymentRequired, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterResponseHeaders(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true

		response := service.NewEmpty(http.StatusOK)
		response.Headers().Add("zalgo", "he comes")
		return response
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, "he comes", w.HeaderMap.Get("zalgo"))
}

func TestRouterResponseTrailers(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true

		response := service.NewEmpty(http.StatusOK)
		response.Trailers().Add("zalgo", "he comes")
		return response
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, "he comes", w.HeaderMap.Get("zalgo"))
}

func TestRouterLoggingEnabled(t *testing.T) {
	cut := &Gateway{
		Logger: zap.NewExample(),
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true

		user := &models.FakeUser{IDHook: func() string { return "123" }}
		ctx = ctx.WithActor(user)
		return service.NewEmpty(http.StatusOK)
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouterSerializationError(t *testing.T) {
	cut := &Gateway{
		Logger: zap.NewExample(),
	}
	cut.init()

	var handlerCalled bool
	handler := func(ctx context.Context, r *service.Request) service.Response {
		handlerCalled = true

		response := failingResponse{code: http.StatusOK}
		return response
	}
	installHandler(t, cut, handler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, handlerCalled, "handler not called")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestRouter(t *testing.T) {
	cut := &Gateway{}
	cut.init()

	installHandler(t, cut, dummyHandler)

	w := httptest.NewRecorder()
	cut.ServeHTTP(w, fakeRequest)

	assert.True(t, w.Flushed)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, w.Body.Len())
	assert.NotEmpty(t, w.HeaderMap.Get(cut.RequestIDHeaderName))
}
