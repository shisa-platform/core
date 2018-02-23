package gateway

import (
	"net/http"
	"net/http/httptest"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

var (
	expectedRoute = "/test"
	fakeRequest   = httptest.NewRequest(http.MethodGet, expectedRoute, nil)
)

func dummyHandler(context.Context, *httpx.Request) httpx.Response {
	return httpx.NewEmpty(http.StatusOK)
}

func newFakeService(es []service.Endpoint) *service.FakeService {
	return &service.FakeService{
		NameHook:                       func() string { return "test" },
		EndpointsHook:                  func() []service.Endpoint { return es },
		HandlersHook:                   func() []httpx.Handler { return nil },
		MalformedRequestHandlerHook:    func() httpx.Handler { return nil },
		MethodNotAllowedHandlerHook:    func() httpx.Handler { return nil },
		RedirectHandlerHook:            func() httpx.Handler { return nil },
		InternalServerErrorHandlerHook: func() httpx.ErrorHandler { return nil },
	}
}
