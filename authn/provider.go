package authn

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
	"github.com/percolate/shisa/models"
	"github.com/percolate/shisa/service"
)

//go:generate charlatan -output=./provider_charlatan.go Provider

type Provider interface {
	Authenticate(context.Context, *service.Request) (models.User, merry.Error)
}
