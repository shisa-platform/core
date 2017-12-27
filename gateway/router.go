package gateway

import (
	stdctx "context"
	"net/http"
	"net/url"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

var (
	backgroundContext = stdctx.Background()
)

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()

	request := &service.Request{Request: r}

	requestIDGenerationStart := time.Now().UTC()
	requestID := g.RequestIDGenerator(request)
	if requestID == "" {
		g.Logger.Warn("request generator failed, falling back")
		requestID = request.GenerateID()
	}
	requestIDGenerationTime := time.Now().UTC().Sub(requestIDGenerationStart)

	ctx := context.WithRequestID(backgroundContext, requestID)

	var (
		err           merry.Error
		response      service.Response
		pipeline      *service.Pipeline
		pipelineStart time.Time
		pipelineTime  time.Duration
	)

	findPathStart := time.Now().UTC()
	path := request.URL.EscapedPath()
	endpoint, pathParams, tsr, err := g.tree.getValue(path)
	findPathTime := time.Now().UTC().Sub(findPathStart)

	if err != nil {
		response = defaultInternalServerErrorHandler(ctx, request, err)
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

	request.PathParams = pathParams
	if !pipeline.Policy.PreserveEscapedPathParameters {
		for i := range request.PathParams {
			if esc, r := url.PathUnescape(request.PathParams[i].Value); r == nil {
				request.PathParams[i].Value = esc
			}
		}
	}

	if qp, pe := url.ParseQuery(request.URL.RawQuery); pe != nil && !pipeline.Policy.AllowMalformedQueryParameters {
		response = endpoint.queryParamHandler(ctx, request)
		goto finish
	} else {
		request.QueryParams = qp
	}

	if pipeline.Policy.TimeBudget != 0 {
		// xxx - watch for timeout and kill pipeline, return
		var cancel stdctx.CancelFunc
		ctx, cancel = ctx.WithTimeout(pipeline.Policy.TimeBudget)
		defer cancel()
	}

	pipelineStart = time.Now().UTC()
	for _, handler := range pipeline.Handlers {
		response = run(handler, ctx, request, &err)
		if err != nil {
			response = endpoint.iseHandler(ctx, request, err)
			goto finish
		}
		if response != nil {
			break
		}
	}
	pipelineTime = time.Now().UTC().Sub(pipelineStart)

	if response == nil {
		err = merry.New("internal error").WithUserMessage("no response from pipeline")
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

	if ce := g.requestLog.Check(zap.InfoLevel, "request"); ce != nil {
		fs := make([]zapcore.Field, 13, 15)
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
		if endpoint != nil {
			fs = append(fs, zap.String("service", endpoint.serviceName))
		}
		if u := ctx.Actor(); u != nil {
			fs = append(fs, zap.String("user-id", u.ID()))
		}
		ce.Write(fs...)
	}
	if err != nil {
		g.Logger.Error(merry.UserMessage(err), zap.String("request-id", requestID), zap.Error(err))
	}
	if writeErr != nil {
		g.Logger.Error("error serializing response", zap.String("request-id", requestID), zap.Error(writeErr))
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

func run(handler service.Handler, ctx context.Context, request *service.Request, err *merry.Error) service.Response {
	defer recovery(err)
	return handler(ctx, request)
}
