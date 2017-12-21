package authn

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/models"
)

//go:generate charlatan -output=./idp_charlatan.go IdentityProvider

// IdentityProvider is an service that resolves tokens into
// pricipals.
type IdentityProvider interface {
	Authenticate(string) (models.User, merry.Error)
}
