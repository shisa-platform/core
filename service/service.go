package service

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

//go:generate charlatan -output=./service_charlatan.go Service

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *Request, merry.Error) Response

type Service interface {
	Name() string          // The name of the service.  Must not be empty
	Endpoints() []Endpoint // The enpoints of this service.  Must not be empty.
	Handlers() []Handler   // Optional common handlers used for every endpoint.

	// MethodNotAllowedHandler optionally customizes the
	// response returned to the user agent when an endpoint isn't
	// configured to service the method of a request.
	// If nil the default handler will return a 405 status code
	// with an empty body.
	MethodNotAllowedHandler() Handler

	// InternalServerErrorHandler optionally customizes the
	// response returned to the user agent when no the gateway
	// encounters an error trying to service a request.
	// If nil the default handler will return a 500 status code
	// with an empty body.
	InternalServerErrorHandler() ErrorHandler
}
