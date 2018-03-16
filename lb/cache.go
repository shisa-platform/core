package lb

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/sd"
)

// Cache encapsulates the stateful parts of a load balancer
type Cache interface {
	// Next returns an addr given a service name and current addrs
	Next(service string, nodes []string) string
}

var _ Balancer = &CacheBalancer{}

type CacheBalancer struct {
	Cache    Cache
	Resolver sd.Resolver
}

func (r *CacheBalancer) Balance(service string) (string, merry.Error) {
	if r.Cache == nil {
		return "", merry.New("cache must not be nil")
	}

	if r.Resolver == nil {
		return "", merry.New("resolver must not be nil")
	}

	nodes, err := r.Resolver.Resolve(service)
	if err != nil {
		return "", err.Prepend("balance")
	}

	next := r.Cache.Next(service, nodes)

	if next == "" {
		return "", merry.New("no nodes available")
	}

	return next, nil
}
