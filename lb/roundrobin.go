package lb

import (
	"github.com/ansel1/merry"

	"github.com/percolate/shisa/sd"
)

var _ Balancer = &roundRobin{}

func NewRoundRobin(r sd.Resolver) Balancer {
	cache := NewRRCache()
	return &roundRobin{
		cache:    cache,
		resolver: r,
	}
}

func NewLNCNRC(r sd.Resolver) Balancer {
	cache := NewLNCNRCCache(2)
	return &roundRobin{
		cache:    cache,
		resolver: r,
	}
}

type roundRobin struct {
	cache    Cache
	resolver sd.Resolver
}

func (r *roundRobin) Balance(service string) (string, merry.Error) {
	nodes, err := r.resolver.Resolve(service)
	if err != nil {
		return "", err.Prepend("balance")
	}

	on := r.cache.Next(service, nodes)

	if on == "" {
		return "", merry.New("no nodes available")
	}

	return on, nil
}
