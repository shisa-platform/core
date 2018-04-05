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

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/errorx"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parent := opentracing.StartSpan("ServiceRequest", opentracing.Tags{
		string(ext.SpanKind):  "server",
		string(ext.Component): "router",
	})
	defer parent.Finish()

	ri := httpx.NewInterceptor(w)

	ctx := context.Get(r.Context())
	defer context.Put(ctx)

	request := httpx.GetRequest(r)
	defer httpx.PutRequest(request)

	path := request.URL.EscapedPath()
	ext.HTTPUrl.Set(parent, path)
	ext.HTTPMethod.Set(parent, request.Method)

	ctx = ctx.WithSpan(parent)

	requestID, idErr := g.generateRequestID(ctx, request)

	parent.SetTag("request_id", requestID)
	ctx = ctx.WithRequestID(requestID)
	ri.Header().Set(g.RequestIDHeaderName, requestID)

	span := opentracing.StartSpan("ParseQueryParameters", opentracing.ChildOf(parent.Context()))
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
		endpoint   *endpoint
		pipeline   *service.Pipeline
		err        merry.Error
		response   httpx.Response
		tsr        bool
		responseCh chan httpx.Response = make(chan httpx.Response, 1)
	)

	span = opentracing.StartSpan("RunGatewayHandlers", opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	subCtx := ctx.WithSpan(span)
	if g.HandlersTimeout != 0 {
		var subCancel stdctx.CancelFunc
		subCtx, subCancel = subCtx.WithTimeout(g.HandlersTimeout - ri.Elapsed())
		defer subCancel()
	}

	for _, handler := range g.Handlers {
		go func() {
			response, exception := handler.InvokeSafely(subCtx, request)
			if exception != nil {
				err = exception.Prepend("gateway: route: run gateway handler")
				response = g.handleError(subCtx, request, err)
			}

			responseCh <- response
		}()
		select {
		case <-subCtx.Done():
			cancel()
			err = merry.Prepend(subCtx.Err(), "gateway: route: request aborted")
			if merry.Is(subCtx.Err(), stdctx.DeadlineExceeded) {
				err = err.WithHTTPCode(http.StatusGatewayTimeout)
			}
			response = g.handleError(subCtx, request, err)
			goto finish
		case response = <-responseCh:
			if response != nil {
				goto finish
			}
		}
	}
	span.Finish()

	span = opentracing.StartSpan("FindEndpoint", opentracing.ChildOf(parent.Context()))
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

	span = opentracing.StartSpan("ValidateQueryParameters", opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	if malformed, unknown, exception := request.ValidateQueryParameters(pipeline.QueryFields); exception != nil {
		response, exception = endpoint.handleError(ctx, request, exception)
		if exception != nil {
			g.invokeErrorHookSafely(ctx, request, exception)
		}
		goto finish
	} else if malformed && !pipeline.Policy.AllowMalformedQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
		goto finish
	} else if unknown && !pipeline.Policy.AllowUnknownQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
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

	span = opentracing.StartSpan("RunGatewayHandlers", opentracing.ChildOf(parent.Context()))
	defer span.Finish()

	if pipeline.Policy.TimeBudget != 0 {
		var cancel stdctx.CancelFunc
		ctx, cancel = ctx.WithTimeout(pipeline.Policy.TimeBudget - ri.Elapsed())
		defer cancel()
	}

	select {
	case <-ctx.Done():
		err = merry.Prepend(ctx.Err(), "gateway: route: request aborted")
		if merry.Is(ctx.Err(), stdctx.DeadlineExceeded) {
			err = err.WithHTTPCode(http.StatusGatewayTimeout)
		}
		response = g.handleEndpointError(endpoint, ctx, request, err)
		goto finish
	default:
	}

	subCtx = ctx.WithSpan(span)
endpointHandlers:
	for _, handler := range pipeline.Handlers {
		go func() {
			response, exception := handler.InvokeSafely(subCtx, request)
			if exception != nil {
				err = exception.Prepend("gateway: route: run endpoint handler")
				response = g.handleEndpointError(endpoint, subCtx, request, err)
			}

			responseCh <- response
		}()
		select {
		case <-subCtx.Done():
			err = merry.Prepend(subCtx.Err(), "gateway: route: request aborted")
			if merry.Is(subCtx.Err(), stdctx.DeadlineExceeded) {
				err = err.WithHTTPCode(http.StatusGatewayTimeout)
			}
			response = g.handleEndpointError(endpoint, subCtx, request, err)
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

	if response == nil {
		err = merry.New("gateway: route: no response from pipeline")
		response = g.handleEndpointError(endpoint, subCtx, request, err)
	}

finish:
	if span != nil {
		span.Finish()
	}
	span = opentracing.StartSpan("SerializeResponse", opentracing.ChildOf(parent.Context()))
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

	if g.CompletionHook != nil {
		g.invokeCompletionHookSafely(ctx, request, snapshot)
	}

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
	parent := opentracing.SpanFromContext(ctx)
	span := opentracing.StartSpan("GenerateRequestID", opentracing.ChildOf(parent.Context()))
	defer span.Finish()

	if g.RequestIDGenerator == nil {
		return request.ID(), nil
	}

	requestID, err, exception := g.RequestIDGenerator.InvokeSafely(ctx, request)
	if exception != nil {
		err = exception.Prepend("gateway: route: generate request id")
		span.LogFields(otlog.String("error", err.Error()))
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
	if g.ErrorHook == nil {
		g.fallbackErrorHook(ctx, err)
	}

	if exception := g.ErrorHook.InvokeSafely(ctx, request, err); exception != nil {
		g.fallbackErrorHook(ctx, err)
		exception = exception.Prepend("gateway: route: run ErrorHook")
		g.fallbackErrorHook(ctx, exception)
	}
}

func (g *Gateway) fallbackErrorHook(ctx context.Context, err merry.Error) {
	log.Println(ctx.RequestID(), err.Error())
	if errorx.IsPanic(err) {
		log.Print(ctx.RequestID(), " stack trace\n", merry.Stacktrace(err))
	}
}

func (g *Gateway) invokeCompletionHookSafely(ctx context.Context, request *httpx.Request, snapshot httpx.ResponseSnapshot) {
	if exception := g.CompletionHook.InvokeSafely(ctx, request, snapshot); exception != nil {
		exception = exception.Prepend("gateway: route: run CompletionHook")
		g.invokeErrorHookSafely(ctx, request, exception)
	}
}
