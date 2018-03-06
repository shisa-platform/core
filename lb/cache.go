package lb

import (
	"sync"
	"sync/atomic"
)

type Cache interface {
	Next(service string, nodes []string) string
}

type ring struct {
	mtx sync.RWMutex

	c     uint64
	nodes []string
}

var _ Cache = &BasicCache{}

type BasicCache struct {
	mtx sync.RWMutex
	r   map[string]*ring
}

func NewBasic() *BasicCache {

	return &BasicCache{
		r: make(map[string]*ring),
	}
}

func (c *BasicCache) Next(service string, nodes []string) string {
	if len(nodes) == 0 {
		return ""
	}

	c.mtx.RLock()
	r, ok := c.r[service]
	if !ok {
		c.mtx.RUnlock()

		// acquire write lock in order to add to add a `*ring` to the cache
		c.mtx.Lock()
		defer c.mtx.Unlock()

		c.r[service] = &ring{
			nodes: nodes,
		}

		return nodes[0]
	}
	c.mtx.RUnlock()

	r.mtx.Lock()
	defer r.mtx.Unlock()

	newNodes := make([]string, len(nodes))
	newNodeSet := make(map[string]struct{}, len(nodes))

	for _, host := range nodes {
		newNodeSet[host] = struct{}{}
	}

	i := 0
	for _, o := range r.nodes {
		if _, ok := newNodeSet[o]; ok {
			newNodes[i] = o
			delete(newNodeSet, o)
			i++
		}
	}

	for node := range newNodeSet {
		newNodes[i] = node
		i++
	}

	r.nodes = newNodes
	inew := atomic.AddUint64(&r.c, 1)
	idx := inew % uint64(len(r.nodes))
	return r.nodes[idx]
}
