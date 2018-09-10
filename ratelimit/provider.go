package ratelimit

import (
	"github.com/ansel1/merry"
	"time"

	"github.com/shisa-platform/core/context"
)

//go:generate charlatan -output=./provider_charlatan.go Provider

// Provider is an interface providing a means to limiting
// requests based on actor/action/path parameters.
type Provider interface {
	// Limit returns the policy based rate limit for the given actor performing the action on the path.
	Limit(ctx context.Context, actor, action, path string) (RateLimit, merry.Error)
	// Allow returns true if the rate limit policy allows the given actor to perform the action on the path.
	// If the rate limit policy disallows the action, the cooldown duration is also returned. Allow only
	// returns an error due to internal failure.
	Allow(ctx context.Context, actor, action, path string) (bool, time.Duration, merry.Error)
	Close()
}
