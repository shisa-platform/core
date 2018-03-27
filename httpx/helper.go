package httpx

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

// StringExtractor is a function type that extracts a string from
// the given `context.Context` and `*httpx.Request`.
// An error is returned if the string could not be extracted.
type StringExtractor func(context.Context, *Request) (string, merry.Error)

func (h StringExtractor) InvokeSafely(ctx context.Context, request *Request) (str string, err merry.Error, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			exception = merry.Prepend(err1, "panic in request extractor")
			return
		}

		exception = merry.Errorf("panic in request extractor: \"%v\"", arg)
	}()

	str, err = h(ctx, request)

	return
}

// RequestPredicate examines the given context and request and
// returns a determination based on that analysis.
type RequestPredicate func(context.Context, *Request) bool

func (h RequestPredicate) InvokeSafely(ctx context.Context, request *Request) (_ bool, exception merry.Error) {
	defer func() {
		arg := recover()
		if arg == nil {
			return
		}

		if err1, ok := arg.(error); ok {
			exception = merry.Prepend(err1, "panic in request predicate")
			return
		}

		exception = merry.Errorf("panic in request predicate: \"%v\"", arg)
	}()

	return h(ctx, request), nil
}
