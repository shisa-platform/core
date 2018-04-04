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

func newFakeService(es []service.Endpoint) *service.Service {
	return &service.Service{
		Name:      "test",
		Endpoints: es,
	}
}
