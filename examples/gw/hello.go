package main

import (
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/examples/rpc/service"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/lb"
	"github.com/percolate/shisa/service"
)

var (
	language = httpx.Field{
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

type HelloService struct {
	service.ServiceAdapter
	endpoints []service.Endpoint
	balancer  lb.Balancer
}

func NewHelloService(bal lb.Balancer) *HelloService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 15,
		AllowTrailingSlashRedirects: true,
	}

	svc := &HelloService{
		balancer: bal,
	}

	greeting := service.GetEndpointWithPolicy("/api/greeting", policy, svc.Greeting)
	greeting.Get.QueryFields = []httpx.Field{
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

func (s *HelloService) Greeting(ctx context.Context, r *httpx.Request) httpx.Response {
	client, err := s.connect()
	if err != nil {
		return httpx.NewEmptyError(http.StatusInternalServerError, err)
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
		return httpx.NewEmptyError(http.StatusInternalServerError, rpcErr)
	}

	response := httpx.NewOK(Greeting{reply})
	addCommonHeaders(response)

	return response
}

func (s *HelloService) Healthcheck(ctx context.Context) merry.Error {
	_, err := s.balancer.Balance(s.Name())
	if err != nil {
		return err.Prepend("healthcheck")
	}
	return nil
}

func (s *HelloService) connect() (*rpc.Client, merry.Error) {
	addr, err := s.balancer.Balance(s.Name())
	if err != nil {
		return nil, err.Prepend("connect")
	}

	fmt.Println(addr)

	client, rpcErr := rpc.DialHTTP("tcp", addr)
	if rpcErr != nil {
		return nil, merry.Wrap(rpcErr).WithUserMessage("unable to connect")
	}

	return client, nil
}
