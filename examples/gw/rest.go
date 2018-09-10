package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/lb"
	"github.com/shisa-platform/core/middleware"
	"github.com/shisa-platform/core/service"
)

type Farewell struct {
	Message string
}

func (g Farewell) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"farewell\": %q}", g.Message)), nil
}

type RestService struct {
	service.Service
	balancer lb.Balancer
}

func NewRestService(res lb.Balancer) *RestService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 15,
		AllowTrailingSlashRedirects: true,
	}

	svc := &RestService{
		Service: service.Service{
			Name: "rest",
		},
		balancer: res,
	}

	proxy := middleware.ReverseProxy{
		Router:    svc.router,
		Responder: svc.responder,
	}
	farewell := service.GetEndpointWithPolicy("/api/farewell", policy, proxy.Service)
	farewell.Get.QuerySchemas = []httpx.ParameterSchema{{Name: "name", Multiplicity: 1}}

	svc.Endpoints = []service.Endpoint{farewell}

	return svc
}

func (s *RestService) Name() string {
	return s.Service.Name
}

func (s *RestService) Healthcheck(ctx context.Context) merry.Error {
	_, err := s.balancer.Balance(s.Service.Name)
	if err != nil {
		return err.Prepend("healthcheck")
	}
	return nil
}

func (s *RestService) router(ctx context.Context, request *httpx.Request) (*httpx.Request, merry.Error) {
	addr, err := s.balancer.Balance(s.Service.Name)
	if err != nil {
		return nil, err.Prepend("router")
	}

	request.URL.Scheme = "http"
	request.URL.Host = addr
	request.URL.Path = "/goodbye"

	request.Header.Set("X-Request-Id", ctx.RequestID())
	request.Header.Set("X-User-Id", ctx.Actor().ID())

	return request, nil
}

func (s *RestService) responder(_ context.Context, _ *httpx.Request, response httpx.Response) (httpx.Response, merry.Error) {
	var buf bytes.Buffer
	if err := response.Serialize(&buf); err != nil {
		return nil, err.Prepend("serializing response")
	}
	body := make(map[string]string)
	if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
		return nil, merry.Prepend(err, "unmarshaling response")
	}
	who, ok := body["goodbye"]
	if !ok {
		return nil, merry.New("goodbye key missing from response")
	}

	farewell := httpx.NewOK(Farewell{"Goodbye " + who})
	addCommonHeaders(farewell)

	return farewell, nil
}
