package auxiliary

import (
	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
)

//go:generate charlatan -output=./authorizer_charlatan.go Authorizer

// Authorizer defines a provider for authorizing principals to
// make requests.
type Authorizer interface {
	// Authorize returns `true` if the principal in the context
	// object is allowed to peform the given request.  If the
	// principal is not allowed `false` must be returned, not an
	// error.  If there is a problem completing authorization an
	// error should be returned.
	Authorize(context.Context, *httpx.Request) (bool, merry.Error)
}
