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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/examples/idp/server"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

const idpServiceAddrEnv = "IDP_ADDR"

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

func (r Healthcheck) Serialize(w io.Writer) (int, error) {
	p, err := json.Marshal(r)
	if err != nil {
		return 0, nil
	}
	return w.Write(p)
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

func (r SimpleResponse) Serialize(w io.Writer) (int, error) {
	return fmt.Fprint(w, string(r))
}

type Goodbye struct {
	Logger *zap.Logger
}

func (s *Goodbye) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ri := httpx.NewInterceptor(w, s.Logger)

	ctx := context.New(req.Context())
	request := &service.Request{Request: req}
	request.ParseQueryParameters()

	var requestID string
	if values, exists := request.Header["X-Request-Id"]; exists {
		requestID = values[0]
	} else {
		requestID = request.ID()
		s.Logger.Warn("missing upstream request id", zap.String("request-id", requestID))
	}

	ctx = context.WithRequestID(ctx, requestID)

	var response service.Response

	if req.Method != http.MethodGet {
		response = service.NewEmpty(http.StatusMethodNotAllowed)
		goto respond
	}

	switch req.URL.Path {
	case "/healthcheck":
		response = s.healthcheck(ctx, request)
		hits.Add(req.URL.Path, 1)
	case "/goodbye":
		response = s.goodbye(ctx, request)
		hits.Add(req.URL.Path, 1)
	case "/debug/vars":
		expvar.Handler().ServeHTTP(ri, req)
		hits.Add(req.URL.Path, 1)
		goto finish
	default:
		response = service.NewEmpty(http.StatusNotFound)
	}

respond:
	if err := httpx.WriteResponse(ri, response); err != nil {
		s.Logger.Error("error serializing response", zap.String("request-id", requestID), zap.Error(err))
	}

finish:
	ri.Flush(ctx, request)

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

func (s *Goodbye) healthcheck(ctx context.Context, r *service.Request) service.Response {
	response := Healthcheck{
		Ready:   true,
		Details: map[string]string{"idp": "OK"},
	}

	client, err := connect()
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

func (s *Goodbye) goodbye(ctx context.Context, request *service.Request) service.Response {
	var userID string
	if values, exists := request.Header["X-User-Id"]; exists {
		userID = values[0]
	} else {
		return service.NewEmptyError(http.StatusBadRequest, merry.New("missing user id"))
	}

	client, err := connect()
	if err != nil {
		return service.NewEmptyError(http.StatusInternalServerError, err)
	}

	message := idp.Message{RequestID: ctx.RequestID(), Value: userID}
	var user idp.User
	rpcErr := client.Call("Idp.FindUser", &message, &user)
	if rpcErr != nil {
		return service.NewEmptyError(http.StatusInternalServerError, rpcErr)
	}
	if user.Ident == "" {
		return service.NewEmpty(http.StatusUnauthorized)
	}

	who := user.Name
	if len(request.QueryParams) == 1 && request.QueryParams[0].Name == "name" {
		who = request.QueryParams[0].Values[0]
	}

	return SimpleResponse(fmt.Sprintf("{\"goodbye\": %q}", who))
}

func connect() (*rpc.Client, error) {
	addr, envErr := env.Get(idpServiceAddrEnv)
	if envErr != nil {
		return nil, envErr
	}

	client, rpcErr := rpc.DialHTTP("tcp", addr)
	if rpcErr != nil {
		return nil, rpcErr
	}

	return client, nil
}
