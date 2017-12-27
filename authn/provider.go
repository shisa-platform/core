package authn

import (
	"net/http"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

var (
	authHeaderKey = http.CanonicalHeaderKey("Authorization")
)

//go:generate charlatan -output=./provider_charlatan.go Provider

type Provider interface {
	// Authenticate extracts a token from the request and
	// resolves it into a user principal.
	Authenticate(context.Context, *service.Request) (models.User, merry.Error)
	// Challenge returns the value for the "WWW-Authenticate"
	// header if authentication fails.
	Challenge() string
}
