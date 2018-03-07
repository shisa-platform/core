package lb

import (
	"sync"
	"sync/atomic"
)

var _ Cache = &rrCache{}

type ring struct {
	*NodeGroup

	c uint64
}

type rrCache struct {
	mtx sync.RWMutex

	r map[string]*ring
}

func NewRRCache() *rrCache {
	return &rrCache{
		r: make(map[string]*ring),
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
	ng := NewNodeGroup(nodes)

	c.mtx.RLock()
	r, ok := c.r[service]
	c.mtx.RUnlock()
	if !ok {
		// acquire write lock in order to add to add `*nodes` to the cache
		c.mtx.Lock()
		c.r[service] = &ring{NodeGroup: ng}
		c.mtx.Unlock()

		r = c.r[service]

		r.mtx.Lock()
		defer r.mtx.Unlock()

		idx = uint64(0)
	} else {
		r.mtx.Lock()
		defer r.mtx.Unlock()

		r.nodes = UpdateNodes(ng.nodes, r.nodes)

		inew = atomic.AddUint64(&r.c, 1)
		idx = inew % uint64(len(r.nodes))
	}
	return r.nodes[idx].host
}
