package lb

import (
	"sync"
	"sync/atomic"

	"github.com/shisa-platform/core/sd"
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
	return &CacheBalancer{
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

	c.mtx.RLock()
	r, ok := c.services[service]
	c.mtx.RUnlock()
	if !ok {
		c.mtx.Lock()
		// Check service again after acquiring write lock
		if r, ok = c.services[service]; !ok {
			c.services[service] = &ring{
				nodes: nodes,
			}

			c.mtx.Unlock()

			return nodes[0]
		}

		c.mtx.Unlock()
	}

	r.merge(nodes)

	return r.nextNode()
}

func (r *ring) merge(nodes []string) {
	merged := make([]string, 0, len(nodes))
	newSet := make(map[string]struct{}, len(nodes))

	for _, node := range nodes {
		newSet[node] = struct{}{}
	}

	r.mtx.RLock()
	for _, o := range r.nodes {
		if _, ok := newSet[o]; ok {
			merged = append(merged, o)
			delete(newSet, o)
		}
	}
	r.mtx.RUnlock()

	for node := range newSet {
		merged = append(merged, node)
	}

	r.mtx.Lock()
	r.nodes = merged
	r.mtx.Unlock()
}

func (r *ring) nextNode() string {
	inew := atomic.AddUint64(&r.index, 1)

	r.mtx.RLock()
	defer r.mtx.RUnlock()

	idx := inew % uint64(len(r.nodes))
	return r.nodes[idx]
}
