package sd

import "github.com/ansel1/merry"

type Registrar interface {
	Register()
	Deregister()
}

type Resolver interface {
	Resolve(name string, passingOnly bool) ([]string, merry.Error)
}

type AsyncResolver interface {
	Resolver
	Shutdown()
	IsResolving()
}
