package service

import (
	"github.com/percolate/shisa/context"
)

// Handler is a block of logic to apply to a request.
// Returning a non-nil response indicates further request
// processing should be stopped.
type Handler func(context.Context, *Request) Response

// Pipeline is a chain of handlers to be invoked in order on a
// request.  The first non-nil response will be returned to the
// user agent.  If no response is produced an Internal Service
// Error handler will be invoked.
type Pipeline struct {
	Policy   Policy // customizes automated behavior
	Handlers []Handler
}

// Endpoint is collection of pipelines for a route (URL path),
// one for each HTTP method.  Only supported methods should have
// pipelines, but at least one pipleline is requried.
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

// GetEndpoint returns an Endpoint configured for the given route
// with the GET pipeline using the given handlers.
func GetEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Get: &Pipeline{
			Handlers: handlers,
		},
	}
}

// GetEndpointWithPolicy returns an Endpoint configured for the
// given route, with the given policy and the GET pipeline using
// the given handlers.
func GetEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Get: &Pipeline{
			Policy:   policy,
			Handlers: handlers,
		},
	}
}

// PutEndpoint returns an Endpoint configured for the given route
// with the PUT pipeline using the given handlers.
func PutEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Put: &Pipeline{
			Handlers: handlers,
		},
	}
}

// PutEndpointWithPolicy returns an Endpoint configured for the
// given route, with the given policy and the PUT pipeline using
// the given handlers.
func PutEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Put: &Pipeline{
			Policy:   policy,
			Handlers: handlers,
		},
	}
}

// PostEndpoint returns an Endpoint configured for the given route
// with the POST pipeline using the given handlers.
func PostEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Post: &Pipeline{
			Handlers: handlers,
		},
	}
}

// PostEndpointWithPolicy returns an Endpoint configured for the
// given route, with the given policy and the POST pipeline
// using the given handlers.
func PostEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Post: &Pipeline{
			Policy:   policy,
			Handlers: handlers,
		},
	}
}

// PatchEndpoint returns an Endpoint configured for the given
// route with the PATCH pipeline using the given handlers.
func PatchEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Patch: &Pipeline{
			Handlers: handlers,
		},
	}
}

// PatchEndpointWithPolicy returns an Endpoint configured for the
// given route, with the given policy and the PATCH pipeline
// using the given handlers.
func PatchEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Patch: &Pipeline{
			Policy:   policy,
			Handlers: handlers,
		},
	}
}

// DeleteEndpoint returns an Endpoint configured for the given
// route with the DELETE pipeline using the given handlers.
func DeleteEndpoint(route string, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Delete: &Pipeline{
			Handlers: handlers,
		},
	}
}

// DeleteEndpointWithPolicy returns an Endpoint configured for
// the given route, with the given policy and the DELETE
// pipeline using the given handlers.
func DeleteEndpointWithPolicy(route string, policy Policy, handlers ...Handler) Endpoint {
	return Endpoint{
		Route: route,
		Delete: &Pipeline{
			Policy:   policy,
			Handlers: handlers,
		},
	}
}
