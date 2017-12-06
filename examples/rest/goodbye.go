package main

import (
	"fmt"

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
}

func (s *GoodbyeService) Name() string {
	return "goodbye"
}

func (s *GoodbyeService) Endpoints() []service.Endpoint {
	return []service.Endpoint{
		service.Endpoint{
			Service: s,
			Route:   "/farewell",
			Get: &service.Pipeline{
				Policy:   commonPolicy,
				Handlers: []service.Handler{s.Farewell},
			},
		},
	}
}

func (s *GoodbyeService) Handlers() []service.Handler {
	return nil
}

func (s *GoodbyeService) MethodNotAllowedHandler() service.Handler {
	return nil
}

func (s *GoodbyeService) InternalServerErrorHandler() service.ErrorHandler {
	return nil
}

func (s *GoodbyeService) Farewell(context.Context, *service.Request) service.Response {
	response := service.NewOK(Farewell{"goodbye, world"})
	addCommonHeaders(response)

	return response
}
