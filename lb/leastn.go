package lb

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/percolate/shisa/sd"
)

var _ Cache = &leastConnsCache{}

type leastConnsCache struct {
	n int

	mtx sync.RWMutex

	r map[string]*nodeGroup
}

// NewLeastN returns a Balancer that returns the node with the least connections
// of `n` randomly chosen nodes
func NewLeastN(r sd.Resolver, n int) Balancer {
	cache := newLeastNCache(n)
	return &cacheBalancer{
		Cache:    cache,
		Resolver: r,
	}
}

func newLeastNCache(n int) *leastConnsCache {
	return &leastConnsCache{
		n: n,
		r: make(map[string]*nodeGroup),
	}
}

func (c *leastConnsCache) Next(service string, nodeList []string) string {
	if len(nodeList) == 0 {
		return ""
	}

	rando := rand.New(rand.NewSource(time.Now().Unix()))

	ng := newNodeGroup(nodeList)

	c.mtx.RLock()
	r, ok := c.r[service]
	c.mtx.RUnlock()

	if !ok {
		// acquire write lock in order to add to add `*nodes` to the cache
		c.mtx.Lock()
		c.r[service] = ng
		c.mtx.Unlock()

		r = ng
	} else {
		r.mtx.Lock()
		r.nodes = updateNodes(ng.nodes, r.nodes)
		r.mtx.Unlock()
	}

	g := c.n
	if len(nodeList) < c.n {
		g = len(nodeList)
	}

	r.mtx.RLock()
	var final *node
	for i := 0; i < g; i++ {
		choice := r.nodes[rando.Intn(len(nodeList))]
		if final == nil || choice.conns < final.conns {
			final = choice
		}
	}
	r.mtx.RUnlock()

	atomic.AddUint64(&final.conns, 1)
	return final.addr
}
