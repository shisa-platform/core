package gateway

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/service"
)

type endpoint struct {
	service.Endpoint
	serviceName       string
	badQueryHandler   httpx.Handler
	notAllowedHandler httpx.Handler
	redirectHandler   httpx.Handler
	iseHandler        httpx.ErrorHandler
}

func (e endpoint) handleNotAllowed(ctx context.Context, request *httpx.Request) (httpx.Response, merry.Error) {
	if e.notAllowedHandler == nil {
		return httpx.NewEmpty(http.StatusMethodNotAllowed), nil
	}

	response, exception := e.notAllowedHandler.InvokeSafely(ctx, request)
	if exception != nil {
		exception = exception.Prepend("gateway: route: run MethodNotAllowedHandler")
		response = httpx.NewEmpty(http.StatusMethodNotAllowed)
	}

	return response, exception
}

func (e endpoint) handleRedirect(ctx context.Context, request *httpx.Request) (httpx.Response, merry.Error) {
	if e.redirectHandler == nil {
		return redirect(ctx, request), nil
	}

	response, exception := e.redirectHandler.InvokeSafely(ctx, request)
	if exception != nil {
		exception = exception.Prepend("gateway: route: run RedirectHandler")
		response = redirect(ctx, request)
	}

	return response, exception
}

func redirect(ctx context.Context, request *httpx.Request) httpx.Response {
	location := *request.URL
	if location.Path[len(location.Path)-1] == '/' {
		location.Path = location.Path[:len(location.Path)-1]
	} else {
		location.Path = location.Path + "/"
	}

	if request.Method == http.MethodGet {
		return httpx.NewSeeOther(location.String())
	}

	return httpx.NewTemporaryRedirect(location.String())
}

func (e endpoint) handleBadQuery(ctx context.Context, request *httpx.Request) (httpx.Response, merry.Error) {
	if e.badQueryHandler == nil {
		return httpx.NewEmpty(http.StatusBadRequest), nil
	}

	response, exception := e.badQueryHandler.InvokeSafely(ctx, request)
	if exception != nil {
		exception = exception.Prepend("gateway: route: run MalformedRequestHandler")
		response = httpx.NewEmpty(http.StatusBadRequest)
	}

	return response, exception
}

func (e endpoint) handleError(ctx context.Context, request *httpx.Request, err merry.Error) (httpx.Response, merry.Error) {
	if e.iseHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err), nil
	}

	response, exception := e.iseHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		response = httpx.NewEmptyError(merry.HTTPCode(err), err)
		exception = exception.Prepend("gateway: route: run InternalServerErrorHandler")
	}

	return response, exception
}
