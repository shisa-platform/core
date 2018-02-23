package gateway

import (
	"net/http"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

func defaultNotFoundHandler(ctx context.Context, request *service.Request) httpx.Response {
	return service.NewEmpty(http.StatusNotFound)
}

func defaultMethodNotAlowedHandler(ctx context.Context, request *service.Request) httpx.Response {
	return service.NewEmpty(http.StatusMethodNotAllowed)
}

func defaultMalformedRequestHandler(ctx context.Context, request *service.Request) httpx.Response {
	return service.NewEmpty(http.StatusBadRequest)
}

func defaultRedirectHandler(c context.Context, r *service.Request) (resp httpx.Response) {
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

func defaultInternalServerErrorHandler(ctx context.Context, r *service.Request, err merry.Error) httpx.Response {
	return service.NewEmptyError(http.StatusInternalServerError, err)
}

func defaultRequestIDGenerator(c context.Context, r *service.Request) (string, merry.Error) {
	return r.ID(), nil
}

func (g *Gateway) defaultErrorHook(ctx context.Context, _ *service.Request, err merry.Error) {
	g.Logger.Error(err.Error(), zap.String("request-id", ctx.RequestID()), zap.Error(err))
}
