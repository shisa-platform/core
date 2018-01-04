package main

import (
	"fmt"
	"time"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

var (
	AmericanEnglish   = "en-US"
	BritishEnglish    = "en-GB"
	EuropeanSpanish   = "es-ES"
	Finnish           = "fi"
	French            = "fr"
	Japanese          = "ja"
	SimplifiedChinese = "zh-Hans"
	tags              = []string{
		AmericanEnglish,
		BritishEnglish,
		EuropeanSpanish,
		Finnish,
		French,
		Japanese,
		SimplifiedChinese,
	}
	greetings = map[string]string{
		AmericanEnglish:   "Hello",
		BritishEnglish:    "Cheerio",
		EuropeanSpanish:   "Hola",
		Finnish:           "Hei",
		French:            "Bonjour",
		Japanese:          "こんにちは",
		SimplifiedChinese: "你好",
	}
	language = service.Field{
		Name:         "language",
		Default:      AmericanEnglish,
		Validator:    service.StringSliceValidator{tags}.Validate,
		Multiplicity: 1,
	}
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

	greeting := service.GetEndpointWithPolicy("/greeting", policy, svc.Greeting)
	greeting.Get.QueryFields = []service.Field{
		language,
		{Name: "name", Multiplicity: 1},
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

func (s *HelloService) Greeting(ctx context.Context, r *service.Request) service.Response {
	var greeting string
	interlocutor := ctx.Actor().String()

	for _, param := range r.QueryParams {
		switch param.Name {
		case "language":
			greeting = greetings[param.Values[0]]
		case "name":
			interlocutor = param.Values[0]
		}
	}

	response := service.NewOK(Greeting{fmt.Sprintf("%s %s", greeting, interlocutor)})
	addCommonHeaders(response)

	return response
}

func (s *HelloService) Salutaion(ctx context.Context, r *service.Request) service.Response {
	salutation := greetings[r.QueryParams[0].Values[0]]
	response := service.NewOK(Greeting{fmt.Sprintf("%s %s", salutation, ctx.Actor().String())})
	addCommonHeaders(response)

	return response
}
