package ratelimit

import (
	"fmt"
	"time"
)

type Provider interface {
	// Limit returns the policy based rate limit for the given actor performing the action on the path.
	Limit(actor, action, path string) (RateLimit, error)
	// Allow returns true if the rate limit policy allows the given actor to perform the action on the path.
	Allow(actor, action, path string) (bool, error)
	Ping() error
	Close()
}

type ErrTooManyRequests struct {
	Cooloff time.Duration
}

func (e *ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests.  wait %s", e.Cooloff)
}
