package service

import (
	"github.com/percolate/shisa/context"
)

type Handler func(context.Context, *Request) Response

type Endpoint struct {
	Method  string
	Route   string
	Policy  Policy
	Handler Handler
	// xxx - request (query|body) parameters
	// xxx - pipeline instead of handler
}
