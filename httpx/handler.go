package httpx

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/errorx"
)

// Handler is a block of logic to apply to a request.
// Returning a non-nil Response indicates request processing
// should stop.
type Handler func(context.Context, *Request) Response

func (h Handler) InvokeSafely(ctx context.Context, request *Request) (_ Response, exception merry.Error) {
	defer handlerRecovery(&exception)
	return h(ctx, request), nil
}

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *Request, merry.Error) Response

func (h ErrorHandler) InvokeSafely(ctx context.Context, request *Request, err merry.Error) (_ Response, exception merry.Error) {
	defer handlerRecovery(&exception)
	return h(ctx, request, err), nil
}

func handlerRecovery(exception *merry.Error) {
	arg := recover()
	if arg == nil {
		return
	}

	*exception = errorx.CapturePanic(arg, "panic in handler")
}
