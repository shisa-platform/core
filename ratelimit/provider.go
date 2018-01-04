package ratelimit

import (
	"github.com/ansel1/merry"
)

//go:generate charlatan -output=./provider_charlatan.go Provider
type Provider interface {
	// Limit returns the policy based rate limit for the given actor performing the action on the path.
	Limit(actor, action, path string) (RateLimit, merry.Error)
	// Allow returns true if the rate limit policy allows the given actor to perform the action on the path.
	Allow(actor, action, path string) (bool, merry.Error)
	Ping() error
	Close()
}
