package sd

type Registrar interface {
	Register()
	Deregister()
}

type Node string

type Nodes []Node

type Resolver interface {
	Resolve(name string) Nodes
}

type AsyncResolver interface {
	Resolver
	Shutdown()
	IsResolving()
}
