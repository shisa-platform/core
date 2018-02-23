package main

import (
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/examples/rpc/service"
	"github.com/percolate/shisa/sd"
	"github.com/percolate/shisa/service"
)

var (
	language = service.Field{
		Name:         "language",
		Default:      hello.AmericanEnglish,
		Validator:    service.StringSliceValidator{Target: hello.SupportedLanguages}.Validate,
		Multiplicity: 1,
	}
)

type Greeting struct {
	Message string
}

func (g Greeting) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"greeting\": %q}", g.Message)), nil
}

type HelloService struct {
	service.ServiceAdapter
	env       env.Provider
	endpoints []service.Endpoint
	resolver  sd.Resolver
}

func NewHelloService(environment env.Provider, res sd.Resolver) *HelloService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}

	svc := &HelloService{
		env:      environment,
		resolver: res,
	}

	greeting := service.GetEndpointWithPolicy("/api/greeting", policy, svc.Greeting)
	greeting.Get.QueryFields = []service.Field{
		language,
		{Name: "name", Multiplicity: 1},
	}

	svc.endpoints = []service.Endpoint{greeting}

	return svc
}

func (s *HelloService) Name() string {
	return "hello"
}

func (s *HelloService) Endpoints() []service.Endpoint {
	return s.endpoints
}

func (s *HelloService) Greeting(ctx context.Context, r *service.Request) service.Response {
	client, err := s.connect()
	if err != nil {
		return service.NewEmptyError(http.StatusInternalServerError, err)
	}

	message := hello.Message{
		RequestID: ctx.RequestID(),
		UserID:    ctx.Actor().ID(),
	}

	for _, param := range r.QueryParams {
		switch param.Name {
		case "language":
			message.Language = param.Values[0]
		case "name":
			message.Name = param.Values[0]
		}
	}

	var reply string
	rpcErr := client.Call("Hello.Greeting", &message, &reply)
	if rpcErr != nil {
		return service.NewEmptyError(http.StatusInternalServerError, rpcErr)
	}

	response := service.NewOK(Greeting{reply})
	addCommonHeaders(response)

	return response
}

func (s *HelloService) Healthcheck(ctx context.Context) merry.Error {
	addrs, err := s.resolver.Resolve(s.Name())
	if err != nil {
		return err.WithUserMessage("service registry not found")
	}

	if len(addrs) < 1 {
		return merry.New("no healthy hosts")
	}

	return nil
}

func (s *HelloService) connect() (*rpc.Client, merry.Error) {
	addrs, err := s.resolver.Resolve(s.Name())
	if err != nil {
		return nil, err.WithUserMessage("service registry not found")
	}

	if len(addrs) < 1 {
		return nil, merry.New("no healthy hosts")
	}

	client, rpcErr := rpc.DialHTTP("tcp", addrs[0])
	if rpcErr != nil {
		return nil, merry.Wrap(rpcErr).WithUserMessage("unable to connect")
	}

	return client, nil
}
