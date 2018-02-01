package main

import (
	"expvar"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"

	"go.uber.org/zap"
)

var hits = new(expvar.Map)

func init() {
	goodbye.Set("hits", hits)
}

type SimpleResponse string

func (r SimpleResponse) StatusCode() int {
	return http.StatusOK
}

func (r SimpleResponse) Headers() http.Header {
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

	var (
		response service.Response
		writeErr error
	)

	if req.Method != http.MethodGet {
		response = service.NewEmpty(http.StatusMethodNotAllowed)
		goto respond
	}

	switch req.URL.Path {
	case "/healthcheck":
		response = s.healthcheck(ctx, request)
		hits.Add(req.URL.Path, 1)
	case "/api/goodbye":
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
	writeErr = httpx.WriteResponse(ri, response)

finish:
	ri.Flush(ctx, request)
	if writeErr != nil {
		s.Logger.Error("error serializing response", zap.String("request-id", ctx.RequestID()), zap.Error(writeErr))
	}
}

func (s *Goodbye) healthcheck(ctx context.Context, r *service.Request) service.Response {
	return SimpleResponse("{\"ready\": true}")
}

func (s *Goodbye) goodbye(ctx context.Context, request *service.Request) service.Response {
	if err := request.ParseForm(); err != nil {
		return service.NewEmpty(http.StatusBadRequest)
	}

	who := "world"
	if name, ok := request.Form["name"]; ok {
		who = name[0]
	}

	return SimpleResponse(fmt.Sprintf("{\"goodbye\": %q}", who))
}
