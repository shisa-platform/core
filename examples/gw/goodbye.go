package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/middleware"
	"github.com/percolate/shisa/sd"
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
	endpoints []service.Endpoint
	resolver  sd.Resolver
}

func NewGoodbyeService(res sd.Resolver) *GoodbyeService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}

	svc := &GoodbyeService{
		resolver: res,
	}

	proxy := middleware.ReverseProxy{
		Router:    svc.router,
		Responder: svc.responder,
	}
	farewell := service.GetEndpointWithPolicy("/api/farewell", policy, proxy.Service)
	farewell.Get.QueryFields = []service.Field{{Name: "name", Multiplicity: 1}}

	svc.endpoints = []service.Endpoint{farewell}

	return svc
}

func (s *GoodbyeService) Name() string {
	return "goodbye"
}

func (s *GoodbyeService) Endpoints() []service.Endpoint {
	return s.endpoints
}

func (s *GoodbyeService) Healthcheck(ctx context.Context) merry.Error {
	addrs, merr := s.resolver.Resolve(s.Name())
	if merr != nil {
		return merry.Prepend(merr, "healthcheck")
	}

	if len(addrs) < 1 {
		return merry.New("no healthy hosts")
	}

	return nil
}

func (s *GoodbyeService) router(ctx context.Context, request *httpx.Request) (*httpx.Request, merry.Error) {
	addrs, err := s.resolver.Resolve(s.Name())
	if err != nil {
		return nil, merry.Prepend(err, "router")
	}

	if len(addrs) < 1 {
		return nil, merry.New("no healthy hosts")
	}

	request.URL.Scheme = "http"
	request.URL.Host = addrs[0]
	request.URL.Path = "/goodbye"

	request.Header.Set("X-Request-Id", ctx.RequestID())
	request.Header.Set("X-User-Id", ctx.Actor().ID())

	return request, nil
}

func (s *GoodbyeService) responder(_ context.Context, _ *httpx.Request, response httpx.Response) httpx.Response {
	var buf bytes.Buffer
	if err := response.Serialize(&buf); err != nil {
		return httpx.NewEmptyError(http.StatusBadGateway, err)
	}
	body := make(map[string]string)
	if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
		return httpx.NewEmptyError(http.StatusBadGateway, err)
	}
	who, ok := body["goodbye"]
	if !ok {
		err := merry.New("goodbye key missing from response")
		return httpx.NewEmptyError(http.StatusBadGateway, err)
	}

	farewell := httpx.NewOK(Farewell{"Goodbye " + who})
	addCommonHeaders(farewell)

	return farewell
}
