package gateway

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func defaultNotFoundHandler(ctx context.Context, request *service.Request) service.Response {
	return service.NewEmpty(http.StatusNotFound)
}

func defaultMethodNotAlowedHandler(ctx context.Context, request *service.Request) service.Response {
	return service.NewEmpty(http.StatusMethodNotAllowed)
}

func defaultMalformedQueryParameterHandler(ctx context.Context, request *service.Request) service.Response {
	return service.NewEmpty(http.StatusBadRequest)
}

func defaultRedirectHandler(c context.Context, r *service.Request) (resp service.Response) {
	if r.Method == http.MethodGet {
		resp = service.NewEmpty(http.StatusSeeOther)
	} else {
		resp = service.NewEmpty(http.StatusTemporaryRedirect)
	}
	return
}

func defaultInternalServerErrorHandler(context.Context, *service.Request, merry.Error) service.Response {
	return service.NewEmpty(http.StatusInternalServerError)
}

func defaultRequestIDGenerator(request *service.Request) string {
	return request.GenerateID()
}
