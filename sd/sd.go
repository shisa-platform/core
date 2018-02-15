package sd

import "github.com/ansel1/merry"

type Registrar interface {
	Register(serviceID, addr string) merry.Error
	Deregister(serviceID string) merry.Error
	AddHealthcheck(serviceID, url string) merry.Error
	RemoveHealthcheck(serviceID, url string) merry.Error
}

type Resolver interface {
	Resolve(name string) ([]string, merry.Error)
}

type AsyncResolver interface {
	Resolver
	Shutdown()
	IsResolving() bool
}
