package lb

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var _ Cache = &leastConnsCache{}

type leastConnsCache struct {
	n int

	mtx sync.RWMutex

	r map[string]*NodeGroup
}

func NewLeastConnsCache(n int) *leastConnsCache {
	rand.Seed(time.Now().Unix())
	return &leastConnsCache{
		n: n,
		r: make(map[string]*NodeGroup),
	}
}

func (c *leastConnsCache) Next(service string, nodeList []string) string {
	if len(nodeList) == 0 {
		return ""
	}

	ng := NewNodeGroup(nodeList)

	c.mtx.RLock()
	r, ok := c.r[service]
	c.mtx.RUnlock()

	if !ok {
		// acquire write lock in order to add to add `*nodes` to the cache
		c.mtx.Lock()
		c.r[service] = ng
		c.mtx.Unlock()

		r = c.r[service]

		r.mtx.Lock()
		defer r.mtx.Unlock()
	} else {
		r.mtx.Lock()
		r.nodes = UpdateNodes(ng.nodes, r.nodes)
		defer r.mtx.Unlock()
	}

	var g int
	if len(nodeList) < c.n {
		g = len(nodeList)
	} else {
		g = c.n
	}

	var final *node
	for i := 0; i < g; i++ {
		ch := r.nodes[rand.Intn(len(nodeList))]
		if final == nil || ch.conns < final.conns {
			final = ch
		}
	}
	atomic.AddUint64(&final.conns, 1)
	return final.host
}
