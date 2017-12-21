package authn

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/models"
)

//go:generate charlatan -output=./idp_charlatan.go IdentityProvider

// IdentityProvider is a service that resolves tokens into
// principals.
type IdentityProvider interface {
	Authenticate(string) (models.User, merry.Error)
}
