package httpx

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

// StringExtractor is a function type that extracts a string from
// the given `context.Context` and `*httpx.Request`.
// An error is returned if the string could not be extracted.
type StringExtractor func(context.Context, *Request) (string, merry.Error)

func (h StringExtractor) InvokeSafely(ctx context.Context, request *Request) (str string, err merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			err = merry.Prepend(err1, "panic in extractor")
			return
		}

		err = merry.New("panic in extractor").WithValue("context", arg)
	}()

	return h(ctx, request)
}

// RequestPredicate examines the given context and request and
// returns a determination based on that analysis.
type RequestPredicate func(context.Context, *Request) bool

func (h RequestPredicate) InvokeSafely(ctx context.Context, request *Request) (ok bool, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			exception = merry.Prepend(err1, "panic in request predicate")
			return
		}

		exception = merry.New("panic in request predicate").WithValue("context", arg)
	}()

	return h(ctx, request), nil
}
