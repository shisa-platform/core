package service

import (
	"time"
)

// Policy controls the behavior of per-endpoint request
// processing before control is passed to the handler.
type Policy struct {
	// Will malformed query parameters be passed through or
	// rejected?
	AllowMalformedQueryParameters bool
	// Will unknown query parameters be passed through or
	// rejected?
	AllowUnknownQueryParameters bool
	// Redirect requests for routes with the opposite trailing
	// slash to this endpoint
	AllowTrailingSlashRedirects bool
	// Will URL escaped path parameters be preserved?
	PreserveEscapedPathParameters bool
	// The time budget for requests to the endpoint
	TimeBudget time.Duration
}
