package authn

import (
	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/models"
)

//go:generate charlatan -output=./idp_charlatan.go IdentityProvider

// IdentityProvider is a service that resolves tokens into
// principals.
type IdentityProvider interface {
	Authenticate(context.Context, string) (models.User, merry.Error)
}
