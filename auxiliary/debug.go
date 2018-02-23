package auxiliary

import (
	"expvar"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

const (
	defaultDebugServerPath = "/debug/vars"
)

var (
	debugStats    = new(expvar.Map)
	debugResponse = expvarResponse{
		headers: map[string][]string{
			"Content-Type": []string{"application/json"},
		},
	}
)

type expvarResponse struct {
	headers http.Header
}

func (r expvarResponse) StatusCode() int {
	return http.StatusOK
}

func (r expvarResponse) Headers() http.Header {
	return r.headers
}

func (r expvarResponse) Trailers() http.Header {
	return nil
}

func (r expvarResponse) Err() error {
	return nil
}

func (r expvarResponse) Serialize(w io.Writer) (size int, err error) {
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			n, _ := fmt.Fprintf(w, ",\n")
			size += n
		}
		first = false
		n, _ := fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
		size += n
	})
	n, _ := fmt.Fprintf(w, "\n}\n")
	size += n

	return
}

// DebugServer serves values from the `expvar` package to the
// configured address and path.
type DebugServer struct {
	HTTPServer
	Path string // URL path to listen on, "/debug/vars" if empty
}

func (s *DebugServer) init() {
	now := time.Now().UTC().Format(startTimeFormat)

	debugStats = debugStats.Init()

	AuxiliaryStats.Set("debug", debugStats)

	debugStats.Set("hits", new(expvar.Int))

	startTime := new(expvar.String)
	startTime.Set(now)
	debugStats.Set("starttime", startTime)

	debugStats.Set("addr", expvar.Func(func() interface{} {
		return s.Address()
	}))

	if s.Path == "" {
		s.Path = defaultDebugServerPath
	}

	s.Router = s.Route
}

func (s *DebugServer) Name() string {
	return "debug"
}

func (s *DebugServer) Serve() error {
	s.init()

	return s.HTTPServer.Serve()
}

func (s *DebugServer) Route(ctx context.Context, request *service.Request) service.Handler {
	if request.URL.Path == s.Path {
		return s.Service
	}

	return nil
}

func (s *DebugServer) Service(ctx context.Context, request *service.Request) service.Response {
	debugStats.Add("hits", 1)

	return debugResponse
}