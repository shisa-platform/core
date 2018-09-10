package gateway

import (
	"net/http"
	"net/http/httptest"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/service"
)

var (
	expectedRoute = "/test"
	fakeRequest   = httptest.NewRequest(http.MethodGet, expectedRoute, nil)
)

func dummyHandler(context.Context, *httpx.Request) httpx.Response {
	return httpx.NewEmpty(http.StatusOK)
}

func newFakeService(es []service.Endpoint) *service.Service {
	return &service.Service{
		Name:      "test",
		Endpoints: es,
	}
}
