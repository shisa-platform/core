package lb

import (
	"math/rand"
)

var _ Cache = &randomCache{}

type randomCache struct{}

func NewRandomCache() *randomCache {
	return &randomCache{}
}

func (c *randomCache) Next(service string, nodes []string) string {
	if len(nodes) == 0 {
		return ""
	}

	return nodes[rand.Intn(len(nodes))]
}
