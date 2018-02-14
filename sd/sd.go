package sd

import "github.com/ansel1/merry"

type Registrar interface {
	Register(name, addr string) merry.Error
	Deregister(name string) merry.Error
}

type Resolver interface {
	Resolve(name string) ([]string, merry.Error)
}

type AsyncResolver interface {
	Resolver
	Shutdown()
	IsResolving() bool
}
