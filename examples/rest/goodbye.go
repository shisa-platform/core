package main

import (
	"fmt"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

type Farewell struct {
	Message string
}

func (g Farewell) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"farewell\": %q}", g.Message)), nil
}

type GoodbyeService struct {
	service.ServiceAdapter
	Policy    service.Policy
	endpoints []service.Endpoint
}

func NewGoodbyeService() *GoodbyeService {
	svc := &GoodbyeService{
		Policy: service.Policy{
			TimeBudget:                  time.Millisecond * 5,
			AllowTrailingSlashRedirects: true,
		},
	}

	svc.endpoints = []service.Endpoint{
		service.GetEndpointWithPolicy("/farewell", svc.Policy, svc.Farewell),
	}

	return svc
}

func (s *GoodbyeService) Name() string {
	return "goodbye"
}

func (s *GoodbyeService) Endpoints() []service.Endpoint {
	return s.endpoints
}

func (s *GoodbyeService) Farewell(context.Context, *service.Request) service.Response {
	response := service.NewOK(Farewell{"goodbye, world"})
	addCommonHeaders(response)

	return response
}
