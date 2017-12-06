package service

import (
	"time"
)

// Policy controls the behavior of per-endpoint request processing before control is passed to the handler.
type Policy struct {
	AllowUnknownQueryParameters   bool          // will unknown query parameters be silently ignored?
	AllowMalformedQueryParameters bool          // will malformed query parameter pairs be silently ignored?
	CustomMalformedRequestStatus  int           // this status code will be returned instead of 400 on malformed requests
	TimeBudget                    time.Duration // the time budget for requests to the endpoint
	RequestIDResponseHeaderName   string        // customize the name of the response header for the request id [default: X-Request-ID]
}
