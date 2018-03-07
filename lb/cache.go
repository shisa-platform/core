package lb

import (
	"sync"
)

type Cache interface {
	Next(service string, nodes []string) string
}

type NodeGroup struct {
	mtx sync.RWMutex

	nodes []*node
}

type node struct {
	host  string
	conns uint64
}

func UpdateNodes(newNodes, old []*node) []*node {
	merged := make([]*node, len(newNodes))
	newSet := make(map[string]*node, len(newNodes))

	for _, node := range newNodes {
		newSet[node.host] = node
	}

	i := 0
	for _, o := range old {
		if _, ok := newSet[o.host]; ok {
			merged[i] = o
			delete(newSet, o.host)
			i++
		}
	}

	for _, node := range newSet {
		merged[i] = node
		i++
	}
	return merged
}

func NewNodeGroup(nodeList []string) *NodeGroup {
	ng := &NodeGroup{
		nodes: make([]*node, len(nodeList)),
	}

	for i, host := range nodeList {
		ng.nodes[i] = &node{
			host: host,
		}
	}

	return ng
}
