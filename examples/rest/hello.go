package main

import (
	"fmt"

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
}

func (s *HelloService) Name() string {
	return "hello"
}

func (s *HelloService) Endpoints() []service.Endpoint {
	return []service.Endpoint{
		service.Endpoint{
			Route: "/greeting",
			Get: &service.Pipeline{
				Policy:   commonPolicy,
				Handlers: []service.Handler{s.Greeting},
			},
		},
	}
}

func (s *HelloService) Handlers() []service.Handler {
	return nil
}

func (s *HelloService) MethodNotAllowedHandler() service.Handler {
	return nil
}

func (s *HelloService) RedirectHandler() service.Handler {
	return nil
}

func (s *HelloService) InternalServerErrorHandler() service.ErrorHandler {
	return nil
}

func (s *HelloService) Greeting(context.Context, *service.Request) service.Response {
	response := service.NewOK(Greeting{"hello, world"})
	addCommonHeaders(response)
	response.Trailers().Add("test", "foo")

	return response
}
