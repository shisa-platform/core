package httpx

import (
	"github.com/percolate/shisa/context"
)

// Handler is a block of logic to apply to a request.
// Returning a non-nil response indicates further request
// processing should be stopped.
type Handler func(context.Context, *service.Request) Response

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *service.Request, merry.Error) Response
