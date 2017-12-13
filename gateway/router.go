package gateway

import (
	stdctx "context"
	"io"
	"net/http"
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

	// xxx - get request  from pool
	request := &service.Request{Request: r}

	requestID := g.RequestIDGenerator(request)
	if requestID == "" {
		g.Logger.Warn("request generator failed, falling back")
		requestID = request.GenerateID()
	}

	// xxx - fetch context from pool
	ctx := context.WithRequestID(backgroundContext, requestID)

	var response service.Response
	var pipeline *service.Pipeline

	path := request.URL.Path
	// xxx - use params
	// xxx - escape path ?
	// xxx - does getValue need 2 params?
	endpoint, _, tsr, err := g.tree.getValue(path, false)
	if err != nil {
		err.WithValue("request-id", requestID)
		err.WithValue("method", r.Method).WithValue("path", path).WithValue("route", endpoint.Route)
		// xxx - don't log here, wait until the end.  add a failure response with embedded cause (merry error)
		g.Logger.Error("configuration error", zap.Error(err))
		response = g.NotFoundHandler(ctx, request)
		goto end
	}

	if endpoint == nil {
		response = g.NotFoundHandler(ctx, request)
		goto end
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
		goto end
	}

	if tsr {
		if path != "/" && pipeline.Policy.AllowTrailingSlashRedirects {
			response = endpoint.redirectHandler(ctx, request)
		} else {
			response = g.NotFoundHandler(ctx, request)
		}
		goto end
	}

	if pipeline.Policy.TimeBudget != 0 {
		// xxx - watch for timeout and kill pipeline, return
		var cancel stdctx.CancelFunc
		ctx, cancel = ctx.WithTimeout(pipeline.Policy.TimeBudget)
		defer cancel()
	}

	for i, handler := range pipeline.Handlers {
		var handlerError merry.Error
		response = run(handler, ctx, request, &handlerError)
		if handlerError != nil {
			handlerError.WithValue("request-id", requestID).WithValue("pipline-index", i)
			handlerError.WithValue("method", request.Method).WithValue("path", path).WithValue("route", endpoint.Route)
			response = endpoint.iseHandler(ctx, request, handlerError)
			handlerError.WithHTTPCode(response.StatusCode())
			// xxx - log this to handler panic channel
			g.Logger.Error("internal error", zap.Error(handlerError))
			goto end
		}
		if response != nil {
			break
		}
	}

	if response == nil {
		err = merry.New("no response from pipeline").WithValue("request-id", requestID)
		err.WithValue("method", request.Method).WithValue("path", path).WithValue("route", endpoint.Route)
		response = endpoint.iseHandler(ctx, request, err)
		err.WithHTTPCode(response.StatusCode())
		// xxx - be better
		g.Logger.Error("internal error", zap.Error(err))
	}

end:
	for k, vs := range response.Headers() {
		w.Header()[k] = vs
	}
	for k := range response.Trailers() {
		w.Header().Add("Trailer", k)
	}

	w.Header().Set(g.RequestIDHeaderName, requestID)

	w.WriteHeader(response.StatusCode())
	// xxx - handle error here
	shim := countingWriter{delegate: w}
	response.Serialize(&shim)

	for k, vs := range response.Trailers() {
		w.Header()[k] = vs
	}

	end := time.Now().UTC()
	elapsed := end.Sub(start)

	if ce := g.requestLog.Check(zap.InfoLevel, "request"); ce != nil {
		fs := make([]zapcore.Field, 9, 11)
		fs[0] = zap.String("request-id", requestID)
		fs[1] = zap.String("client-ip-address", request.ClientIP())
		fs[2] = zap.Time("start-time", start)
		fs[3] = zap.String("method", request.Method)
		fs[4] = zap.String("uri", request.URL.RequestURI())
		fs[5] = zap.Int("status-code", response.StatusCode())
		fs[6] = zap.Uint64("response-size", shim.count)
		fs[7] = zap.Duration("elapsed-time", elapsed)
		fs[8] = zap.String("user-agent", request.UserAgent())
		if endpoint != nil {
			fs = append(fs, zap.String("service", endpoint.serviceName))
		}
		if u := ctx.Actor(); u != nil {
			fs = append(fs, zap.String("user-id", u.ID()))
		}
		ce.Write(fs...)
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

	*fatalError = merry.New("panic").WithValue("context", arg)
}

func run(handler service.Handler, ctx context.Context, request *service.Request, err *merry.Error) service.Response {
	defer recovery(err)
	return handler(ctx, request)
}

type countingWriter struct {
	count    uint64
	delegate io.Writer
}

func (w *countingWriter) Write(p []byte) (n int, err error) {
	n, err = w.delegate.Write(p)
	w.count += uint64(n)
	return
}
