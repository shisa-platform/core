package gateway

import (
	stdctx "context"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/ansel1/merry"

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

type byName []httpx.QueryParameter

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()

	ctx := context.Get(r.Context())
	defer context.Put(ctx)

	request := httpx.GetRequest(r)
	defer httpx.PutRequest(request)

	requestIDGenerationStart := time.Now().UTC()
	requestID, idErr := g.RequestIDGenerator(ctx, request)
	if idErr != nil {
		idErr = merry.WithMessage(idErr, "generating request id")
		requestID = request.ID()
	}
	if requestID == "" {
		idErr = merry.New("generator returned empty request id")
		requestID = request.ID()
	}
	requestIDGenerationStop := time.Now().UTC()

	ctx = ctx.WithRequestID(requestID)

	request.ParseQueryParameters()

	var (
		err           merry.Error
		response      httpx.Response
		pipeline      *service.Pipeline
		findPathStart time.Time
		findPathStop  time.Time
		pipelineStart time.Time
		pipelineStop  time.Time
		handlersStop  time.Time
		endpoint      *endpoint
		path          string
		params        []httpx.QueryParameter
		pathParams    []httpx.PathParameter
		tsr           bool
		invalidParams bool
		missingParams bool
	)

	handlersStart := time.Now().UTC()
	for i, handler := range g.Handlers {
		response = handler.InvokeSafely(ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running gateway handler").WithValue("index", i)
			response = g.handleErrorSafely(g.InternalServerErrorHandler, ctx, request, err)
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
	endpoint, pathParams, tsr, err = g.tree.getValue(path)
	findPathStop = time.Now().UTC()

	if err != nil {
		err = merry.WithMessage(err, "routing request")
		response = g.handleErrorSafely(g.InternalServerErrorHandler, ctx, request, err)
		goto finish
	}

	if endpoint == nil {
		response = g.NotFoundHandler.InvokeSafely(ctx, request, &err)
		if err != nil {
			response = defaultNotFoundHandler(ctx, request)
			err = merry.WithMessage(err, "running NotFoundHandler")
		}
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
			response = g.NotFoundHandler.InvokeSafely(ctx, request, &err)
			if err != nil {
				err = merry.WithMessage(err, "running NotFoundHandler")
				response = defaultNotFoundHandler(ctx, request)
			}
		} else {
			response = endpoint.notAllowedHandler.InvokeSafely(ctx, request, &err)
			if err != nil {
				err = merry.WithMessage(err, "running MethodNotAllowedHandler")
				response = defaultMethodNotAlowedHandler(ctx, request)
			}
		}
		goto finish
	}

	if tsr {
		if path != "/" && pipeline.Policy.AllowTrailingSlashRedirects {
			response = endpoint.redirectHandler.InvokeSafely(ctx, request, &err)
			if err != nil {
				err = merry.WithMessage(err, "running RedirectHandler")
				response = defaultRedirectHandler(ctx, request)
			}
		} else {
			response = g.NotFoundHandler.InvokeSafely(ctx, request, &err)
			if err != nil {
				err = merry.WithMessage(err, "running NotFoundHandler")
				response = defaultNotFoundHandler(ctx, request)
			}
		}
		goto finish
	}

	if len(pipeline.QueryFields) == 0 {
		for _, p := range request.QueryParams {
			if p.Invalid {
				invalidParams = true
				break
			}
		}
	} else {
		params = append(params, request.QueryParams...)
		sort.Sort(byName(params))
	}

	for _, field := range pipeline.QueryFields {
		var found bool
		for j, p := range params {
			if p.Invalid {
				invalidParams = true
			}
			if field.Match(p.Name) {
				found = true
				request.QueryParams[p.Ordinal].Unknown = false
				params = append(params[:j], params[j+1:]...)
				if err := field.Validate(p.Values); err != nil {
					request.QueryParams[p.Ordinal].Invalid = true
					invalidParams = true
				}
				break
			}
		}

		if !found {
			if field.Default != "" {
				dp := httpx.QueryParameter{
					Name:    field.Name,
					Values:  []string{field.Default},
					Ordinal: -1,
				}
				request.QueryParams = append(request.QueryParams, dp)
			} else if field.Required {
				missingParams = true
			}
		}
	}

	if (invalidParams || missingParams) && !pipeline.Policy.AllowMalformedQueryParameters {
		response = endpoint.badQueryHandler.InvokeSafely(ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running MalformedRequestHandler")
			response = defaultMalformedRequestHandler(ctx, request)
		}
		goto finish
	}

	if len(params) != 0 && !pipeline.Policy.AllowUnknownQueryParameters {
		response = endpoint.badQueryHandler.InvokeSafely(ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running MalformedRequestHandler")
			response = defaultMalformedRequestHandler(ctx, request)
		}
		goto finish
	}

	request.PathParams = pathParams
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
			response = g.handleErrorSafely(endpoint.iseHandler, ctx, request, err)
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
		response = g.handleErrorSafely(endpoint.iseHandler, ctx, request, err)
	}

finish:
	serializationStart := time.Now().UTC()
	for k, vs := range response.Headers() {
		w.Header()[k] = vs
	}
	for k := range response.Trailers() {
		w.Header().Add("Trailer", k)
	}

	w.Header().Set(g.RequestIDHeaderName, requestID)

	w.WriteHeader(response.StatusCode())
	size, writeErr := response.Serialize(w)

	for k, vs := range response.Trailers() {
		w.Header()[k] = vs
	}

	if f, impl := w.(http.Flusher); impl {
		f.Flush()
	}

	end := time.Now().UTC()

	if g.CompletionHook != nil {
		snapshot := httpx.ResponseSnapshot{
			StatusCode: response.StatusCode(),
			Size:       size,
			Start:      start,
			Elapsed:    end.Sub(start),
			Metrics:    make(map[string]time.Duration),
		}
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

func (g *Gateway) handleErrorSafely(h httpx.ErrorHandler, ctx context.Context, request *httpx.Request, err merry.Error) httpx.Response {
	var exception merry.Error
	response := h.InvokeSafely(ctx, request, err, &exception)
	if exception != nil {
		response = defaultInternalServerErrorHandler(ctx, request, err)
		exception = merry.WithMessage(exception, "while invoking InternalServerErrorHandler")
		g.invokeErrorHookSafely(ctx, request, exception)
	}

	return response
}

func (g *Gateway) invokeErrorHookSafely(ctx context.Context, request *httpx.Request, err merry.Error) {
	var exception merry.Error

	g.ErrorHook.invokeSafely(ctx, request, err, &exception)
	if exception != nil {
		g.defaultErrorHook(ctx, request, err)
		exception = merry.WithMessage(exception, "while invoking ErrorHook")
		g.defaultErrorHook(ctx, request, exception)
	}
}

func (g *Gateway) invokeCompletionHookSafely(ctx context.Context, request *httpx.Request, snapshot httpx.ResponseSnapshot) {
	if g.CompletionHook == nil {
		return
	}

	var exception merry.Error
	g.CompletionHook.invokeSafely(ctx, request, snapshot, &exception)
	if exception != nil {
		exception = merry.WithMessage(exception, "while invoking CompletionHook")
		g.invokeErrorHookSafely(ctx, request, exception)
	}
}
