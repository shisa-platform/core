package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/env"
	"github.com/percolate/shisa/httpx"
	"github.com/percolate/shisa/middleware"
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
	env       env.Provider
	endpoints []service.Endpoint
}

func NewGoodbyeService(environment env.Provider) *GoodbyeService {
	policy := service.Policy{
		TimeBudget:                  time.Millisecond * 5,
		AllowTrailingSlashRedirects: true,
	}

	svc := &GoodbyeService{
		env: environment,
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
	addr, envErr := s.env.Get(goodbyeServiceAddrEnv)
	if envErr != nil {
		return envErr.WithUserMessage("address environment variable not found")
	}

	url := "http://" + addr + "/healthcheck"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return merry.Wrap(err).WithUserMessage("unable to create request")
	}
	request.Header.Set("X-Request-Id", ctx.RequestID())

	if ctx.Actor() != nil {
		request.Header.Set("X-User-Id", ctx.Actor().ID())
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return merry.Wrap(err).WithUserMessage("unable to complete request")
	}
	if response.StatusCode != http.StatusOK {
		return merry.New("not ready").WithUserMessage("not ready")
	}

	return nil
}

func (s *GoodbyeService) router(ctx context.Context, request *httpx.Request) (*httpx.Request, merry.Error) {
	addr, envErr := s.env.Get(goodbyeServiceAddrEnv)
	if envErr != nil {
		return nil, envErr
	}

	request.URL.Scheme = "http"
	request.URL.Host = addr
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
