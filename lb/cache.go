package lb

import (
	"sync"
)

// Cache encapsulates the stateful parts of a load balancer
type Cache interface {
	// Next returns an addr given a service name and available addrs
	Next(service string, nodes []string) string
}

type nodeGroup struct {
	mtx   sync.RWMutex
	nodes []*node
}

type node struct {
	addr  string
	conns uint64
}

func updateNodes(newNodes, old []*node) []*node {
	merged := make([]*node, len(newNodes))
	newSet := make(map[string]*node, len(newNodes))

	for _, node := range newNodes {
		newSet[node.addr] = node
	}

	i := 0
	for _, o := range old {
		if _, ok := newSet[o.addr]; ok {
			merged[i] = o
			delete(newSet, o.addr)
			i++
		}
	}

	for _, node := range newSet {
		merged[i] = node
		i++
	}
	return merged
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
