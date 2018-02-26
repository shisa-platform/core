package utility

import (
	"sync"
)

type Cache interface {
	Update(service string, nodes []string)
	Get(service string) []string
}

var _ Cache = &BasicCache{}

type BasicCache struct {
	mtx sync.RWMutex
	r   map[string][]string
}

func NewCache() *BasicCache {
	return &BasicCache{
		r: make(map[string][]string),
	}
}

func (c *BasicCache) Update(service string, nodes []string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	newNodes := make([]string, len(nodes))
	newNodeSet := make(map[string]struct{}, len(nodes))

	for _, host := range nodes {
		newNodeSet[host] = struct{}{}
	}

	i := 0
	for _, o := range c.r[service] {
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

	c.r[service] = newNodes
	return
}

func (c *BasicCache) Get(service string) []string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	return c.r[service]
}
