package service

import (
	"github.com/percolate/shisa/context"
)

type Handler func(context.Context, *Request) Response

type Pipeline struct {
	Policy   Policy
	Handlers []Handler
}

type Endpoint struct {
	Service Service
	Route   string
	Head    *Pipeline
	Get     *Pipeline
	Put     *Pipeline
	Post    *Pipeline
	Patch   *Pipeline
	Delete  *Pipeline
	Connect *Pipeline
	Options *Pipeline
	Trace   *Pipeline
	// xxx - request (query|body) parameters
}
