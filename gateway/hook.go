package gateway

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

type ErrorHook func(context.Context, *httpx.Request, merry.Error)

func (h ErrorHook) InvokeSafely(ctx context.Context, request *httpx.Request, err merry.Error, exception *merry.Error) {
	if h == nil {
		return
	}

	defer recovery(exception)
	h(ctx, request, err)
}

type CompletionHook func(context.Context, *httpx.Request, httpx.ResponseSnapshot)

func (h CompletionHook) InvokeSafely(ctx context.Context, request *httpx.Request, snapshot httpx.ResponseSnapshot, exception *merry.Error) {
	if h == nil {
		return
	}

	defer recovery(exception)
	h(ctx, request, snapshot)
}

func recovery(exception *merry.Error) {
	arg := recover()
	if arg == nil {
		return
	}

	if err, ok := arg.(error); ok {
		*exception = merry.WithMessage(err, "panic in hook")
		return
	}

	*exception = merry.New("panic in hook").WithValue("context", arg)
}
