package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"time"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/examples/idp/service"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/lb"
)

const idpServiceName = "idp"

var hits = new(expvar.Map)

func init() {
	goodbye.Set("hits", hits)
}

type Healthcheck struct {
	Ready   bool              `json:"ready"`
	Details map[string]string `json:"details"`
	Error   error             `json:"-"`
}

func (r Healthcheck) StatusCode() int {
	if r.Ready {
		return http.StatusOK
	}

	return http.StatusServiceUnavailable
}

func (r Healthcheck) Headers() http.Header {
	now := time.Now().UTC().Format(time.RFC1123)
	headers := make(http.Header)

	headers.Set("Content-Type", "application/json")

	headers.Set("Cache-Control", "private, max-age=0")
	headers.Set("Date", now)
	headers.Set("Expires", now)
	headers.Set("Last-Modified", now)
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")
	headers.Set("X-Xss-Protection", "1; mode=block")

	return headers
}

func (r Healthcheck) Trailers() http.Header {
	return nil
}

func (r Healthcheck) Err() error {
	return r.Error
}

func (r Healthcheck) Serialize(w io.Writer) merry.Error {
	p, err := json.Marshal(r)
	if err != nil {
		goto done
	}
	_, err = w.Write(p)
done:
	return merry.Prepend(err, "marshaling healthcheck")
}

type SimpleResponse string

func (r SimpleResponse) StatusCode() int {
	return http.StatusOK
}

func (r SimpleResponse) Headers() http.Header {
	headers := make(http.Header)

	headers.Set("Content-Type", "application/json")
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")

	return headers
}

func (r SimpleResponse) Trailers() http.Header {
	return nil
}

func (r SimpleResponse) Err() error {
	return nil
}

func (r SimpleResponse) Serialize(w io.Writer) merry.Error {
	_, err := fmt.Fprint(w, string(r))
	return merry.Prepend(err, "writing response")
}

type Goodbye struct {
	Balancer lb.Balancer
	Logger   *zap.Logger
}

func (s *Goodbye) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ri := httpx.NewInterceptor(w)

	ctx := context.New(req.Context())

	request := &httpx.Request{Request: req}
	request.ParseQueryParameters()

	var requestID string
	if values, exists := request.Header["X-Request-Id"]; exists {
		requestID = values[0]
	} else {
		requestID = request.ID()
	}

	ctx = ctx.WithRequestID(requestID)

	var response httpx.Response
	logRequest := false

	if req.Method != http.MethodGet {
		response = httpx.NewEmpty(http.StatusMethodNotAllowed)
		goto respond
	}

	switch req.URL.Path {
	case "/healthcheck":
		response = s.healthcheck(ctx, request)
		hits.Add(req.URL.Path, 1)
	case "/goodbye":
		span := s.startSpan(ctx, request, ri.Start())
		defer span.Finish()

		response = s.goodbye(ctx, request)

		if err := response.Err(); err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.String("error", err.Error()))
		}

		hits.Add(req.URL.Path, 1)
		logRequest = true
	case "/debug/vars":
		expvar.Handler().ServeHTTP(ri, req)
		hits.Add(req.URL.Path, 1)
		goto end
	default:
		response = httpx.NewEmpty(http.StatusNotFound)
	}

respond:
	if err := ri.WriteResponse(response); err != nil {
		s.Logger.Error("error serializing response", zap.String("request-id", requestID), zap.Error(err))
	}

	ri.Flush()

	if logRequest {
		snapshot := ri.Snapshot()

		fs := make([]zapcore.Field, 9, 10)
		fs[0] = zap.String("request-id", ctx.RequestID())
		fs[1] = zap.String("client-ip-address", request.ClientIP())
		fs[2] = zap.String("method", request.Method)
		fs[3] = zap.String("uri", request.URL.RequestURI())
		fs[4] = zap.Int("status-code", snapshot.StatusCode)
		fs[5] = zap.Int("response-size", snapshot.Size)
		fs[6] = zap.String("user-agent", request.UserAgent())
		fs[7] = zap.Time("start", snapshot.Start)
		fs[8] = zap.Duration("elapsed", snapshot.Elapsed)
		if values, exists := request.Header["X-User-Id"]; exists {
			fs = append(fs, zap.String("user-id", values[0]))
		}
		s.Logger.Info("request", fs...)

		if response != nil && response.Err() != nil {
			values := merry.Values(response.Err())
			fs := make([]zapcore.Field, 0, len(values)+1)
			fs = append(fs, zap.String("request-id", requestID))
			for name, value := range values {
				if key, ok := name.(string); ok {
					if key == "stack" || key == "http status code" || key == "message" {
						continue
					}
					fs = append(fs, zap.Reflect(key, value))
				}
			}
			s.Logger.Error(response.Err().Error(), fs...)
		}
	}

end:
}

func (s *Goodbye) goodbye(ctx context.Context, request *httpx.Request) httpx.Response {
	var userID string
	if values, exists := request.Header["X-User-Id"]; exists {
		userID = values[0]
	} else {
		return httpx.NewEmptyError(http.StatusBadRequest, merry.New("missing user id"))
	}

	client, err := s.connect()
	if err != nil {
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	message := idp.Message{
		RequestID: ctx.RequestID(),
		Value:     userID,
		Metadata:  make(map[string]string),
	}

	span := ctx.Span()
	carrier := opentracing.TextMapCarrier(message.Metadata)
	if err = span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	var user idp.User
	if err := client.Call("Idp.FindUser", &message, &user); err != nil {
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}
	if user.Ident == "" {
		return httpx.NewEmpty(http.StatusUnauthorized)
	}

	who := user.Name
	if len(request.QueryParams) == 1 && request.QueryParams[0].Name == "name" {
		who = request.QueryParams[0].Values[0]
	}

	return SimpleResponse(fmt.Sprintf("{\"goodbye\": %q}", who))
}

func (s *Goodbye) healthcheck(ctx context.Context, r *httpx.Request) httpx.Response {
	response := Healthcheck{
		Ready:   true,
		Details: map[string]string{"idp": "OK"},
	}

	client, err := s.connect()
	if err != nil {
		response.Error = err
		response.Ready = false
		response.Details["idp"] = err.Error()
		return response
	}

	var ready bool
	arg := ctx.RequestID()
	rpcErr := client.Call("Idp.Healthcheck", &arg, &ready)
	if rpcErr != nil {
		response.Error = rpcErr
		response.Ready = false
		response.Details["idp"] = rpcErr.Error()
	}
	if !ready {
		response.Ready = false
		response.Details["idp"] = "not ready"
	}

	return response
}

func (s *Goodbye) connect() (*rpc.Client, error) {
	node, resErr := s.Balancer.Balance(idpServiceName)
	if resErr != nil {
		return nil, resErr
	}

	client, rpcErr := rpc.DialHTTP("tcp", node)
	if rpcErr != nil {
		return nil, rpcErr
	}

	return client, nil
}

func (s *Goodbye) startSpan(ctx context.Context, request *httpx.Request, start time.Time) opentracing.Span {
	var span opentracing.Span
	opts := []opentracing.StartSpanOption{opentracing.StartTime(start)}

	carrier := opentracing.HTTPHeadersCarrier(request.Header)
	if spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier); err == nil {
		opts = append(opts, ext.RPCServerOption(spanContext))
	} else {
		s.Logger.Error("error extracting client trace", zap.String("request-id", ctx.RequestID()), zap.String("error", err.Error()))
		opts = append(opts, opentracing.Tag{string(ext.SpanKind), "server"})
	}

	span = opentracing.StartSpan("Goodbye", opts...)
	span.SetTag("request_id", ctx.RequestID())

	ctx = ctx.WithSpan(span)

	return span
}
