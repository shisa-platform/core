package lb

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/sd"
)

var _ Balancer = &BasicBalancer{}

type Balancer interface {
	Balance(string) (string, merry.Error)
}

type BasicBalancer struct {
	Cache    Cache
	Resolver sd.Resolver
}

func NewRoundRobin(r sd.Resolver) Balancer {
	cache := NewRRCache()
	return &BasicBalancer{
		Cache:    cache,
		Resolver: r,
	}
}

func NewRandom(r sd.Resolver) Balancer {
	cache := NewRandomCache()
	return &BasicBalancer{
		Cache:    cache,
		Resolver: r,
	}
}

func NewLeastConns(r sd.Resolver, n int) Balancer {
	cache := NewLeastConnsCache(n)
	return &BasicBalancer{
		Cache:    cache,
		Resolver: r,
	}
}

func (r *BasicBalancer) Balance(service string) (string, merry.Error) {
	nodes, err := r.Resolver.Resolve(service)
	if err != nil {
		return "", err.Prepend("balance")
	}

	on := r.Cache.Next(service, nodes)

	if on == "" {
		return "", merry.New("no nodes available")
	}

	return on, nil
}
