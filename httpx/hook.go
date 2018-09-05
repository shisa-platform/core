package httpx

import (
	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
)

type ErrorHook func(context.Context, *Request, merry.Error)

func (h ErrorHook) InvokeSafely(ctx context.Context, request *Request, err merry.Error) (exception merry.Error) {
	if h == nil {
		return
	}

	defer errorx.CapturePanic(&exception, "panic in hook")
	h(ctx, request, err)

	return
}

type CompletionHook func(context.Context, *Request, ResponseSnapshot)

func (h CompletionHook) InvokeSafely(ctx context.Context, request *Request, snapshot ResponseSnapshot) (exception merry.Error) {
	if h == nil {
		return
	}

	defer errorx.CapturePanic(&exception, "panic in hook")
	h(ctx, request, snapshot)

	return
}
