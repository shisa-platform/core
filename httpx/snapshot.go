package httpx

import (
	"time"
)

// ResponseSnapshot captures details about the response sent to
// the user agent.
type ResponseSnapshot struct {
	// Status code returned to the user agent
	StatusCode int
	// Size of the response body in bytes
	Size int
	// Time request servicing began
	Start time.Time
	// Duration of request servicing
	Elapsed time.Duration
}
