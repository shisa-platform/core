package lb

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Cache interface {
	Next(service string, nodes []string) string
}

type ring struct {
	mtx sync.RWMutex

	c     uint64
	nodes []string
}

var (
	_ Cache = &baseCache{}
	_ Cache = &rrCache{}
	_ Cache = &lncnrcCache{}
)

type baseCache struct {
	mtx sync.RWMutex
	r   map[string]*ring
}

func (c *baseCache) Next(service string, nodes []string) string {
	return ""
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

type lncnrcCache struct {
	n uint64

	mtx sync.RWMutex

	r map[string]*nodes
}

type nodes struct {
	mtx sync.RWMutex

	ns []*node
}

type node struct {
	host  string
	conns uint64
}

func NewLNCNRCCache(n uint64) *lncnrcCache {
	rand.Seed(time.Now().Unix())
	return &lncnrcCache{
		n: n,
		r: make(map[string]*nodes),
	}
}

func (c *lncnrcCache) Next(service string, nodeList []string) string {
	if len(nodeList) == 0 {
		return ""
	}

	nodeobjs := &nodes{
		ns: make([]*node, len(nodeList)),
	}

	for i, host := range nodeList {
		nodeobjs.ns[i] = &node{
			host: host,
		}
	}

	c.mtx.RLock()
	r, ok := c.r[service]
	c.mtx.RUnlock()

	if !ok {
		// acquire write lock in order to add to add `*nodes` to the cache
		c.mtx.Lock()
		c.r[service] = nodeobjs
		r = c.r[service]
		c.mtx.Unlock()
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	newNodes := make([]*node, len(nodeList))
	newNodeSet := make(map[*node]struct{}, len(nodeList))

	for _, host := range nodeobjs.ns {
		newNodeSet[host] = struct{}{}
	}

	i := 0
	for _, o := range r.ns {
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

	var g uint64
	if uint64(len(nodeList)) < c.n {
		g = uint64(len(nodeList))
	} else {
		g = c.n
	}

	final := &node{
		conns: math.MaxUint64,
	}
	for i := uint64(0); i < g; i++ {
		ch := newNodes[rand.Intn(len(nodeList))]
		if ch.conns < final.conns {
			final = ch
		}
	}
	final.conns++
	return final.host
}
