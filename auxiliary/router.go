package auxiliary

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/httpx"
)

type Router func(context.Context, *httpx.Request) httpx.Handler

func (r Router) InvokeSafely(ctx context.Context, request *httpx.Request, exception *merry.Error) httpx.Handler {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err, ok := arg.(error); ok {
			*exception = merry.WithMessage(err, "panic in router")
			return
		}

		*exception = merry.New("panic in router").WithValue("context", arg)
	}()

	return r(ctx, request)
}
