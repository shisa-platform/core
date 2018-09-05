package gateway

import (
	stdctx "context"
	"log"
	"net/http"
	"net/url"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/service"
)

var (
	routerTags = opentracing.Tags{
		string(ext.SpanKind):  "server",
		string(ext.Component): "router",
	}
)

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parent := opentracing.StartSpan("ServiceRequest", routerTags)
	defer parent.Finish()

	ri := httpx.NewInterceptor(w)

	ctx := context.New(r.Context())
	ctx = ctx.WithSpan(parent)

	request := httpx.GetRequest(r)
	defer httpx.PutRequest(request)

	ext.HTTPUrl.Set(parent, request.URL.RequestURI())
	ext.HTTPMethod.Set(parent, request.Method)

	requestID, idErr := g.generateRequestID(ctx, request)

	parent.SetTag("request_id", requestID)
	ctx = ctx.WithRequestID(requestID)
	ri.Header().Set(g.RequestIDHeaderName, requestID)

	span := ctx.StartSpan("ParseQueryParameters")
	parseOK := request.ParseQueryParameters()
	span.Finish()

	var cancel stdctx.CancelFunc
	ctx, cancel = ctx.WithCancel()
	defer cancel()

	if cn, ok := w.(http.CloseNotifier); ok {
		go func() {
			select {
			case <-cn.CloseNotify():
				cancel()
			case <-ctx.Done():
			}
		}()
	}

	var (
		path       string = request.URL.EscapedPath()
		endpoint   *endpoint
		pipeline   *service.Pipeline
		err        merry.Error
		response   httpx.Response
		tsr        bool
		responseCh chan httpx.Response = make(chan httpx.Response, 1)
	)

	span = ctx.StartSpan("RunGatewayHandlers")
	pipelineCtx := ctx
	pipelineCtx.WithSpan(span)

	if g.HandlersTimeout != 0 {
		var subCancel stdctx.CancelFunc
		pipelineCtx, subCancel = pipelineCtx.WithTimeout(g.HandlersTimeout - ri.Elapsed())
		defer subCancel()
	}

	for _, handler := range g.Handlers {
		go func() {
			response, exception := handler.InvokeSafely(pipelineCtx, request)
			if exception != nil {
				err = exception.Prepend("gateway: route: run gateway handler")
				response = g.handleError(pipelineCtx, request, err)
			}

			responseCh <- response
		}()
		select {
		case <-pipelineCtx.Done():
			cancel()
			err = merry.Prepend(pipelineCtx.Err(), "gateway: route: request aborted")
			if merry.Is(pipelineCtx.Err(), stdctx.DeadlineExceeded) {
				err = err.WithHTTPCode(http.StatusGatewayTimeout)
			}
			response = g.handleError(pipelineCtx, request, err)
			span.Finish()
			goto finish
		case response = <-responseCh:
			if response != nil {
				span.Finish()
				goto finish
			}
		}
	}
	span.Finish()
	ctx = ctx.WithSpan(parent)

	span = ctx.StartSpan("FindEndpoint")
	endpoint, request.PathParams, tsr, err = g.tree.getValue(path)
	span.Finish()

	if err != nil {
		err = err.Prepend("gateway: route")
		response = g.handleError(ctx, request, err)
		goto finish
	}

	if endpoint == nil {
		response, err = g.handleNotFound(ctx, request)
		goto finish
	}

	switch request.Method {
	case http.MethodHead:
		pipeline = endpoint.Head
	case http.MethodGet:
		pipeline = endpoint.Get
	case http.MethodPut:
		pipeline = endpoint.Put
	case http.MethodPost:
		pipeline = endpoint.Post
	case http.MethodPatch:
		pipeline = endpoint.Patch
	case http.MethodDelete:
		pipeline = endpoint.Delete
	case http.MethodConnect:
		pipeline = endpoint.Connect
	case http.MethodOptions:
		pipeline = endpoint.Options
	case http.MethodTrace:
		pipeline = endpoint.Trace
	}

	if pipeline == nil {
		if tsr {
			response, err = g.handleNotFound(ctx, request)
		} else {
			response, err = endpoint.handleNotAllowed(ctx, request)
		}
		goto finish
	}

	if tsr {
		if path != "/" && pipeline.Policy.AllowTrailingSlashRedirects {
			response, err = endpoint.handleRedirect(ctx, request)
		} else {
			response, err = g.handleNotFound(ctx, request)
		}
		goto finish
	}

	if !parseOK && !pipeline.Policy.AllowMalformedQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
		goto finish
	}

	span = ctx.StartSpan("ValidateQueryParameters")
	if malformed, unknown, exception := request.ValidateQueryParameters(pipeline.QuerySchemas); exception != nil {
		response, exception = endpoint.handleError(ctx, request, exception)
		if exception != nil {
			g.invokeErrorHookSafely(ctx, request, exception)
		}
		span.Finish()
		goto finish
	} else if malformed && !pipeline.Policy.AllowMalformedQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
		span.Finish()
		goto finish
	} else if unknown && !pipeline.Policy.AllowUnknownQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
		span.Finish()
		goto finish
	}
	span.Finish()

	if !pipeline.Policy.PreserveEscapedPathParameters {
		for i := range request.PathParams {
			if esc, r := url.PathUnescape(request.PathParams[i].Value); r == nil {
				request.PathParams[i].Value = esc
			}
		}
	}

	span = ctx.StartSpan("RunPipelineHandlers")
	pipelineCtx = ctx
	pipelineCtx.WithSpan(span)

	if pipeline.Policy.TimeBudget != 0 {
		var subCancel stdctx.CancelFunc
		pipelineCtx, subCancel = pipelineCtx.WithTimeout(pipeline.Policy.TimeBudget - ri.Elapsed())
		defer subCancel()
	}

endpointHandlers:
	for _, handler := range pipeline.Handlers {
		go func() {
			response, exception := handler.InvokeSafely(pipelineCtx, request)
			if exception != nil {
				err = exception.Prepend("gateway: route: run endpoint handler")
				response = g.handleEndpointError(endpoint, pipelineCtx, request, err)
			}

			responseCh <- response
		}()
		select {
		case <-pipelineCtx.Done():
			err = merry.Prepend(pipelineCtx.Err(), "gateway: route: request aborted")
			if merry.Is(pipelineCtx.Err(), stdctx.DeadlineExceeded) {
				err = err.WithHTTPCode(http.StatusGatewayTimeout)
			}
			response = g.handleEndpointError(endpoint, pipelineCtx, request, err)
			span.Finish()
			goto finish
		case response = <-responseCh:
			if response != nil {
				if respErr := response.Err(); respErr != nil {
					ext.Error.Set(span, true)
					span.LogFields(otlog.String("error", respErr.Error()))
				}
				break endpointHandlers
			}
		}
	}
	span.Finish()
	ctx = ctx.WithSpan(parent)

	if response == nil {
		err = merry.New("gateway: route: no response from pipeline")
		response = g.handleEndpointError(endpoint, pipelineCtx, request, err)
	}

finish:
	ctx = ctx.WithSpan(parent)

	span = ctx.StartSpan("SerializeResponse")
	var (
		writeErr merry.Error
		snapshot httpx.ResponseSnapshot
	)
	if merry.Is(ctx.Err(), stdctx.Canceled) {
		writeErr = merry.New("gateway: route: user agent disconnect")
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", writeErr.Error()))
		snapshot = ri.Snapshot()
	} else {
		writeErr = ri.WriteResponse(response)
		writeErr = merry.Prepend(writeErr, "gateway: route: serialize response")
		if writeErr != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.String("error", writeErr.Error()))
		}
		snapshot = ri.Flush()
	}
	span.Finish()

	ext.HTTPStatusCode.Set(parent, uint16(snapshot.StatusCode))

	g.invokeCompletionHookSafely(ctx, request, snapshot)

	if idErr != nil {
		g.invokeErrorHookSafely(ctx, request, idErr)
	}

	if err != nil {
		ext.Error.Set(parent, true)
		parent.LogFields(otlog.String("error", err.Error()))
		g.invokeErrorHookSafely(ctx, request, err)
	}

	if writeErr != nil {
		g.invokeErrorHookSafely(ctx, request, writeErr)
	}

	if respErr := response.Err(); respErr != nil && respErr != err {
		respErr1 := merry.Prepend(respErr, "gateway: route: handler failed")
		g.invokeErrorHookSafely(ctx, request, respErr1)
	}
}

func (g *Gateway) generateRequestID(ctx context.Context, request *httpx.Request) (string, merry.Error) {
	span := ctx.StartSpan("GenerateRequestID")
	defer span.Finish()

	if g.RequestIDGenerator == nil {
		return request.ID(), nil
	}

	requestID, err, exception := g.RequestIDGenerator.InvokeSafely(ctx, request)
	if exception != nil {
		err = exception.Prepend("gateway: route: generate request id")
		span.LogFields(otlog.String("exception", err.Error()))
		requestID = request.ID()
	} else if err != nil {
		err = err.Prepend("gateway: route: generate request id")
		span.LogFields(otlog.String("error", err.Error()))
		requestID = request.ID()
	} else if requestID == "" {
		err = merry.New("gateway: route: generate request id: empty value")
		span.LogFields(otlog.String("error", err.Error()))
		requestID = request.ID()
	}

	return requestID, err
}

func (g *Gateway) handleNotFound(ctx context.Context, request *httpx.Request) (httpx.Response, merry.Error) {
	if g.NotFoundHandler == nil {
		return httpx.NewEmpty(http.StatusNotFound), nil
	}

	response, exception := g.NotFoundHandler.InvokeSafely(ctx, request)
	if exception != nil {
		err := exception.Prepend("gateway: route: run NotFoundHandler")
		return httpx.NewEmpty(http.StatusNotFound), err
	}

	return response, nil
}

func (g *Gateway) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	if g.InternalServerErrorHandler == nil {
		return httpx.NewEmptyError(merry.HTTPCode(err), err)
	}

	response, exception := g.InternalServerErrorHandler.InvokeSafely(ctx, request, err)
	if exception != nil {
		response = httpx.NewEmptyError(merry.HTTPCode(err), err)
		exception = exception.Prepend("gateway: route: run InternalServerErrorHandler")
		g.invokeErrorHookSafely(ctx, request, exception)
	}

	return response
}

func (g *Gateway) handleEndpointError(endpoint *endpoint, ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	response, exception := endpoint.handleError(ctx, request, err)
	if exception != nil {
		g.invokeErrorHookSafely(ctx, request, exception)
	}

	return response
}

func (g *Gateway) invokeErrorHookSafely(ctx context.Context, request *httpx.Request, err merry.Error) {
	span := ctx.StartSpan("ErrorHook")
	defer span.Finish()

	if g.ErrorHook == nil {
		g.fallbackErrorHook(ctx, err)
	}

	if exception := g.ErrorHook.InvokeSafely(ctx, request, err); exception != nil {
		g.fallbackErrorHook(ctx, err)
		exception = exception.Prepend("gateway: route: run ErrorHook")
		g.fallbackErrorHook(ctx, exception)
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("exception", err.Error()))
	}
}

func (g *Gateway) fallbackErrorHook(ctx context.Context, err merry.Error) {
	log.Println(ctx.RequestID(), err.Error())
	if errorx.IsPanic(err) {
		log.Print(ctx.RequestID(), " stack trace\n", merry.Stacktrace(err))
	}
}

func (g *Gateway) invokeCompletionHookSafely(ctx context.Context, request *httpx.Request, snapshot httpx.ResponseSnapshot) {
	if g.CompletionHook == nil {
		return
	}

	span := ctx.StartSpan("CompletionHook")
	defer span.Finish()

	if exception := g.CompletionHook.InvokeSafely(ctx, request, snapshot); exception != nil {
		exception = exception.Prepend("gateway: route: run CompletionHook")
		g.invokeErrorHookSafely(ctx, request, exception)
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("exception", exception.Error()))
	}
}
