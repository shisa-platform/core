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
	"github.com/percolate/shisa/service"
)

var (
	language = service.Field{
		Name:         "language",
		Default:      hello.AmericanEnglish,
		Validator:    service.StringSliceValidator{hello.SupportedLanguages}.Validate,
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
}

func NewHelloService(environment env.Provider) *HelloService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}

	svc := &HelloService{
		env: environment,
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
	client, err := s.connect()
	if err != nil {
		return err
	}

	var ready bool
	arg := ctx.RequestID()
	rpcErr := client.Call("Hello.Healthcheck", &arg, &ready)
	if rpcErr != nil {
		return merry.Wrap(rpcErr).WithUserMessage("unable to complete request")
	}
	if !ready {
		return merry.New("not ready").WithUserMessage("not ready")
	}

	return nil
}

func (s *HelloService) connect() (*rpc.Client, merry.Error) {
	addr, envErr := env.Get(helloServiceAddrEnv)
	if envErr != nil {
		return nil, envErr.WithUserMessage("address environment variable not found")
	}

	client, rpcErr := rpc.DialHTTP("tcp", addr)
	if rpcErr != nil {
		return nil, merry.Wrap(rpcErr).WithUserMessage("unable to connect")
	}

	return client, nil
}
