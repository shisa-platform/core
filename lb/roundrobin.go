package lb

import (
	"sync"
	"sync/atomic"

	"github.com/percolate/shisa/sd"
)

var _ Cache = &rrCache{}

type ring struct {
	mtx   sync.RWMutex
	nodes []string
	index uint64
}

type rrCache struct {
	mtx      sync.RWMutex
	services map[string]*ring
}

// NewRoundRobin returns a Balancer that implements round robin load balancing
func NewRoundRobin(r sd.Resolver) Balancer {
	cache := newRRCache()
	return &cacheBalancer{
		Cache:    cache,
		Resolver: r,
	}
}

func newRRCache() *rrCache {
	return &rrCache{
		services: make(map[string]*ring),
	}
}

func (c *rrCache) Next(service string, nodes []string) string {
	if len(nodes) == 0 {
		return ""
	}

	var (
		inew uint64
		idx  uint64
	)

	c.mtx.RLock()
	r, ok := c.services[service]
	c.mtx.RUnlock()
	if !ok {
		// acquire write lock in order to add to add `*nodes` to the cache
		c.mtx.Lock()
		c.services[service] = &ring{
			nodes: nodes,
		}
		c.mtx.Unlock()

		r = c.services[service]

		r.mtx.RLock()
		defer r.mtx.RUnlock()
		idx = uint64(0)
		return r.nodes[idx]
	}

	// Update (potentially) happens here
	merged := make([]string, len(nodes))
	newSet := make(map[string]struct{}, len(nodes))

	for _, node := range nodes {
		newSet[node] = struct{}{}
	}

	r.mtx.RLock()
	i := 0
	for _, o := range r.nodes {
		if _, ok := newSet[o]; ok {
			merged[i] = o
			delete(newSet, o)
			i++
		}
	}
	r.mtx.RUnlock()

	for node := range newSet {
		merged[i] = node
		i++
	}

	r.mtx.Lock()
	r.nodes = merged
	r.mtx.Unlock()

	inew = atomic.AddUint64(&r.index, 1)

	r.mtx.RLock()
	defer r.mtx.RUnlock()

	idx = inew % uint64(len(r.nodes))
	return r.nodes[idx]
}
