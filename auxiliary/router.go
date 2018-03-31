package auxiliary

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/errorx"
	"github.com/percolate/shisa/httpx"
)

type Router func(context.Context, *httpx.Request) httpx.Handler

func (r Router) InvokeSafely(ctx context.Context, request *httpx.Request) (_ httpx.Handler, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		exception = errorx.Panic(arg, "panic in router")
	}()

	return r(ctx, request), nil
}
