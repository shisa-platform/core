package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

type Greeting struct {
	Message string `json:"greeting"`
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
			Method:   http.MethodGet,
			Route:    "/greeting",
			Pipeline: []service.Handler{s.Greeting},
			Policy: service.Policy{
				RequestBudget:     time.Millisecond * 5,
				GenerateRequestID: true,
			},
		},
	}

}

func (s *HelloService) Greeting(context.Context, *service.Request) service.Response {
	now := time.Now().UTC().Format(time.RFC1123)

	response := service.NewOK(Greeting{"hello, world"})
	response.Header().Set("Cache-Control", "private, max-age=0")
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.Header().Set("Date", now)
	response.Header().Set("Expires", now)
	response.Header().Set("Last-Modified", now)
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.Header().Set("X-Frame-Options", "DENY")
	response.Header().Set("X-Xss-Protection", "1; mode=block")

	return response
}
