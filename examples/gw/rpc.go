package main

import (
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"github.com/ansel1/merry"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/examples/rpc/service"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/lb"
	"github.com/shisa-platform/core/service"
)

var (
	language = httpx.ParameterSchema{
		Name:         "language",
		Default:      hello.AmericanEnglish,
		Validator:    httpx.StringSliceValidator{Target: hello.SupportedLanguages}.Validate,
		Multiplicity: 1,
	}
)

type Greeting struct {
	Message string
}

func (g Greeting) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"greeting\": %q}", g.Message)), nil
}

type RpcService struct {
	service.Service
	balancer lb.Balancer
}

func NewRpcService(bal lb.Balancer) *RpcService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 15,
		AllowTrailingSlashRedirects: true,
	}

	svc := &RpcService{
		Service: service.Service{
			Name: "rpc",
		},
		balancer: bal,
	}

	greeting := service.GetEndpointWithPolicy("/api/greeting", policy, svc.Greeting)
	greeting.Get.QuerySchemas = []httpx.ParameterSchema{
		language,
		{Name: "name", Multiplicity: 1},
	}

	svc.Endpoints = []service.Endpoint{greeting}

	return svc
}

func (s *RpcService) Name() string {
	return s.Service.Name
}

func (s *RpcService) Greeting(ctx context.Context, r *httpx.Request) httpx.Response {
	span := ctx.StartSpan("RpcService.GreetingEndpoint")
	defer span.Finish()

	client, err := s.connect()
	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
	}

	message := hello.Message{
		RequestID: ctx.RequestID(),
		UserID:    ctx.Actor().ID(),
		Metadata:  make(map[string]string),
	}

	for _, param := range r.QueryParams {
		switch param.Name {
		case "language":
			message.Language = param.Values[0]
		case "name":
			message.Name = param.Values[0]
		}
	}

	carrier := opentracing.TextMapCarrier(message.Metadata)
	if err := span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return httpx.NewEmptyError(http.StatusBadGateway, err)
	}

	var reply string
	if err := client.Call("Hello.Greeting", &message, &reply); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(otlog.String("error", err.Error()))
		return httpx.NewEmptyError(http.StatusBadGateway, err)
	}

	response := httpx.NewOK(Greeting{reply})
	addCommonHeaders(response)

	return response
}

func (s *RpcService) Healthcheck(ctx context.Context) merry.Error {
	_, err := s.balancer.Balance(s.Service.Name)
	if err != nil {
		return err.Prepend("healthcheck")
	}
	return nil
}

func (s *RpcService) connect() (*rpc.Client, merry.Error) {
	addr, err := s.balancer.Balance(s.Service.Name)
	if err != nil {
		return nil, err.Prepend("connect")
	}

	client, rpcErr := rpc.DialHTTP("tcp", addr)
	if rpcErr != nil {
		return nil, merry.Prepend(rpcErr, "connect")
	}

	return client, nil
}
