package lb

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/sd"
)

var _ Balancer = &cacheBalancer{}

type cacheBalancer struct {
	Cache    Cache
	Resolver sd.Resolver
}

func (r *cacheBalancer) Balance(service string) (string, merry.Error) {
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
