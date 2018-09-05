package lb

import (
	"github.com/ansel1/merry"

	"github.com/shisa-platform/core/sd"
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
		return "", merry.New("load balancer: check invariants: cache nil")
	}

	if r.Resolver == nil {
		return "", merry.New("load balancer: check invariants: resolver nil")
	}

	nodes, err := r.Resolver.Resolve(service)
	if err != nil {
		return "", err.Prepend("load balancer: balance")
	}

	next := r.Cache.Next(service, nodes)

	if next == "" {
		return "", merry.New("load balancer: balance: no nodes available")
	}

	return next, nil
}
