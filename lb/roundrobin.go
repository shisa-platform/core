package lb

import (
	"sync/atomic"

	"github.com/ansel1/merry"

	"github.com/percolate/shisa/sd"
	"github.com/percolate/shisa/utility"
)

var _ Balancer = &roundRobin{}

func NewRoundRobin(r sd.Resolver) Balancer {
	cache := utility.NewCache()
	return &roundRobin{
		cache:    cache,
		resolver: r,
		c:        0,
	}
}

type roundRobin struct {
	cache    utility.Cache
	resolver sd.Resolver
	c        uint64
}

func (r *roundRobin) Balance(service string) (string, merry.Error) {
	nodes, err := r.resolver.Resolve(service)
	if err != nil {
		return "", err.Prepend("balance")
	}

	r.cache.Update(service, nodes)
	on := r.cache.Get(service)

	if len(on) <= 0 {
		return "", merry.New("no nodes available")
	}

	old := atomic.AddUint64(&r.c, 1) - 1
	idx := old % uint64(len(on))

	return on[idx], nil
}
