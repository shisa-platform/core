package auxiliary

import (
	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
	"github.com/shisa-platform/core/httpx"
)

type Router func(context.Context, *httpx.Request) httpx.Handler

func (r Router) InvokeSafely(ctx context.Context, request *httpx.Request) (_ httpx.Handler, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in router")

	return r(ctx, request), nil
}
