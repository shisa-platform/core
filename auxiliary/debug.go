package auxiliary

import (
	stdctx "context"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/service"
)

const (
	defaultDebugServerPath = "/debug/vars"
)

var (
	debugStats   = new(expvar.Map)
	debuResponse = expvarResponse{
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

	// Logger optionally specifies the logger to use by the Debug
	// server.
	// If nil all logging is disabled.
	Logger *zap.Logger
}

func (s *DebugServer) init() {
	now := time.Now().UTC().Format(startTimeFormat)
	s.HTTPServer.init()
	debugStats = debugStats.Init()
	AuxiliaryStats.Set("debug", debugStats)
	debugStats.Set("hits", new(expvar.Int))
	startTime := new(expvar.String)
	startTime.Set(now)
	debugStats.Set("starttime", startTime)

	if s.Path == "" {
		s.Path = defaultDebugServerPath
	}

	s.Router = s.Route

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
	defer s.Logger.Sync()
}

func (s *DebugServer) Name() string {
	return "debug"
}

func (s *DebugServer) Serve() error {
	s.init()

	listener, err := httpx.HTTPListenerForAddress(s.Addr)
	if err != nil {
		return err
	}

	addr := listener.Addr().String()
	addrVar := new(expvar.String)
	addrVar.Set(addr)
	debugStats.Set("addr", addrVar)
	s.Logger.Info("debug service started", zap.String("addr", addr))

	if s.UseTLS {
		return s.base.ServeTLS(listener, "", "")
	}

	return s.base.Serve(listener)
}

func (s *DebugServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(stdctx.Background(), timeout)
	defer cancel()
	return merry.Wrap(s.base.Shutdown(ctx))
}

func (s *DebugServer) Route(ctx context.Context, request *service.Request) service.Handler {
	if request.URL.Path == s.Path {
		return s.Service
	}

	return nil
}

func (s *DebugServer) Service(ctx context.Context, request *service.Request) service.Response {
	debugStats.Add("hits", 1)

	return debuResponse
}
