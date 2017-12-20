package service

import (
	"github.com/percolate/shisa/context"
)

type Handler func(context.Context, *Request) Response

type Pipeline struct {
	Policy   Policy
	Handlers []Handler
}

type Endpoint struct {
	Route   string
	Head    *Pipeline
	Get     *Pipeline
	Put     *Pipeline
	Post    *Pipeline
	Patch   *Pipeline
	Delete  *Pipeline
	Connect *Pipeline
	Options *Pipeline
	Trace   *Pipeline
	// xxx - request (query|body) parameters
}

func GetEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Get: &Pipeline{
			Handlers: handlers,
		},
	}
}

func GetEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Get: &Pipeline{
			Policy: policy,
			Handlers: handlers,
		},
	}
}

func PutEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Put: &Pipeline{
			Handlers: handlers,
		},
	}
}

func PutEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Put: &Pipeline{
			Policy: policy,
			Handlers: handlers,
		},
	}
}

func PostEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Post: &Pipeline{
			Handlers: handlers,
		},
	}
}

func PostEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Post: &Pipeline{
			Policy: policy,
			Handlers: handlers,
		},
	}
}

func PatchEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Patch: &Pipeline{
			Handlers: handlers,
		},
	}
}

func PatchEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Patch: &Pipeline{
			Policy: policy,
			Handlers: handlers,
		},
	}
}

func DeleteEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Delete: &Pipeline{
			Handlers: handlers,
		},
	}
}

func DeleteEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Delete: &Pipeline{
			Policy: policy,
			Handlers: handlers,
		},
	}
}
