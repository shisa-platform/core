package httpx

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

type ErrorHook func(context.Context, *Request, merry.Error)

func (h ErrorHook) InvokeSafely(ctx context.Context, request *Request, err merry.Error) (exception merry.Error) {
	if h == nil {
		return
	}

	defer hookRecovery(&exception)
	h(ctx, request, err)

	return
}

type CompletionHook func(context.Context, *Request, ResponseSnapshot)

func (h CompletionHook) InvokeSafely(ctx context.Context, request *Request, snapshot ResponseSnapshot) (exception merry.Error) {
	if h == nil {
		return
	}

	defer hookRecovery(&exception)
	h(ctx, request, snapshot)

	return
}

func hookRecovery(exception *merry.Error) {
	arg := recover()
	if arg == nil {
		return
	}

	if err, ok := arg.(error); ok {
		*exception = merry.Prepend(err, "panic in hook")
		return
	}

	*exception = merry.New("panic in hook").WithValue("context", arg)
}
