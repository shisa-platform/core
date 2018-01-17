package auxiliary

import (
	stdctx "context"
	"encoding/json"
	"expvar"
	"net/http"
	"time"

	"github.com/ansel1/merry"
	"go.uber.org/zap"

	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

const (
	defaultHealthcheckServerPath = "/healthcheck"
)

var (
	healthcheckStats = new(expvar.Map)
	jsonContentType  = contenttype.ApplicationJson.String()
)

type statusMarshaler struct {
	status map[string]string
}

func (m statusMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.status)
}

// Healthchecker is a named resource that can report its status.
// If the resource is unable to perform its function it should
// return an error with the UserMessage explaining the problem.
type Healthchecker interface {
	// Name is the resource's name for external reporting
	Name() string
	// Healthcheck should return an error if the resource has a
	// problem and should not be considered reliable for further
	// requests.
	Healthcheck() merry.Error
}

// HealthcheckServer serves JSON status reports of all configured
// `Healthchecker` resources to the given address and path.  If
// all resources return no error a 200 status code is returned,
// otherwise a 503 is returned.
// The status report is a JSON Resource with the `Name()` of the
// resources as keys and either the `merry.UserMessage` of the
// error or `"OK"` as the value.
type HealthcheckServer struct {
	HTTPServer
	Path string // URL path to listen on, "/healthcheck" if empty

	// Checkers are the resources to include in the status report.
	Checkers []Healthchecker

	// Logger optionally specifies the logger to use by the
	// Healthcheck server.
	// If nil all logging is disabled.
	Logger *zap.Logger

	requestLog *zap.Logger
}

func (s *HealthcheckServer) init() {
	now := time.Now().UTC().Format(startTimeFormat)

	s.HTTPServer.init()

	healthcheckStats = healthcheckStats.Init()
	AuxiliaryStats.Set("healthcheck", healthcheckStats)
	healthcheckStats.Set("hits", new(expvar.Int))
	startTime := new(expvar.String)
	startTime.Set(now)
	healthcheckStats.Set("starttime", startTime)

	if s.Path == "" {
		s.Path = defaultHealthcheckServerPath
	}

	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
	defer s.Logger.Sync()
	s.requestLog = s.Logger.Named("request")

	s.base.Handler = s
}

func (s *HealthcheckServer) Name() string {
	return "healthcheck"
}

func (s *HealthcheckServer) Serve() error {
	s.init()

	if s.UseTLS {
		return s.base.ListenAndServeTLS("", "")
	}

	return s.base.ListenAndServe()
}

func (s *HealthcheckServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(stdctx.Background(), timeout)
	defer cancel()
	return merry.Wrap(s.base.Shutdown(ctx))
}

func (s *HealthcheckServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	healthcheckStats.Add("hits", 1)

	ri := ResponseInterceptor{
		Logger:   s.requestLog,
		Delegate: w,
		Start:    time.Now().UTC(),
	}

	ctx := context.New(r.Context())
	request := &service.Request{Request: r}

	requestID, idErr := s.RequestIDGenerator(ctx, request)
	if idErr != nil {
		requestID = request.ID()
	}
	if requestID == "" {
		idErr = merry.New("empty request id").WithUserMessage("Request ID Generator returned empty string")
		requestID = request.ID()
	}

	ctx = context.WithRequestID(ctx, requestID)
	ri.Header().Set(s.RequestIDHeaderName, requestID)

	code := http.StatusOK
	status := make(map[string]string, len(s.Checkers))

	var response service.Response
	if response = s.Authenticate(ctx, request); response != nil {
		goto finish
	}

	if s.Path != r.URL.Path {
		response = service.NewEmpty(http.StatusNotFound)
		response.Headers().Set("Content-Type", "text/plain; charset=utf-8")
		goto finish
	}

	for _, check := range s.Checkers {
		if err := check.Healthcheck(); err != nil {
			status[check.Name()] = merry.UserMessage(err)
			code = http.StatusServiceUnavailable
			continue
		}
		status[check.Name()] = "OK"
	}
	response = &service.JsonResponse{
		BasicResponse: service.BasicResponse{
			Code: code,
		},
		Payload: statusMarshaler{status},
	}
	response.Headers().Set(contenttype.ContentTypeHeaderKey, jsonContentType)

finish:
	writeErr := writeResponse(&ri, response)
	ri.Flush(ctx, request)

	if idErr != nil {
		s.Logger.Warn("request id generator failed, fell back to default", zap.String("request-id", requestID), zap.Error(idErr))
	}
	if writeErr != nil {
		s.Logger.Error("error serializing response", zap.String("request-id", requestID), zap.Error(writeErr))
	}
}
