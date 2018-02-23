package httpx

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

// StringExtractor is a function type that extracts a string from
// the given `context.Context` and `*httpx.Request`.
// An error is returned if the string could not be extracted.
type StringExtractor func(context.Context, *Request) (string, merry.Error)
