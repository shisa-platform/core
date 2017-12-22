package main

import (
	"fmt"
	"time"

	"github.com/percolate/shisa/authn"
	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/middleware"
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
	idp := &SimpleIdentityProvider{
		Users: []User{User{"1", "Boss", "password"}},
	}
	authenticator, err := authn.NewBasicAuthenticator(idp, "hello")
	if err != nil {
		panic(err)
	}
	authn := middleware.Authentication{Authenticator: authenticator}
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}

	svc := &HelloService{}
	svc.endpoints = []service.Endpoint{
		service.GetEndpointWithPolicy("/greeting", policy, svc.Greeting),
		service.GetEndpointWithPolicy("/salutation", policy, authn.Service, svc.Salutaion),
	}

	return svc
}

func (s *HelloService) Name() string {
	return "hello"
}

func (s *HelloService) Endpoints() []service.Endpoint {
	return s.endpoints
}

func (s *HelloService) Greeting(context.Context, *service.Request) service.Response {
	response := service.NewOK(Greeting{"hello, world"})
	addCommonHeaders(response)
	response.Trailers().Add("test", "foo")

	return response
}

func (s *HelloService) Salutaion(ctx context.Context, r *service.Request) service.Response {
	response := service.NewOK(Greeting{fmt.Sprintf("hello, %s", ctx.Actor().String())})
	addCommonHeaders(response)

	return response
}
