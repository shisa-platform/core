package service

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

//go:generate charlatan -output=./service_charlatan.go Service

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *Request, merry.Error) Response

type Service interface {
	Name() string          // Service name.  Required.
	Endpoints() []Endpoint // Service endpoints. Requried.

	// Handlers are optional handlers that should be invoked for
	// all endpoints.  These will be prepended to all endpoint
	// handlers when a service is registered.
	Handlers() []Handler

	// MalformedQueryParameterHandler optionally customizes the
	// response to the user agent when malformed query parameters
	// are presented.
	// If nil the default handler wil return a 400 status code
	// with an empty body.
	MalformedQueryParameterHandler() Handler

	// MethodNotAllowedHandler optionally customizes the response
	// returned to the user agent when an endpoint isn't
	// configured to service the method of a request.
	// If nil the default handler will return a 405 status code
	// with an empty body.
	MethodNotAllowedHandler() Handler

	// RedirectHandler optionally customizes the response
	// returned to the user agent when an endpoint is configured
	// to return a redirect for a path based on a missing or
	// extra trailing slash.
	// If nil the default handler will return a 303 status code
	// for GET and a 307 for other methods, both with an empty
	// body.
	RedirectHandler() Handler

	// InternalServerErrorHandler optionally customizes the
	// response returned to the user agent when the gateway
	// encounters an error trying to service a request.
	// If nil the default handler will return a 500 status code
	// with an empty body.
	InternalServerErrorHandler() ErrorHandler
}

// ServiceAdapter implements several of the methods of the
// Service interface to simplify buildings services that don't
// need the customization hooks.
// Add ServiceAdatper as an anonymous field in your service's
// struct to inherit these default method implementations.
type ServiceAdapter struct{}

func (s *ServiceAdapter) Handlers() []Handler {
	return nil
}

func (s *ServiceAdapter) MalformedQueryParameterHandler() Handler {
	return nil
}

func (s *ServiceAdapter) MethodNotAllowedHandler() Handler {
	return nil
}

func (s *ServiceAdapter) RedirectHandler() Handler {
	return nil
}

func (s *ServiceAdapter) InternalServerErrorHandler() ErrorHandler {
	return nil
}
