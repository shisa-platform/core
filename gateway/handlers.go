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
	location := *r.URL
	if location.Path[len(location.Path)-1] == '/' {
		location.Path = location.Path[:len(location.Path)-1]
	} else {
		location.Path = location.Path + "/"
	}

	if r.Method == http.MethodGet {
		resp = service.NewSeeOther(location.String())
	} else {
		resp = service.NewTemporaryRedirect(location.String())
	}

	return
}

func defaultInternalServerErrorHandler(context.Context, *service.Request, merry.Error) service.Response {
	return service.NewEmpty(http.StatusInternalServerError)
}

func defaultRequestIDGenerator(c context.Context, r *service.Request) (string, merry.Error) {
	return r.GenerateID(), nil
}
