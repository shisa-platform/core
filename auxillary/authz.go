package auxillary

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/service"
)

//go:generate charlatan -output=./authorizer_charlatan.go Authorizer

// Authorizer defines a provder for authorizing principals to
// make requests.
type Authorizer interface {
	Authorize(context.Context, *service.Request) merry.Error
}
