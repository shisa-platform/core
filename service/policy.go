package service

import (
	"time"
)

// Policy controls the behavior of per-endpoint request
// processing before control is passed to the handler.
type Policy struct {
	// Will malformed query parameters be passed through or
	// rejected?
	AllowMalformedQueryParameters bool `json:",omitempty"`
	// Will unknown query parameters be passed through or
	// rejected?
	AllowUnknownQueryParameters bool `json:",omitempty"`
	// Will requests  with missing/extra trailing slash
	// be redirected?
	AllowTrailingSlashRedirects bool `json:",omitempty"`
	// Will URL escaped path parameters be preserved?
	PreserveEscapedPathParameters bool `json:",omitempty"`
	// The time budget for the pipeline to complete
	TimeBudget time.Duration `json:",omitempty"`
}
