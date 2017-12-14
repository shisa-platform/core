package service

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/context"
)

//go:generate charlatan -output=./service_charlatan.go Service

// ErrorHandler creates a response for the given error condition.
type ErrorHandler func(context.Context, *Request, merry.Error) Response

type Service interface {
	Name() string          // The name of the service. Must exist
	Endpoints() []Endpoint // The enpoints of this service. Must exist
	Handlers() []Handler   // Optional handlers for all endpoints

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
