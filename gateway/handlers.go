package gateway

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func defaultNotFoundHandler(ctx context.Context, request *service.Request) service.Response {
	response := service.NewEmpty(http.StatusNotFound)
	response.Headers().Set(defaultRequestIDResponseHeader, ctx.RequestID())

	return response
}

func defaultMethodNotAlowedHandler(ctx context.Context, request *service.Request) service.Response {
	response := service.NewEmpty(http.StatusMethodNotAllowed)
	response.Headers().Set(defaultRequestIDResponseHeader, ctx.RequestID())

	return response
}

func defaultInternalServerErrorHandler(context.Context, *service.Request, merry.Error) service.Response {
	return service.NewEmpty(http.StatusInternalServerError)
}

func defaultRequestIDGenerator(request *service.Request) string {
	return request.GenerateID()
}
