package gateway

import (
	stdctx "context"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

type byName []service.QueryParameter

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()

	ctx := context.New(r.Context())
	request := &service.Request{Request: r}

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
	requestIDGenerationTime := time.Now().UTC().Sub(requestIDGenerationStart)

	ctx = context.WithRequestID(ctx, requestID)

	request.ParseQueryParameters()

	var (
		err           merry.Error
		response      service.Response
		pipeline      *service.Pipeline
		findPathStart time.Time
		pipelineStart time.Time
		pipelineTime  time.Duration
		handlersTime  time.Duration
		findPathTime  time.Duration
		endpoint      *endpoint
		path          string
		params        []service.QueryParameter
		pathParams    []service.PathParameter
		tsr           bool
		invalidParams bool
		missingParams bool
	)

	handlersStart := time.Now().UTC()
	for i, handler := range g.Handlers {
		response = runHandler(handler, ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running gateway handler").WithValue("index", i)
			response = g.InternalServerErrorHandler(ctx, request, err)
			goto finish
		}
		if response != nil {
			goto finish
		}
	}
	handlersTime = time.Now().UTC().Sub(handlersStart)

	findPathStart = time.Now().UTC()
	path = request.URL.EscapedPath()
	endpoint, pathParams, tsr, err = g.tree.getValue(path)
	findPathTime = time.Now().UTC().Sub(findPathStart)

	if err != nil {
		err = merry.WithMessage(err, "routing request")
		response = g.InternalServerErrorHandler(ctx, request, err)
		goto finish
	}

	if endpoint == nil {
		response = g.NotFoundHandler(ctx, request)
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
			response = g.NotFoundHandler(ctx, request)
		} else {
			response = endpoint.notAllowedHandler(ctx, request)
		}
		goto finish
	}

	if tsr {
		if path != "/" && pipeline.Policy.AllowTrailingSlashRedirects {
			response = endpoint.redirectHandler(ctx, request)
		} else {
			response = g.NotFoundHandler(ctx, request)
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
				dp := service.QueryParameter{
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
		response = endpoint.badQueryHandler(ctx, request)
		goto finish
	}

	if len(params) != 0 && !pipeline.Policy.AllowUnknownQueryParameters {
		response = endpoint.badQueryHandler(ctx, request)
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
		response = runHandler(handler, ctx, request, &err)
		if err != nil {
			err = merry.WithMessage(err, "running endpoint handler").WithValue("index", i)
			response = endpoint.iseHandler(ctx, request, err)
			goto finish
		}
		if response != nil {
			break
		}
	}
	pipelineTime = time.Now().UTC().Sub(pipelineStart)

	if response == nil {
		err = merry.New("no response from pipeline")
		response = endpoint.iseHandler(ctx, request, err)
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
	serializationTime := end.Sub(serializationStart)
	elapsed := end.Sub(start)

	if ce := g.Logger.Check(zap.InfoLevel, "request"); ce != nil {
		fs := make([]zapcore.Field, 14, 16)
		fs[0] = zap.String("request-id", requestID)
		fs[1] = zap.String("client-ip-address", request.ClientIP())
		fs[2] = zap.String("method", request.Method)
		fs[3] = zap.String("uri", request.URL.RequestURI())
		fs[4] = zap.Int("status-code", response.StatusCode())
		fs[5] = zap.Int("response-size", size)
		fs[6] = zap.String("user-agent", request.UserAgent())
		fs[7] = zap.Time("start", start)
		fs[8] = zap.Duration("elapsed", elapsed)
		fs[9] = zap.Duration("request-id-generation", requestIDGenerationTime)
		fs[10] = zap.Duration("find-endpoint", findPathTime)
		fs[11] = zap.Duration("pipline", pipelineTime)
		fs[12] = zap.Duration("serialization", serializationTime)
		fs[13] = zap.Duration("handlers", handlersTime)
		if endpoint != nil {
			fs = append(fs, zap.String("service", endpoint.serviceName))
		}
		if u := ctx.Actor(); u != nil {
			fs = append(fs, zap.String("user-id", u.ID()))
		}
		ce.Write(fs...)
	}

	if idErr != nil {
		g.ErrorHandler(ctx, request, idErr)
	}
	if err != nil {
		g.ErrorHandler(ctx, request, err)
	}
	writeErr1 := merry.WithMessage(writeErr, "serializing response")
	if writeErr1 != nil {
		g.ErrorHandler(ctx, request, writeErr1)
	}
	respErr := merry.WithMessage(response.Err(), "handler failed")
	if respErr != nil {
		g.ErrorHandler(ctx, request, respErr)
	}
}

func recovery(fatalError *merry.Error) {
	arg := recover()
	if arg == nil {
		return
	}

	if err, ok := arg.(error); ok {
		*fatalError = merry.Wrap(err)
		return
	}

	*fatalError = merry.New("panic in handler").WithValue("context", arg)
}

func runHandler(handler service.Handler, ctx context.Context, request *service.Request, err *merry.Error) service.Response {
	defer recovery(err)
	return handler(ctx, request)
}
