package gateway

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
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

	var err merry.Error
	response := e.notAllowedHandler.InvokeSafely(ctx, request, &err)
	if err != nil {
		err = merry.WithMessage(err, "running MethodNotAllowedHandler")
		response = httpx.NewEmpty(http.StatusMethodNotAllowed)
	}

	return response, err
}

func (e endpoint) handleRedirect(ctx context.Context, request *httpx.Request) (httpx.Response, merry.Error) {
	if e.redirectHandler == nil {
		return redirect(ctx, request), nil
	}

	var err merry.Error
	response := e.redirectHandler.InvokeSafely(ctx, request, &err)
	if err != nil {
		err = merry.WithMessage(err, "running RedirectHandler")
		response = redirect(ctx, request)
	}

	return response, err
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

	var err merry.Error
	response := e.badQueryHandler.InvokeSafely(ctx, request, &err)
	if err != nil {
		err = merry.WithMessage(err, "running MalformedRequestHandler")
		response = httpx.NewEmpty(http.StatusBadRequest)
	}

	return response, err
}

func (e endpoint) handleError(ctx context.Context, request *httpx.Request, err merry.Error) (httpx.Response, merry.Error) {
	if e.iseHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err), nil
	}

	var exception merry.Error
	response := e.iseHandler.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		response = httpx.NewEmptyError(merry.HTTPCode(err), err)
		exception = merry.WithMessage(exception, "invoking InternalServerErrorHandler")
	}

	return response, exception
}
