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
	Policy    service.Policy
	authn     middleware.Authenticator
	endpoints []service.Endpoint
}

func NewHelloService() *HelloService {
	idp := &SimpleIdentityProvider{
		Users: []User{User{"1", "Boss", "password"}},
	}
	provider, err := authn.NewBasicAuthenticationProvider(idp, "hello")
	if err != nil {
		panic(err)
	}
	svc := &HelloService{
		Policy: service.Policy{
			TimeBudget:                  time.Millisecond * 5,
			AllowTrailingSlashRedirects: true,
		},
		authn: middleware.Authenticator{Provider: provider},
	}
	svc.endpoints = []service.Endpoint{
		service.GetEndpointWithPolicy("/greeting", svc.Policy, svc.Greeting),
		service.GetEndpointWithPolicy("/salutation", svc.Policy, svc.authn.Service, svc.Salutaion),
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
