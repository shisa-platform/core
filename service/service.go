package service

import (
	"github.com/percolate/shisa/httpx"
)

//go:generate charlatan -output=./service_charlatan.go Service

// Service is a logical grouping of related endpoints.
// Examples of relationships are serving the same product
// vertical or requiring the same resources.  This is an
// interface instead of a struct so that implementations can
// store dependencies in the struct.
type Service interface {
	Name() string          // Service name.  Required.
	Endpoints() []Endpoint // Service endpoints. Requried.

	// Handlers are optional handlers that should be invoked for
	// all endpoints.  These will be prepended to all endpoint
	// handlers when a service is registered.
	Handlers() []httpx.Handler

	// MalformedRequestHandler optionally customizes the
	// response to the user agent when a malformed request is
	// presented.
	// If nil the default handler wil return a 400 status code
	// with an empty body.
	MalformedRequestHandler() httpx.Handler

	// MethodNotAllowedHandler optionally customizes the response
	// returned to the user agent when an endpoint isn't
	// configured to service the method of a request.
	// If nil the default handler will return a 405 status code
	// with an empty body.
	MethodNotAllowedHandler() httpx.Handler

	// RedirectHandler optionally customizes the response
	// returned to the user agent when an endpoint is configured
	// to return a redirect for a path based on a missing or
	// extra trailing slash.
	// If nil the default handler will return a 303 status code
	// for GET and a 307 for other methods, both with an empty
	// body.
	RedirectHandler() httpx.Handler

	// InternalServerErrorHandler optionally customizes the
	// response returned to the user agent when the gateway
	// encounters an error trying to service a request to an
	// endoint of this service.
	// If nil the default handler will return a 500 status code
	// with an empty body.
	InternalServerErrorHandler() httpx.ErrorHandler
}

// ServiceAdapter implements several of the methods of the
// Service interface to simplify buildings services that don't
// need the customization hooks.
// Add ServiceAdatper as an anonymous field in your service's
// struct to inherit these default method implementations.
type ServiceAdapter struct{}

func (s *ServiceAdapter) Handlers() []httpx.Handler {
	return nil
}

func (s *ServiceAdapter) MalformedRequestHandler() httpx.Handler {
	return nil
}

func (s *ServiceAdapter) MethodNotAllowedHandler() httpx.Handler {
	return nil
}

func (s *ServiceAdapter) RedirectHandler() httpx.Handler {
	return nil
}

func (s *ServiceAdapter) InternalServerErrorHandler() httpx.ErrorHandler {
	return nil
}
