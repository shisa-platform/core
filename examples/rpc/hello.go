package main

import (
	"fmt"

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

	return response
}
