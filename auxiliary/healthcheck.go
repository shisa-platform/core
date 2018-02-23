package auxiliary

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/contenttype"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/sd"
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

	// Registrar implements sd.Healthchecker and registers
	// the healthcheck endpoint with a service registry
	Registrar sd.Healthchecker

	// RegistryURL is a function that configures the URL that the
	// Registrar will use for healthchecks. By default, uses the Address
	// method and Path field, and http or https scheme based on the value
	// of the UseTLS field
	RegistryURLHook func() (*url.URL, error)

	ServiceName string
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

	if s.ServiceName == "" {
		s.ServiceName = "gateway"
	}

	s.Router = s.Route
}

func (s *HealthcheckServer) Name() string {
	return "healthcheck"
}

func (s *HealthcheckServer) Route(ctx context.Context, request *service.Request) service.Handler {
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

	if s.Registrar != nil {
		if s.RegistryURLHook == nil {
			s.RegistryURLHook = s.defaultRegistryURLHook
		}

		if err := s.register(); err != nil {
			return err
		}
	}

	return nil
}

func (s *HealthcheckServer) register() error {
	u, err := s.RegistryURLHook()
	if err != nil {
		return err
	}

	return s.Registrar.AddCheck(s.ServiceName, u)
}

func (s *HealthcheckServer) defaultRegistryURLHook() (*url.URL, error) {
	var scheme string

	if s.UseTLS {
		scheme = "https"
	} else {
		scheme = "http"
	}

	surl := fmt.Sprintf("%s://%s%s", scheme, s.Address(), s.Path)

	return url.Parse(surl)
}

func (s *HealthcheckServer) Service(ctx context.Context, request *service.Request) service.Response {
	healthcheckStats.Add("hits", 1)

	code := http.StatusOK
	status := make(map[string]string, len(s.Checkers))

	for _, check := range s.Checkers {
		if err := check.Healthcheck(ctx); err != nil {
			status[check.Name()] = err.Error()
			code = http.StatusServiceUnavailable
			if s.ErrorHook != nil {
				err1 := merry.WithMessage(err, "check failed").WithValue("name", check.Name())
				s.ErrorHook(ctx, request, err1)
			}
			continue
		}
		status[check.Name()] = "OK"
	}
	response := &service.JsonResponse{
		BasicResponse: service.BasicResponse{
			Code: code,
		},
		Payload: statusMarshaler{status},
	}
	response.Headers().Set(contenttype.ContentTypeHeaderKey, jsonContentType)

	return response
}
