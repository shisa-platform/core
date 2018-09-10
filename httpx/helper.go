package httpx

import (
	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/errorx"
)

// StringExtractor is a function type that extracts a string from
// the given `context.Context` and `*httpx.Request`.
// An error is returned if the string could not be extracted.
type StringExtractor func(context.Context, *Request) (string, merry.Error)

func (h StringExtractor) InvokeSafely(ctx context.Context, request *Request) (str string, err merry.Error, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in request extractor")

	str, err = h(ctx, request)

	return
}

// RequestPredicate examines the given context and request and
// returns a determination based on that analysis.
type RequestPredicate func(context.Context, *Request) bool

func (h RequestPredicate) InvokeSafely(ctx context.Context, request *Request) (_ bool, exception merry.Error) {
	defer errorx.CapturePanic(&exception, "panic in request predicate")

	return h(ctx, request), nil
}
