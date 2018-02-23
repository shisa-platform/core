package httpx

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

// Handler is a block of logic to apply to a request.
// Returning a non-nil response indicates further request
// processing should be stopped.
type Handler func(context.Context, *Request) Response

func (h Handler) InvokeSafely(ctx context.Context, request *Request, exception *merry.Error) Response {
	defer recovery(exception)
	return h(ctx, request)
}

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *Request, merry.Error) Response

func (h ErrorHandler) InvokeSafely(ctx context.Context, request *Request, err merry.Error, exception *merry.Error) Response {
	defer recovery(exception)
	return h(ctx, request, err)
}

func recovery(exception *merry.Error) {
	arg := recover()
	if arg == nil {
		return
	}

	if err, ok := arg.(error); ok {
		*exception = merry.WithMessage(err, "panic in handler")
		return
	}

	*exception = merry.New("panic in handler").WithValue("context", arg)
}
