package gateway

import (
	"net/http"
	"net/http/httptest"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

var (
	expectedRoute = "/test"
	fakeRequest   = httptest.NewRequest(http.MethodGet, expectedRoute, nil)
)

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
