package lb

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shisa-platform/core/sd"
)

var _ Cache = &leastConnsCache{}

type leastConnsCache struct {
	n   int
	mtx sync.RWMutex
	r   map[string]*nodeGroup
}

// NewLeastN returns a Balancer that returns the node with the least connections
// of `n` randomly chosen nodes
func NewLeastN(r sd.Resolver, n int) Balancer {
	return &CacheBalancer{
		Cache: &leastConnsCache{
			n: n,
			r: make(map[string]*nodeGroup),
		},
		Resolver: r,
	}
}

func (c *leastConnsCache) Next(service string, nodeList []string) string {
	if len(nodeList) == 0 {
		return ""
	}

	ng := newNodeGroup(nodeList)

	c.mtx.RLock()
	r, ok := c.r[service]
	c.mtx.RUnlock()

	if !ok {
		// acquire write lock in order to add to add `*nodes` to the cache
		c.mtx.Lock()
		// Check service again after acquiring write lock
		if r, ok = c.r[service]; !ok {
			c.r[service] = ng
			c.mtx.Unlock()

			r = ng
		} else {
			c.mtx.Unlock()
			r.merge(ng.nodes)
		}
	} else {
		r.merge(ng.nodes)
	}

	g := c.n
	if len(nodeList) < c.n {
		g = len(nodeList)
	}

	final := r.choose(g)

	atomic.AddUint64(&final.conns, 1)
	return final.addr
}

func (c *leastConnsCache) connsForService(service string) uint64 {
	c.mtx.RLock()
	ng := c.r[service]
	c.mtx.RUnlock()
	return ng.totalConns()
}

type nodeGroup struct {
	mtx   sync.RWMutex
	nodes []*node
}

type node struct {
	addr  string
	conns uint64
}

func (n *nodeGroup) merge(newNodes []*node) {
	merged := make([]*node, 0, len(newNodes))
	newSet := make(map[string]*node, len(newNodes))

	n.mtx.Lock()
	defer n.mtx.Unlock()

	for i := range newNodes {
		newN := newNodes[i]
		newSet[newN.addr] = newN
	}

	for i, o := range n.nodes {
		if _, ok := newSet[o.addr]; ok {
			merged = append(merged, n.nodes[i])
			delete(newSet, o.addr)
		}
	}

	for _, node := range newSet {
		merged = append(merged, node)
	}

	n.nodes = merged

	return
}

func (n *nodeGroup) choose(g int) (final *node) {
	rando := rand.New(rand.NewSource(time.Now().Unix()))

	n.mtx.RLock()
	defer n.mtx.RUnlock()

	for i := 0; i < g; i++ {
		choice := n.nodes[rando.Intn(len(n.nodes))]
		if final == nil || choice.conns < final.conns {
			final = choice
		}
	}

	return final
}

func (n *nodeGroup) totalConns() (m uint64) {
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	for i := range n.nodes {
		m += n.nodes[i].conns
	}

	return
}

func newNodeGroup(nodeList []string) *nodeGroup {
	ng := &nodeGroup{
		nodes: make([]*node, len(nodeList)),
	}

	for i, addr := range nodeList {
		ng.nodes[i] = &node{
			addr: addr,
		}
	}

	return ng
}
