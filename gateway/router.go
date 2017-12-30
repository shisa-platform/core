package gateway

import (
	stdctx "context"
	"net/http"
	"net/url"
	"sort"
	"strings"
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

type byName []service.QueryParameter

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type byOrdinal []service.QueryParameter

func (p byOrdinal) Len() int           { return len(p) }
func (p byOrdinal) Less(i, j int) bool { return p[i].Ordinal < p[j].Ordinal }
func (p byOrdinal) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()

	ctx := context.New(backgroundContext)
	request := &service.Request{Request: r}

	requestIDGenerationStart := time.Now().UTC()
	requestID, idErr := g.RequestIDGenerator(ctx, request)
	if idErr != nil {
		requestID = request.GenerateID()
	}
	if requestID == "" {
		idErr = merry.New("empty request id").WithUserMessage("Request ID Generator returned empty string")
		requestID = request.GenerateID()
	}
	requestIDGenerationTime := time.Now().UTC().Sub(requestIDGenerationStart)

	ctx = context.WithRequestID(ctx, requestID)

	parseQuery(&request.QueryParams, request.URL.RawQuery)
	sort.Sort(byOrdinal(request.QueryParams))

	var (
		err           merry.Error
		response      service.Response
		pipeline      *service.Pipeline
		pipelineStart time.Time
		pipelineTime  time.Duration
		invalidParams bool
		missingParams bool
		params        []service.QueryParameter
	)

	findPathStart := time.Now().UTC()
	path := request.URL.EscapedPath()
	endpoint, pathParams, tsr, err := g.tree.getValue(path)
	findPathTime := time.Now().UTC().Sub(findPathStart)

	if g.Authentication != nil {
		if response = g.Authentication.Service(ctx, request); response != nil {
			goto finish
		}
	}

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

	if len(pipeline.Fields) == 0 {
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

	for _, f := range pipeline.Fields {
		var found bool
		for j, p := range params {
			if p.Invalid {
				invalidParams = true
			}
			if f.Match(p.Name) {
				found = true
				request.QueryParams[p.Ordinal].Unknown = false
				params = append(params[:j], params[j+1:]...)
				if err := f.Validate(p.Values); err != nil {
					request.QueryParams[p.Ordinal].Invalid = true
					invalidParams = true
				}
				break
			}
		}

		if !found {
			if f.Default != "" {
				dp := service.QueryParameter{
					Name:    f.Name,
					Values:  []string{f.Default},
					Ordinal: -1,
				}
				request.QueryParams = append(request.QueryParams, dp)
			} else if f.Required {
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

	if idErr != nil {
		g.Logger.Warn("request id generator failed, fell back to default", zap.String("request-id", requestID), zap.Error(idErr))
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

func parseQuery(ps *[]service.QueryParameter, query string) {
	m := make(map[string]service.QueryParameter)
	i := 0

	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}

		key1, err1 := url.QueryUnescape(key)
		if err1 == nil {
			key = key1
		}
		value1, err1 := url.QueryUnescape(value)
		if err1 == nil {
			value = value1
		}

		p, found := m[key]
		if !found {
			p.Name = key
			p.Ordinal = i
			p.Unknown = true
		}
		p.Values = append(p.Values, value)
		if err1 != nil {
			p.Invalid = true
		}

		m[key] = p
		i++
	}

	*ps = make([]service.QueryParameter, 0, len(m))
	for _, v := range m {
		*ps = append(*ps, v)
	}
}
