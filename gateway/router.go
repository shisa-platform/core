package gateway

import (
	stdctx "context"
	"net/http"
	"net/url"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

const (
	// RequestIdGenerationMetricKey is the `ResponseSnapshot` metric for generating the request id
	RequestIdGenerationMetricKey = "request-id-generation"
	// FindEndpointMetricKey is the `ResponseSnapshot` metric for resolving the request's endpoint
	FindEndpointMetricKey = "find-endpoint"
	// RunGatewayHandlersMetricKey is the `ResponseSnapshot` metric for running the Gateway level handlers
	RunGatewayHandlersMetricKey = "handlers"
	// RunEndpointPipelineMetricKey is the `ResponseSnapshot` metric for running the endpoint's pipeline
	RunEndpointPipelineMetricKey = "pipeline"
	// SerializeResponseMetricKey is the `ResponseSnapshot` metric for serializing the response
	SerializeResponseMetricKey = "serialization"
)

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ri := httpx.NewInterceptor(w)

	ctx := context.Get(r.Context())
	defer context.Put(ctx)

	request := httpx.GetRequest(r)
	defer httpx.PutRequest(request)

	requestIDGenerationStart := time.Now().UTC()
	requestID, idErr := g.generateRequestID(ctx, request)
	requestIDGenerationStop := time.Now().UTC()

	ctx = ctx.WithRequestID(requestID)
	ri.Header().Set(g.RequestIDHeaderName, requestID)

	parseOK := request.ParseQueryParameters()

	var (
		path          string
		endpoint      *endpoint
		pipeline      *service.Pipeline
		err           merry.Error
		response      httpx.Response
		findPathStart time.Time
		findPathStop  time.Time
		pipelineStart time.Time
		pipelineStop  time.Time
		handlersStop  time.Time
		tsr           bool
	)

	handlersStart := time.Now().UTC()
	for i, handler := range g.Handlers {
		response = handler.InvokeSafely(ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running gateway handler").WithValue("index", i)
			response = g.handleError(ctx, request, err)
			handlersStop = time.Now().UTC()
			goto finish
		}
		if response != nil {
			handlersStop = time.Now().UTC()
			goto finish
		}
	}
	handlersStop = time.Now().UTC()

	findPathStart = time.Now().UTC()
	path = request.URL.EscapedPath()
	endpoint, request.PathParams, tsr, err = g.tree.getValue(path)
	findPathStop = time.Now().UTC()

	if err != nil {
		err = merry.WithMessage(err, "routing request")
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

	if malformed, unknown, vErr := request.ValidateQueryParameters(pipeline.QueryFields); vErr != nil {
		var exception merry.Error
		response, exception = endpoint.handleError(ctx, request, vErr)
		if exception != nil {
			g.invokeErrorHookSafely(ctx, request, exception)
		}
	} else if malformed && !pipeline.Policy.AllowMalformedQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
		goto finish
	} else if unknown && !pipeline.Policy.AllowUnknownQueryParameters {
		response, err = endpoint.handleBadQuery(ctx, request)
		goto finish
	}

	if !pipeline.Policy.PreserveEscapedPathParameters {
		for i := range request.PathParams {
			if esc, r := url.PathUnescape(request.PathParams[i].Value); r == nil {
				request.PathParams[i].Value = esc
			}
		}
	}

	if pipeline.Policy.TimeBudget != 0 {
		// xxx - watch for timeout and kill pipeline, return
		var cancel stdctx.CancelFunc
		ctx, cancel = ctx.WithTimeout(pipeline.Policy.TimeBudget)
		defer cancel()
	}

	pipelineStart = time.Now().UTC()
	for i, handler := range pipeline.Handlers {
		response = handler.InvokeSafely(ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running endpoint handler").WithValue("index", i)
			var exception merry.Error
			response, exception = endpoint.handleError(ctx, request, err)
			if exception != nil {
				g.invokeErrorHookSafely(ctx, request, exception)
			}
			pipelineStop = time.Now().UTC()
			goto finish
		}
		if response != nil {
			break
		}
	}
	pipelineStop = time.Now().UTC()

	if response == nil {
		err = merry.New("no response from pipeline")
		var exception merry.Error
		response, exception = endpoint.handleError(ctx, request, err)
		if exception != nil {
			g.invokeErrorHookSafely(ctx, request, exception)
		}
	}

finish:
	serializationStart := time.Now().UTC()
	writeErr := ri.WriteResponse(response)
	snapshot := ri.Flush()

	end := time.Now().UTC()

	if g.CompletionHook != nil {
		idGeneration := requestIDGenerationStop.Sub(requestIDGenerationStart)
		snapshot.Metrics[RequestIdGenerationMetricKey] = idGeneration
		if len(g.Handlers) != 0 {
			snapshot.Metrics[RunGatewayHandlersMetricKey] = handlersStop.Sub(handlersStart)
		}
		if !findPathStart.IsZero() {
			snapshot.Metrics[FindEndpointMetricKey] = findPathStop.Sub(findPathStart)
		}
		if !pipelineStart.IsZero() {
			snapshot.Metrics[RunEndpointPipelineMetricKey] = pipelineStop.Sub(pipelineStart)
		}
		snapshot.Metrics[SerializeResponseMetricKey] = end.Sub(serializationStart)

		g.invokeCompletionHookSafely(ctx, request, snapshot)
	}

	if idErr != nil {
		g.invokeErrorHookSafely(ctx, request, idErr)
	}

	if err != nil {
		g.invokeErrorHookSafely(ctx, request, err)
	}

	writeErr1 := merry.WithMessage(writeErr, "serializing response")
	if writeErr1 != nil {
		g.invokeErrorHookSafely(ctx, request, writeErr1)
	}

	respErr := response.Err()
	if respErr != nil && respErr != err {
		g.invokeErrorHookSafely(ctx, request, merry.WithMessage(respErr, "handler failed"))
	}
}

func (g *Gateway) generateRequestID(ctx context.Context, request *httpx.Request) (string, merry.Error) {
	if g.RequestIDGenerator == nil {
		return request.ID(), nil
	}

	requestID, err := g.RequestIDGenerator.InvokeSafely(ctx, request)
	if err != nil {
		err = merry.WithMessage(err, "generating request id")
		requestID = request.ID()
	}
	if requestID == "" {
		err = merry.New("generator returned empty request id")
		requestID = request.ID()
	}

	return requestID, err
}

func (g *Gateway) handleNotFound(ctx context.Context, request *httpx.Request) (httpx.Response, merry.Error) {
	if g.NotFoundHandler == nil {
		return httpx.NewEmpty(http.StatusNotFound), nil
	}

	var exception merry.Error
	response := g.NotFoundHandler.InvokeSafely(ctx, request, &exception)
	if exception != nil {
		err := merry.WithMessage(exception, "running NotFoundHandler")
		return httpx.NewEmpty(http.StatusNotFound), err
	}

	return response, nil
}

func (g *Gateway) handleError(ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	if g.InternalServerErrorHandler == nil {
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	var exception merry.Error
	response := g.InternalServerErrorHandler.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		response = httpx.NewEmptyError(http.StatusInternalServerError, err)
		exception = merry.WithMessage(exception, "while invoking InternalServerErrorHandler")
		g.invokeErrorHookSafely(ctx, request, exception)
	}

	return response
}

func (g *Gateway) invokeErrorHookSafely(ctx context.Context, request *httpx.Request, err merry.Error) {
	var exception merry.Error

	g.ErrorHook.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		g.Logger.Error(err.Error(), zap.String("request-id", ctx.RequestID()), zap.Error(err))
		exception = merry.WithMessage(exception, "while invoking ErrorHook")
		g.Logger.Error(exception.Error(), zap.String("request-id", ctx.RequestID()), zap.Error(exception))
	}
}

func (g *Gateway) invokeCompletionHookSafely(ctx context.Context, request *httpx.Request, snapshot httpx.ResponseSnapshot) {
	var exception merry.Error

	g.CompletionHook.InvokeSafely(ctx, request, snapshot, &exception)
	if exception != nil {
		exception = merry.WithMessage(exception, "while invoking CompletionHook")
		g.invokeErrorHookSafely(ctx, request, exception)
	}
}
