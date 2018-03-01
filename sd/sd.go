package sd

import (
	"net/url"

	"github.com/ansel1/merry"
)

//go:generate charlatan -output=./registrar_charlatan.go Registrar
type Registrar interface {
	Register(serviceID, addr string) merry.Error
	Deregister(serviceID string) merry.Error
	AddCheck(service string, url *url.URL) merry.Error
	RemoveChecks(service string) merry.Error
}

type Resolver interface {
	Resolve(name string) ([]string, merry.Error)
}

type AsyncResolver interface {
	Resolver
	Shutdown()
	IsResolving() bool
}
