package sd

import (
	"net/url"

	"github.com/ansel1/merry"
)

type Registrar interface {
	Register(serviceID, addr string) merry.Error
	Deregister(serviceID string) merry.Error
}

type Resolver interface {
	Resolve(name string) ([]string, merry.Error)
}

type AsyncResolver interface {
	Resolver
	Shutdown()
	IsResolving() bool
}

type Healthchecker interface {
	AddCheck(service string, url *url.URL) merry.Error
	RemoveChecks(service string) merry.Error
}
