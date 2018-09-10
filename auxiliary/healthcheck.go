package auxiliary

import (
	"encoding/json"
	"expvar"
	"net/http"
	"time"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/contenttype"
	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
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
	Healthcheck(context.Context) merry.Error
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
}

func (s *HealthcheckServer) init() {
	now := time.Now().UTC().Format(startTimeFormat)

	healthcheckStats = healthcheckStats.Init()

	AuxiliaryStats.Set("healthcheck", healthcheckStats)

	healthcheckStats.Set("hits", new(expvar.Int))

	startTime := new(expvar.String)
	startTime.Set(now)
	healthcheckStats.Set("starttime", startTime)

	healthcheckStats.Set("addr", expvar.Func(func() interface{} {
		return s.Address()
	}))

	if s.Path == "" {
		s.Path = defaultHealthcheckServerPath
	}

	s.Router = s.Route
}

func (s *HealthcheckServer) Name() string {
	return "healthcheck"
}

func (s *HealthcheckServer) Route(ctx context.Context, request *httpx.Request) httpx.Handler {
	if request.URL.Path == s.Path {
		return s.Service
	}

	return nil
}

func (s *HealthcheckServer) Listen() error {
	if err := s.HTTPServer.Listen(); err != nil {
		return err
	}

	s.init()

	return nil
}

func (s *HealthcheckServer) Service(ctx context.Context, request *httpx.Request) httpx.Response {
	healthcheckStats.Add("hits", 1)

	code := http.StatusOK
	status := make(map[string]string, len(s.Checkers))

	for _, check := range s.Checkers {
		if err := invokeHealthcheckSafely(ctx, check); err != nil {
			status[check.Name()] = err.Error()
			code = http.StatusServiceUnavailable
			err1 := err.Prepend(check.Name()).Prepend("healthcheck")
			s.invokeErrorHookSafely(ctx, request, err1)
			continue
		}
		status[check.Name()] = "OK"
	}
	response := &httpx.JsonResponse{
		BasicResponse: httpx.BasicResponse{
			Code: code,
		},
		Payload: statusMarshaler{status},
	}
	response.Headers().Set(contenttype.ContentTypeHeaderKey, jsonContentType)

	return response
}

func invokeHealthcheckSafely(ctx context.Context, h Healthchecker) (err merry.Error) {
	defer errorx.CapturePanic(&err, "panic in healthcheck")

	return h.Healthcheck(ctx)
}
