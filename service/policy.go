package service

import (
	"time"
)

// Policy controls the behavior of per-endpoint request
// processing before control is passed to the handler.
type Policy struct {
	// Will malformed query parameter pairs be silently ignored?
	AllowMalformedQueryParameters bool
	// Will unknown query parameters be silently ignored?
	AllowUnknownQueryParameters bool
	// Redirect requests for routes with the opposite trailing
	// slash to this endpoint
	AllowTrailingSlashRedirects bool
	// This status code will be returned instead of 400 on
	// malformed requests
	CustomMalformedRequestStatus int
	// Will URL escaped path parameters be preserved?
	PreserveEscapedPathParameters bool
	// The time budget for requests to the endpoint
	TimeBudget time.Duration
}
