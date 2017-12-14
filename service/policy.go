package service

import (
	"time"
)

// Policy controls the behavior of per-endpoint request processing before control is passed to the handler.
type Policy struct {
	AllowMalformedQueryParameters bool          // will malformed query parameter pairs be silently ignored?
	AllowUnknownQueryParameters   bool          // will unknown query parameters be silently ignored?
	AllowTrailingSlashRedirects   bool          // redirect requests for routes with the opposite trailing slash to this endpoint
	CustomMalformedRequestStatus  int           // this status code will be returned instead of 400 on malformed requests
	PreserveEscapedPathParameters bool          // will URL escaped path parameters be preserved?
	TimeBudget                    time.Duration // the time budget for requests to the endpoint
}
