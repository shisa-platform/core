package main

import (
	"fmt"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
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
}

func NewHelloService() *HelloService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}

	svc := &HelloService{}

	language := service.Field{Name: "language"}
	greeting := service.GetEndpointWithPolicy("/greeting", policy, svc.Greeting)
	greeting.Get.QueryFields = []service.Field{
		language,
		{Name: "name"},
	}

	salutation := service.GetEndpointWithPolicy("/salutation", policy, svc.Salutaion)
	salutation.Get.QueryFields = []service.Field{language}

	svc.endpoints = []service.Endpoint{greeting, salutation}

	return svc
}

func (s *HelloService) Name() string {
	return "hello"
}

func (s *HelloService) Endpoints() []service.Endpoint {
	return s.endpoints
}

func (s *HelloService) Greeting(ctx context.Context, r *service.Request) (response service.Response) {
	if ctx.Actor() != nil {
		response = service.NewOK(Greeting{fmt.Sprintf("hello, %s", ctx.Actor().String())})
	} else {
		response = service.NewOK(Greeting{"hello, world"})
	}
	addCommonHeaders(response)
	response.Trailers().Add("test", "foo")

	return
}

func (s *HelloService) Salutaion(ctx context.Context, r *service.Request) service.Response {
	response := service.NewOK(Greeting{fmt.Sprintf("hello, %s", ctx.Actor().String())})
	addCommonHeaders(response)

	return response
}
